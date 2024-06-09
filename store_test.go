package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestInitStore(t *testing.T) {
	t.Run("create a new store", func(t *testing.T) {
		dir := t.TempDir()

		if _, err := initStore(dir); err != nil {
			t.Errorf("initStore(%q) = _, %q; want _, nil", dir, err)
		}
	})
	t.Run("fail when initialising with root of existing store", func(t *testing.T) {
		dir := t.TempDir()

		_, err := initStore(dir)
		noErr(t, err)

		wantErr := storeAlreadyExistsError(dir)

		if _, err := initStore(dir); !errors.Is(err, wantErr) {
			t.Errorf("initStore(%q) = _, %q; want _, %q", dir, err, wantErr)
		}
	})
	t.Run("fail when target path not a directory", func(t *testing.T) {
		rootPath := newTestFile(t, "")

		wantErr := notDirectoryError(rootPath)

		if _, err := initStore(rootPath); !errors.Is(err, wantErr) {
			t.Errorf("initStore(%q) = _, %q; want _, %q", rootPath, err, wantErr)
		}
	})
}

func TestEntries(t *testing.T) {
	t.Run("return 0 entries for a new store", func(t *testing.T) {
		s, _ := newTestStore(t)

		entries, err := s.entries()
		noErr(t, err)

		if len(entries) != 0 {
			t.Errorf("len(entries) = %d; want 0", len(entries))
		}
	})
}

func TestAdd(t *testing.T) {
	t.Run("add a new link", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)

		err := s.add(e)
		noErr(t, err)

		entries, err := s.entries()
		noErr(t, err)

		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d; want 1", len(entries))
		}
		if entries[0] != e {
			t.Errorf("entries[0] = %+v; want %+v", e, entries[0])
		}
	})
	t.Run("behave idempotently when adding links", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)

		err := s.add(e)
		noErr(t, err)

		s.add(e)
		noErr(t, err)

		entries, err := s.entries()
		noErr(t, err)

		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d; want 1", len(entries))
		}
		if entries[0] != e {
			t.Errorf("entries[0] = %+v; want %+v", e, entries[0])
		}
	})
	t.Run("add multiple entries", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)
		err := s.add(e)
		noErr(t, err)

		e2 := newTestEntry(rootPath)
		err = s.add(e2)
		noErr(t, err)

		entries, err := s.entries()
		noErr(t, err)

		if len(entries) != 2 {
			t.Fatalf("len(entries) = %d; want 2", len(entries))
		}
	})
	t.Run("fail when adding a link with source outside of store", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		l := link{
			SourcePath:      "/outside-store",
			DestinationPath: "",
		}

		wantErr := sourceOutsideRootError{
			rootPath:   rootPath,
			sourcePath: l.SourcePath,
		}

		if err := s.add(l); !errors.Is(err, wantErr) {
			t.Errorf("s.add(%+v) = %q; want %q", l, err, wantErr)
		}
	})
	t.Run("store homedir for path internally as '~'", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		homeDir, err := os.UserHomeDir()
		noErr(t, err)

		e := newTestEntry(rootPath)
		e.DestinationPath = fmt.Sprintf("%s/test-link", homeDir)

		err = s.add(e)
		noErr(t, err)

		assert(t, len(s.Entries) == 1)
		internalEntry := s.Entries[0]

		wantPath := "~/test-link"

		if internalEntry.DestinationPath != wantPath {
			t.Errorf("entry.destinationPath = %q; want %q", internalEntry.DestinationPath, wantPath)
		}
	})
	t.Run("entries with homedir are expanded when retrieved", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		homeDir, err := os.UserHomeDir()
		noErr(t, err)

		e := newTestEntry(rootPath)
		e.DestinationPath = fmt.Sprintf("%s/test-link", homeDir)

		err = s.add(e)
		noErr(t, err)

		entries, err := s.entries()
		noErr(t, err)

		assert(t, len(entries) == 1)
		externalEntry := entries[0]

		if externalEntry != e {
			t.Errorf("store.entries()[0] = %+v; want %+v", externalEntry, e)
		}
	})
}

func TestOpen(t *testing.T) {
	t.Run("restore the state of an existing store", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)
		err := s.add(e)
		noErr(t, err)

		restoredStore, err := openStore(rootPath)
		noErr(t, err)

		entries, err := restoredStore.entries()
		noErr(t, err)

		if len(entries) != 1 {
			t.Errorf("len(entries) = %d; want 1", len(entries))
		}
	})
	t.Run("restore an empty store", func(t *testing.T) {
		_, rootPath := newTestStore(t)

		restoredStore, err := openStore(rootPath)
		noErr(t, err)

		entries, err := restoredStore.entries()
		noErr(t, err)

		if len(entries) != 0 {
			t.Errorf("len(entries) = %d; want 0", len(entries))
		}
	})
	t.Run("restore entries with correct state", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)
		err := s.add(e)
		noErr(t, err)

		restoredStore, err := openStore(rootPath)
		noErr(t, err)

		entries, err := restoredStore.entries()
		noErr(t, err)

		assert(t, len(entries) == 1)
		restoredEntry := entries[0]

		if restoredEntry != e {
			t.Errorf("store.entries()[0] = %+v; want %+v", restoredEntry, e)
		}
	})
	t.Run("fail when opening a non-existent store", func(t *testing.T) {
		dir := t.TempDir()

		wantErr := storeNotExistError(dir)

		if _, err := openStore(dir); !errors.Is(err, wantErr) {
			t.Errorf("openStore(%q) = %q; want %q", dir, err, wantErr)
		}
	})
}

func newTestStore(t *testing.T) (store, string) {
	dir := t.TempDir()

	s, err := initStore(dir)
	noErr(t, err)

	return s, dir
}

func newTestEntry(dir string) link {
	return link{
		SourcePath:      fmt.Sprintf("%s/source-file-%s", dir, randomString(8)),
		DestinationPath: "",
	}
}
