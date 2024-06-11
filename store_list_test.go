package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLoadStoreList(t *testing.T) {
	t.Run("create a new store list when target directory empty", func(t *testing.T) {
		dir := t.TempDir()

		storeList, err := loadStoreList(dir)
		noErr(t, err)

		stores := storeList.stores()

		if len(stores) != 0 {
			t.Errorf("len(stores) = %d; want 0", len(stores))
		}
	})
	t.Run("restore existing entries", func(t *testing.T) {
		dir := t.TempDir()

		storeList, err := loadStoreList(dir)
		noErr(t, err)

		_, storePath := newTestStore(t)

		storeName := "test-store"

		err = storeList.addStore(storeName, storePath)
		noErr(t, err)

		storeList, err = loadStoreList(dir)
		noErr(t, err)

		stores := storeList.stores()

		if len(stores) != 1 {
			t.Fatalf("len(stores) = %d; want 1", len(stores))
		}
		if _, ok := stores[storeName]; !ok {
			t.Errorf("stores[%q] = _, false; want _, true", storeName)
		}
	})
	t.Run("fail when target directory is not a directory", func(t *testing.T) {
		dir := newTestFile(t, "")

		wantErr := notDirectoryError(dir)

		if _, err := loadStoreList(dir); err.Error() != wantErr.Error() {
			t.Errorf("loadStoreList(%q) = %q; want %q", dir, err, wantErr)
		}
	})
	t.Run("fail when target path does not exist", func(t *testing.T) {
		dir := "/does-not-exist"

		wantErr := pathNotExistError(dir)

		if _, err := loadStoreList(dir); err.Error() != wantErr.Error() {
			t.Errorf("loadStoreList(%q) = %q; want %q", dir, err, wantErr)
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
	t.Run("use store directory name as default name", func(t *testing.T) {
		storeList := newTestStoreList(t)

		wantName := "default-name"

		dir := t.TempDir()
		storePath := fmt.Sprintf("%s/%s", dir, wantName)

		err := os.Mkdir(storePath, 0755)
		noErr(t, err)

		_, err = initStore(storePath)
		noErr(t, err)

		err = storeList.addStore("", storePath)
		noErr(t, err)

		if _, ok := storeList.stores()[wantName]; !ok {
			t.Errorf("storeList.stores()[%q] = _, false; want true", wantName)
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
	sl, err := loadStoreList(dir)
	noErr(t, err)
	return sl
}
