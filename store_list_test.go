package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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

func TestAddStore(t *testing.T) {
	t.Run("add a store", func(t *testing.T) {
		storeList := newTestStoreList(t)
		s, storePath := newTestStore(t)

		storeName := "test-store"

		err := storeList.addStore(storeName, storePath)
		noErr(t, err)

		stores := storeList.stores()

		if len(stores) != 1 {
			t.Fatalf("len(stores) = %d; want 1", len(stores))
		}
		if retrievedStore, ok := stores[storeName]; !ok ||
			!cmp.Equal(*retrievedStore, s, cmpopts.IgnoreUnexported(store{})) {
			t.Errorf(
				"stores[%q] = %+v, %v; want %+v, true", storeName, retrievedStore, ok, s)
		}
	})
	t.Run("behave idempotently when adding stores", func(t *testing.T) {
		storeList := newTestStoreList(t)
		_, storePath := newTestStore(t)

		storeName := "test-store"

		err := storeList.addStore(storeName, storePath)
		noErr(t, err)
		err = storeList.addStore(storeName, storePath)
		noErr(t, err)

		stores := storeList.stores()

		if len(stores) != 1 {
			t.Fatalf("len(stores) = %d; want 1", len(stores))
		}
	})
	t.Run("fail when adding a new store with an existing name", func(t *testing.T) {
		storeList := newTestStoreList(t)

		_, storePath := newTestStore(t)
		_, conflictStorePath := newTestStore(t)

		storeName := "test-store"

		err := storeList.addStore(storeName, storePath)
		noErr(t, err)

		wantErr := storeNameExistError(storeName)

		if err := storeList.addStore(storeName, conflictStorePath); err == nil || err.Error() != wantErr.Error() {
			t.Errorf("storeList.add(%q, %q) = %q; want %q", storeName, conflictStorePath, err, wantErr)
		}
	})
}

func newTestStoreList(t *testing.T) storeList {
	dir := t.TempDir()
	return loadStoreList(dir)
}
