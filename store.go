package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const storeFileName = ".sto"

type store struct {
	rootPath     string
	storeEntries []link
}

func initStore(rootPath string) (store, error) {
	storeFilePath := fmt.Sprintf("%s/%s", rootPath, storeFileName)

	_, err := os.Stat(storeFilePath)
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

func (s *store) entries() []link {
	return s.storeEntries
}

func (s *store) add(l link) error {
	if !strings.HasPrefix(l.sourcePath, s.rootPath) {
		return sourceOutsideRootError{
			rootPath:   s.rootPath,
			sourcePath: l.sourcePath,
		}
	}
	for _, entry := range s.storeEntries {
		if entry.sourcePath == l.sourcePath &&
			entry.destinationPath == l.destinationPath {
			return nil
		}
	}
	s.storeEntries = append(s.storeEntries, l)
	return nil
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
