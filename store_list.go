package main

import (
	"fmt"
	"path/filepath"
)

type storeList struct {
	root     string
	storeMap *map[string]*store
}

func loadStoreList(path string) storeList {
	storeMap := make(map[string]*store)
	return storeList{
		root:     path,
		storeMap: &storeMap,
	}
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

	return nil
}

type storeNameExistError string

func (e storeNameExistError) Error() string {
	return fmt.Sprintf("store %q already exists", string(e))
}
