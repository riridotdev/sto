package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type storeList struct {
	root     string
	storeMap *map[string]*store
}

const storeListFileName = "storelist"

func loadStoreList(path string) (storeList, error) {
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return storeList{}, pathNotExistError(path)
	}
	if err != nil {
		return storeList{}, fmt.Errorf("reading stat %q: %v", path, err)
	}
	if !stat.IsDir() {
		return storeList{}, notDirectoryError(path)
	}

	storeListFilePath := fmt.Sprintf("%s/%s", path, storeListFileName)

	storeMap := make(map[string]*store)

	sl := storeList{
		root:     path,
		storeMap: &storeMap,
	}

	storeListFile, err := os.Open(storeListFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return sl, nil
	}
	if err != nil {
		return storeList{}, fmt.Errorf("reading storeList file %q: %v", storeListFilePath, err)
	}

	dec := json.NewDecoder(storeListFile)
	if err := dec.Decode(&storeMap); err != nil {
		return storeList{}, fmt.Errorf("reading storeList file %q: %v", storeListFilePath, err)
	}

	return sl, nil
}

func (sl storeList) stores() map[string]*store {
	return *sl.storeMap
}

func (sl storeList) addStore(name string, storePath string) error {
	if name == "" {
		_, name = filepath.Split(storePath)
	}

	store, err := openStore(storePath)
	if err != nil {
		return fmt.Errorf("opening store %q: %v", storePath, err)
	}

	if retrievedStore, ok := (*sl.storeMap)[name]; ok {
		if retrievedStore.rootPath == store.rootPath {
			return nil
		}
		return storeNameExistError(name)
	}

	(*sl.storeMap)[name] = &store

	if err := sl.persist(); err != nil {
		return fmt.Errorf("persisting store list: %v", err)
	}

	return nil
}

func (sl storeList) persist() error {
	storeListFilePath := fmt.Sprintf("%s/%s", sl.root, storeListFileName)

	f, err := os.OpenFile(storeListFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening file %q: %v", storeListFilePath, err)
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(sl.storeMap); err != nil {
		return fmt.Errorf("writing to file %q: %v", storeListFilePath, err)
	}

	return nil
}

type storeNameExistError string

func (e storeNameExistError) Error() string {
	return fmt.Sprintf("store %q already exists", string(e))
}
