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

type storeListEntry struct {
	Name string
	Root string
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

	var storeListEntries []storeListEntry

	storeListFile, err := os.Open(storeListFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return sl, nil
	}
	if err != nil {
		return storeList{}, fmt.Errorf("reading storeList file %q: %v", storeListFilePath, err)
	}

	dec := json.NewDecoder(storeListFile)
	if err := dec.Decode(&storeListEntries); err != nil {
		return storeList{}, fmt.Errorf("reading storeList file %q: %v", storeListFilePath, err)
	}

	for _, entry := range storeListEntries {
		store, err := openStore(entry.Root)
		if err != nil {
			return storeList{}, fmt.Errorf("opening store %q: %v", entry.Name, err)
		}
		storeMap[entry.Name] = &store
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

	var storeListEntries []storeListEntry

	for name, store := range *sl.storeMap {
		storeListEntries = append(storeListEntries, storeListEntry{
			Name: name,
			Root: store.rootPath,
		})
	}

	enc := json.NewEncoder(f)
	if err := enc.Encode(&storeListEntries); err != nil {
		return fmt.Errorf("writing to file %q: %v", storeListFilePath, err)
	}

	return nil
}

type storeNameExistError string

func (e storeNameExistError) Error() string {
	return fmt.Sprintf("store %q already exists", string(e))
}
