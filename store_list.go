package main

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
