package sto

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
	Store map[string]storeEntry
	Root  string
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
		Store: map[string]storeEntry{},
		Root:  root,
	}

	for _, entry := range entries {
		s.Store[entry.Name] = entry
	}

	return s, nil
}

// Write persists any changes to the store back to disk.
//
// The changes will be written to the .sto file at the store's root.
func (s *store) Write() error {
	if !s.dirty {
		return nil
	}

	storeFilePath := fmt.Sprintf("%s/.sto", s.Root)

	file, err := os.OpenFile(storeFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error opening store file at %q: %w", storeFilePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(s.Entries()); err != nil {
		return fmt.Errorf("error writing to store file: %w", err)
	}

	s.dirty = false

	return nil
}

// Entries returns all of the Entries managed by a store, sorted alphabetically by name.
func (s *store) Entries() []storeEntry {
	entries := []storeEntry{}
	for _, entry := range s.Store {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	return entries
}

// Add creates a new entry in the store for a given source directory and destination directory,
// then creates a symlink from the source to the destination.
//
// The entry source should be provided as a relative path in relation to the store's root,
// the entry destination should be provided as an absolute path.
func (s *store) Add(source, destination string) error {
	source = fmt.Sprintf("%s/%s", s.Root, source)

	if !strings.HasPrefix(source, s.Root) {
		return fmt.Errorf("cannot add sources from outside the sto root (%q)\n", s.Root)
	}

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("couldn't locate valid sto item at %q\n", source)
	}

	name := trimStoRoot(source, s.Root)

	if linkedLocation, ok := s.Store[name]; ok {
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

	s.Store[name] = entry

	s.dirty = true

	return nil
}

// RemoveEntry removes an entry from the store.
//
// If the entry is currently linked, the link will be removed.
func (s *store) RemoveEntry(name string) error {
	entry, ok := s.Store[name]
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

	delete(s.Store, name)

	s.dirty = true

	return nil
}

// CheckEntry evaluates the current state of an item.
//
// This includes a series of checks including the presence of
// the source item, the presence of a symlink at the destination,
// and parity between the source/destination of both the entry definition
// and any pre-existing symlinks.
// Will return false if the entry is unlinked, and will return an error if
// there are inconsistencies between the system state and definition state.
func (s store) CheckEntry(name string) (bool, error) {
	entry, ok := s.Store[name]
	if !ok {
		return false, fmt.Errorf("Entry %q not found", name)
	}

	sourcePath := fmt.Sprintf("%s/%s", s.Root, entry.Source)
	if _, err := os.Stat(sourcePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("Error checking file stat at %q: %s", sourcePath, err)
		}
		return false, ErrEntrySourceInvalid
	}

	stat, err := os.Lstat(entry.Destination)
	if err != nil {
		return false, fmt.Errorf("Error checking link stat at %q: %s", entry.Destination, err)
	}
	if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		return false, ErrExistingFileAtDestination
	}

	link, err := os.Readlink(entry.Destination)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("Error reading symlink at %q: %w", entry.Destination, err)
		}
		return false, nil
	}

	if link != sourcePath {
		return false, ErrExistingSymlinkMismatch(link)
	}

	return true, nil
}

// ApplyEntry creates a symlink as defined by the entry definition managed under the given name.
func (s store) ApplyEntry(name string) error {
	entry, ok := s.Store[name]
	if !ok {
		return fmt.Errorf("Item %q not found", name)
	}
	return s._applyEntry(entry)
}

// _applyEntry creates a symlink for a given entry.
// This is a helper method that provides shared functionality and should not be called on its own.
// store.applyEntry() or store.add() should be used instead.
func (s store) _applyEntry(entry storeEntry) error {
	sourcePath := fmt.Sprintf("%s/%s", s.Root, entry.Source)

	stat, err := os.Lstat(entry.Destination)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Error reading link stat at %q: %w", entry.Destination, err)
	}
	if err == nil && stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		fmt.Printf("Found existing file at %q\n", entry.Destination)
		fmt.Printf("Do you want to delete this file? [y/n]\n")

		input := bufio.NewReader(os.Stdin)
		line, _, err := input.ReadLine()
		if err != nil {
			Fail("Error reading input: %s\n", err)
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
		return ErrLinkAlreadyExists
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Error reading link at %q: %w", entry.Destination, err)
	}

	if err := os.Symlink(sourcePath, entry.Destination); err != nil {
		return fmt.Errorf("Error creating symlink %q -> %q: %s\n", sourcePath, entry.Destination, err)
	}

	return nil
}

// UnapplyEntry removes a symlink as defined by the entry managed under the given name.
func (s *store) UnapplyEntry(name string) error {
	entry, ok := s.Store[name]
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

// RenameEntry updates the name under which a symlink is managed.
//
// Calling this method has no effect on the name of the entry's source directory.
func (s *store) RenameEntry(entryName, newName string) error {
	entry, ok := s.Store[entryName]
	if !ok {
		return errEntryNotFound
	}

	if _, ok := s.Store[newName]; ok {
		return errEntryAlreadyExists
	}

	entry.Name = newName

	s.Store[newName] = entry
	delete(s.Store, entryName)

	s.dirty = true

	return nil
}

// MoveEntry moves the file managed under the given name to a new destination at the given path.
//
// If the entry is currently linked, the old symlink will be removed and recreated with the updated path.
func (s *store) MoveEntry(entryName, newPath string) error {
	entry, ok := s.Store[entryName]
	if !ok {
		return errEntryNotFound
	}

	newPath, err := filepath.Abs(newPath)
	if !strings.HasPrefix(newPath, s.Root) {
		return fmt.Errorf("cannot move a source outside the sto root (%q)\n", s.Root)
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

	sourcePath := fmt.Sprintf("%s/%s", s.Root, entry.Source)
	if err := os.Rename(sourcePath, newPath); err != nil {
		return fmt.Errorf("Error moving entry from %q to %q: %s", sourcePath, newPath, err)
	}

	entry.Source = trimStoRoot(newPath, s.Root)

	if linked {
		if err := s._applyEntry(entry); err != nil {
			return fmt.Errorf("Error applying entry %q: %s", entry.Name, err)
		}
	}

	s.Store[entry.Name] = entry

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
	return link == fmt.Sprintf("%s/%s", s.Root, entry.Source), nil
}
