package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A store maintains a list of storeEntries as defined by the .sto file in the root directory.
//
// It also provides an interface for all Sto operations to be performed as they relate to a single profile.
//
// To persist changes to disk call the write method.
type store struct {
	store map[string]storeEntry
	root  string
	dirty bool // True if there are changes not persisted to disk.
}

// A storeEntry represents a mapping between a source path relative to the Sto root,
// and a destination path at which a symlink should be constructed.
type storeEntry struct {
	Name        string // The name that Sto will use to manage the entry.
	Source      string // Must be a relative path.
	Destination string // Must be an absolute path.
}

// readStore constructs and returns a store from the .sto file at the given root directory.
func readStore(root string) (store, error) {
	storeFilePath := fmt.Sprintf("%s/.sto", root)

	storeFile, err := os.Open(storeFilePath)
	if err != nil {
		return store{}, fmt.Errorf("error opening store file %q: %s", storeFilePath, err)
	}
	defer storeFile.Close()

	decoder := json.NewDecoder(storeFile)

	entries := []storeEntry{}
	decoder.Decode(&entries)

	s := store{
		store: map[string]storeEntry{},
		root:  root,
	}

	for _, entry := range entries {
		s.store[entry.Name] = entry
	}

	return s, nil
}

// write persists any changes to the store back to disk.
//
// The changes will be written to the .sto file at the store's root.
func (s *store) write() error {
	if !s.dirty {
		return nil
	}

	storeFilePath := fmt.Sprintf("%s/.sto", s.root)

	file, err := os.OpenFile(storeFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error opening store file at %q: %w", storeFilePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(s.entries()); err != nil {
		return fmt.Errorf("error writing to store file: %w", err)
	}

	s.dirty = false

	return nil
}

// entries returns all of the entries managed by a store, sorted alphabetically by name.
func (s *store) entries() []storeEntry {
	entries := []storeEntry{}
	for _, entry := range s.store {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// add creates a new entry in the store for a given source directory and destination directory,
// then creates a symlink from the source to the destination.
//
// The entry source should be provided as a relative path in relation to the store's root,
// the entry destination should be provided as an absolute path.
func (s *store) add(source, destination string) error {
	source = fmt.Sprintf("%s/%s", s.root, source)

	if !strings.HasPrefix(source, s.root) {
		return fmt.Errorf("cannot add sources from outside the sto root (%q)\n", defaultRoot)
	}

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("couldn't locate valid sto item at %q\n", source)
	}

	name := trimStoRoot(source, s.root)

	if linkedLocation, ok := s.store[name]; ok {
		return fmt.Errorf("item %q already linked at %q\n", name, linkedLocation)
	}

	entry := storeEntry{
		Name:        name,
		Source:      name,
		Destination: destination,
	}

	if err := s._applyEntry(entry); err != nil {
		return fmt.Errorf("Error linking entry %q: %w", entry.Name, err)
	}

	s.store[name] = entry

	s.dirty = true

	return nil
}

// removeEntry removes an entry from the store.
//
// If the entry is currently linked, the link will be removed.
func (s *store) removeEntry(name string) error {
	entry, ok := s.store[name]
	if !ok {
		return fmt.Errorf("Entry %q not found", name)
	}

	linked, err := s.isLinked(entry)
	if err != nil {
		return fmt.Errorf("Error checking entry %q: %w", entry.Name, err)
	}
	if linked {
		if err := s._unapplyEntry(entry); err != nil {
			return err
		}
	}

	delete(s.store, name)

	s.dirty = true

	return nil
}

// checkEntry evaluates the current state of an item.
//
// This includes a series of checks including the presence of
// the source item, the presence of a symlink at the destination,
// and parity between the source/destination of both the entry definition
// and any pre-existing symlinks.
// Will return false if the entry is unlinked, and will return an error if
// there are inconsistencies between the system state and definition state.
func (s store) checkEntry(name string) (bool, error) {
	entry, ok := s.store[name]
	if !ok {
		return false, fmt.Errorf("Entry %q not found", name)
	}

	sourcePath := fmt.Sprintf("%s/%s", s.root, entry.Source)
	if _, err := os.Stat(sourcePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("Error checking file stat at %q: %s", sourcePath, err)
		}
		return false, errEntrySourceInvalid
	}

	stat, err := os.Lstat(entry.Destination)
	if err != nil {
		return false, fmt.Errorf("Error checking link stat at %q: %s", entry.Destination, err)
	}
	if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		return false, errExistingFileAtDestination
	}

	link, err := os.Readlink(entry.Destination)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("Error reading symlink at %q: %w", entry.Destination, err)
		}
		return false, nil
	}

	if link != sourcePath {
		return false, errExistingSymlinkMismatch(link)
	}

	return true, nil
}

// applyEntry creates a symlink as defined by the entry definition managed under the given name.
func (s store) applyEntry(name string) error {
	entry, ok := s.store[name]
	if !ok {
		return fmt.Errorf("Item %q not found", name)
	}
	return s._applyEntry(entry)
}

// _applyEntry creates a symlink for a given entry.
// This is a helper method that provides shared functionality and should not be called on its own.
// store.applyEntry() or store.add() should be used instead.
func (s store) _applyEntry(entry storeEntry) error {
	sourcePath := fmt.Sprintf("%s/%s", s.root, entry.Source)

	stat, err := os.Lstat(entry.Destination)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Error reading link stat at %q: %w", entry.Destination, err)
	}
	if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		fmt.Printf("Found existing file at %q\n", entry.Destination)
		fmt.Printf("Do you want to delete this file? [y/n]\n")

		input := bufio.NewReader(os.Stdin)
		line, _, err := input.ReadLine()
		if err != nil {
			fail("Error reading input: %s\n", err)
		}
		if !(line[0] == 'y' || line[0] == 'Y') {
			return nil
		}

		if err := os.RemoveAll(entry.Destination); err != nil {
			return fmt.Errorf("Error deleting file at %q: %w", entry.Destination, err)
		}
	}

	link, err := os.Readlink(entry.Destination)
	if err == nil {
		if link != sourcePath {
			return fmt.Errorf("Prexisting link %q -> %q", link, entry.Destination)
		}
		return errLinkAlreadyExists
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Error reading link at %q: %w", entry.Destination, err)
	}

	if err := os.Symlink(sourcePath, entry.Destination); err != nil {
		return fmt.Errorf("Error creating symlink %q -> %q: %s\n", sourcePath, entry.Destination, err)
	}

	return nil
}

// unapplyEntry removes a symlink as defined by the entry managed under the given name.
func (s *store) unapplyEntry(name string) error {
	entry, ok := s.store[name]
	if !ok {
		return fmt.Errorf("Entry %q not found", name)
	}
	return s._unapplyEntry(entry)
}

func (s *store) _unapplyEntry(entry storeEntry) error {
	if err := os.Remove(entry.Destination); err != nil {
		return fmt.Errorf("Error removing symlink at %q: %w", entry.Destination, err)
	}
	return nil
}

// renameEntry updates the name under which a symlink is managed.
//
// Calling this method has no effect on the name of the entry's source directory.
func (s *store) renameEntry(entryName, newName string) error {
	entry, ok := s.store[entryName]
	if !ok {
		return errEntryNotFound
	}

	if _, ok := s.store[newName]; ok {
		return errEntryAlreadyExists
	}

	entry.Name = newName

	s.store[newName] = entry
	delete(s.store, entryName)

	s.dirty = true

	return nil
}

// moveEntry moves the file managed under the given name to a new destination at the given path.
//
// If the entry is currently linked, the old symlink will be removed and recreated with the updated path.
func (s *store) moveEntry(entryName, newPath string) error {
	entry, ok := s.store[entryName]
	if !ok {
		return errEntryNotFound
	}

	newPath, err := filepath.Abs(newPath)
	if !strings.HasPrefix(newPath, s.root) {
		return fmt.Errorf("cannot move a source outside the sto root (%q)\n", defaultRoot)
	}

	_, err = os.Stat(newPath)
	if err == nil {
		return errSourceAlreadyExists
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Error checking file %q: %s", newPath, err)
	}

	linked, err := s.isLinked(entry)
	if err != nil {
		return fmt.Errorf("Error checking entry status for %q: %s", entryName, err)
	}
	if linked {
		if err := s._unapplyEntry(entry); err != nil {
			return fmt.Errorf("Error unlinking entry %s: %s", entryName, err)
		}
	}

	sourcePath := fmt.Sprintf("%s/%s", s.root, entry.Source)
	if err := os.Rename(sourcePath, newPath); err != nil {
		return fmt.Errorf("Error moving entry from %q to %q: %s", sourcePath, newPath, err)
	}

	entry.Source = trimStoRoot(newPath, s.root)

	if linked {
		if err := s._applyEntry(entry); err != nil {
			return fmt.Errorf("Error applying entry %q: %s", entry.Name, err)
		}
	}

	s.store[entry.Name] = entry

	s.dirty = true

	return nil
}

// isLinked checks if an entry is actively linked
func (s store) isLinked(entry storeEntry) (bool, error) {
	link, err := os.Readlink(entry.Destination)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("Error reading symlink at %q: %w", entry.Destination, err)
		}
		return false, nil
	}
	return link == fmt.Sprintf("%s/%s", s.root, entry.Source), nil
}
