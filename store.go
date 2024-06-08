package main

import (
	"errors"
	"fmt"
	"os"
)

const storeFileName = ".sto"

type store struct{}

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

	return store{}, nil
}

func (s store) entries() []link {
	return nil
}

type storeAlreadyExistsError string

func (e storeAlreadyExistsError) Error() string {
	return fmt.Sprintf("store at %q already exists", string(e))
}
