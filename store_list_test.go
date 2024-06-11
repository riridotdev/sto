package main

import "testing"

func TestLoadStoreList(t *testing.T) {
	t.Run("create a new store list when target directory empty", func(t *testing.T) {
		dir := t.TempDir()

		storeList := loadStoreList(dir)

		stores := storeList.stores()

		if len(stores) != 0 {
			t.Errorf("len(stores) = %d; want 0", len(stores))
		}
	})
}
