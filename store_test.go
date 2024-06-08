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
}

func TestEntries(t *testing.T) {
	t.Run("return 0 entries for a new store", func(t *testing.T) {
		s, _ := newTestStore(t)

		entries := s.entries()

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

		entries := s.entries()

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

		entries := s.entries()

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

		entries := s.entries()

		if len(entries) != 2 {
			t.Fatalf("len(entries) = %d; want 2", len(entries))
		}
	})
	t.Run("fail when adding a link with source outside of store", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		l := link{
			sourcePath:      "/outside-store",
			destinationPath: "",
		}

		wantErr := sourceOutsideRootError{
			rootPath:   rootPath,
			sourcePath: l.sourcePath,
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
		e.destinationPath = fmt.Sprintf("%s/test-link", homeDir)

		err = s.add(e)
		noErr(t, err)

		assert(t, len(s.Entries) == 1)
		internalEntry := s.Entries[0]

		wantPath := "~/test-link"

		if internalEntry.destinationPath != wantPath {
			t.Errorf("entry.sourcePath = %q; want %q", internalEntry.destinationPath, wantPath)
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

		entries := restoredStore.entries()

		if len(entries) != 1 {
			t.Errorf("len(entries) = %d; want 1", len(entries))
		}
	})
	t.Run("restore an empty store", func(t *testing.T) {
		_, rootPath := newTestStore(t)

		restoredStore, err := openStore(rootPath)
		noErr(t, err)

		entries := restoredStore.entries()

		if len(entries) != 0 {
			t.Errorf("len(entries) = %d; want 0", len(entries))
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
		sourcePath:      fmt.Sprintf("%s/source-file-%s", dir, randomString(8)),
		destinationPath: "",
	}
}
