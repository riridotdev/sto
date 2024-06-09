package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

const storeFileName = ".sto"

type store struct {
	rootPath string
	Entries  []link
}

func initStore(rootPath string) (store, error) {
	stat, err := os.Stat(rootPath)
	if err != nil {
		return store{}, fmt.Errorf("reading stat %q: %v", rootPath, err)
	}
	if !stat.IsDir() {
		return store{}, notDirectoryError(rootPath)
	}

	storeFilePath := fmt.Sprintf("%s/%s", rootPath, storeFileName)

	_, err = os.Stat(storeFilePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return store{}, fmt.Errorf("reading stat %q: %v", storeFilePath, err)
	}
	if err == nil {
		return store{}, storeAlreadyExistsError(rootPath)
	}

	f, err := os.Create(storeFilePath)
	if err != nil {
		return store{}, fmt.Errorf("creating store file %q: %v", storeFilePath, err)
	}
	if err := f.Close(); err != nil {
		return store{}, fmt.Errorf("closing store file %q: %v", storeFilePath, err)
	}

	return store{rootPath: rootPath}, nil
}

func openStore(rootPath string) (s store, err error) {
	s.rootPath = rootPath

	storeFilePath := fmt.Sprintf("%s/%s", rootPath, storeFileName)

	stat, err := os.Stat(storeFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return store{}, storeNotExistError(rootPath)
	}
	if err != nil {
		return store{}, fmt.Errorf("reading stat %q: %v", storeFilePath, err)
	}
	if stat.Size() == 0 {
		return s, err
	}

	f, err := os.Open(storeFilePath)
	if err != nil {
		return store{}, fmt.Errorf("opening store file %q: %v", storeFilePath, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = fmt.Errorf("closing store file %q: %v", storeFilePath, closeErr)
		}
	}()

	dec := json.NewDecoder(f)
	if err := dec.Decode(&s); err != nil {
		return store{}, fmt.Errorf("reading store file %q: %v", storeFilePath, err)
	}

	return s, err
}

func (s *store) entries() ([]link, error) {
	var (
		entries []link
		err     error
	)
	for _, entry := range s.Entries {
		entry.DestinationPath, err = expand(entry.DestinationPath)
		if err != nil {
			return nil, fmt.Errorf("expanding homedir %q: %v", entry.DestinationPath, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (s *store) add(l link) error {
	if !strings.HasPrefix(l.SourcePath, s.rootPath) {
		return sourceOutsideRootError{
			rootPath:   s.rootPath,
			sourcePath: l.SourcePath,
		}
	}

	var err error
	l.DestinationPath, err = compress(l.DestinationPath)
	if err != nil {
		return fmt.Errorf("compressing path %q: %v", l.DestinationPath, err)
	}

	for _, entry := range s.Entries {
		if entry.SourcePath == l.SourcePath &&
			entry.DestinationPath == l.DestinationPath {
			return nil
		}
	}

	s.Entries = append(s.Entries, l)

	if err := s.persist(); err != nil {
		return fmt.Errorf("persisting store: %v", err)
	}

	return nil
}

func (s *store) persist() (err error) {
	storeFilePath := fmt.Sprintf("%s/%s", s.rootPath, storeFileName)

	f, err := os.OpenFile(storeFilePath, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening store file %q: %v", storeFilePath, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = fmt.Errorf("closing store file %q: %v", storeFilePath, closeErr)
		}
	}()

	enc := json.NewEncoder(f)
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("writing to store file %q: %v", storeFilePath, err)
	}

	return err
}

type notDirectoryError string

func (e notDirectoryError) Error() string {
	return fmt.Sprintf("%q is not a directory", string(e))
}

type storeAlreadyExistsError string

func (e storeAlreadyExistsError) Error() string {
	return fmt.Sprintf("store at %q already exists", string(e))
}

type sourceOutsideRootError struct {
	rootPath   string
	sourcePath string
}

func (e sourceOutsideRootError) Error() string {
	return fmt.Sprintf("source %q is outside of root %q", e.sourcePath, e.rootPath)
}

type storeNotExistError string

func (e storeNotExistError) Error() string {
	return fmt.Sprintf("store at %q does not exist", string(e))
}
