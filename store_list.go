package main

import (
	"fmt"
	"path/filepath"
)

type storeList struct {
	storeListFile jsonFile
	storeMap      map[string]*store
}

type storeListEntry struct {
	Name string
	Root string
}

const storeListFileName = "storelist"

func loadStoreList(path string) (storeList, error) {
	storeListFilePath := fmt.Sprintf("%s/%s", path, storeListFileName)
	storeListFile, err := newJsonFile(storeListFilePath)
	if err != nil {
		return storeList{}, err
	}

	storeMap := make(map[string]*store)
	sl := storeList{
		storeListFile: storeListFile,
		storeMap:      storeMap,
	}

	var storeListEntries []storeListEntry
	if err := sl.storeListFile.read(&storeListEntries); err != nil {
		return storeList{}, fmt.Errorf("reading store list file: %v", err)
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
	return sl.storeMap
}

func (sl storeList) addStore(name string, storePath string) error {
	if name == "" {
		_, name = filepath.Split(storePath)
	}

	store, err := openStore(storePath)
	if err != nil {
		return fmt.Errorf("opening store %q: %v", storePath, err)
	}

	if retrievedStore, ok := (sl.storeMap)[name]; ok {
		if retrievedStore.rootPath == store.rootPath {
			return nil
		}
		return storeNameExistError(name)
	}

	(sl.storeMap)[name] = &store

	if err := sl.persist(); err != nil {
		return fmt.Errorf("persisting store list: %v", err)
	}

	return nil
}

func (sl storeList) persist() error {
	var storeListEntries []storeListEntry

	for name, store := range sl.storeMap {
		storeListEntries = append(storeListEntries, storeListEntry{
			Name: name,
			Root: store.rootPath,
		})
	}

	if err := sl.storeListFile.write(&storeListEntries); err != nil {
		return fmt.Errorf("writing store file: %v", err)
	}

	return nil
}

type storeNameExistError string

func (e storeNameExistError) Error() string {
	return fmt.Sprintf("store %q already exists", string(e))
}
