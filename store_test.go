package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
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
			t.Errorf("entries[0] = %+v; want %+v", entries[0], e)
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
	t.Run("use relative path as default name", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)
		e.Name = ""

		err := s.add(e)
		noErr(t, err)

		fileName := strings.TrimPrefix(e.SourcePath, fmt.Sprintf("%s/", rootPath))
		e.Name = fileName

		retrievedEntry, _, err := s.get(fileName)
		noErr(t, err)

		if retrievedEntry != e {
			t.Errorf("store.get(%q) = %+v; want %+v", fileName, retrievedEntry, e)
		}
	})
	t.Run("fail when adding a link with source outside of store", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		l := link{
			SourcePath:      "/outside-store",
			DestinationPath: "/destination",
		}

		wantErr := sourceOutsideRootError{
			rootPath:   rootPath,
			sourcePath: l.SourcePath,
		}

		if err := s.add(l); err.Error() != wantErr.Error() {
			t.Errorf("s.add(%+v) = %q; want %q", l, err, wantErr)
		}
	})
	t.Run("fail when adding link with duplicate name", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		entryName := "entry-name"

		e := newTestEntry(rootPath)
		e.Name = entryName

		err := s.add(e)
		noErr(t, err)

		conflictEntry := newTestEntry(rootPath)
		conflictEntry.Name = entryName

		wantErr := entryExistError(entryName)

		if err := s.add(conflictEntry); err == nil || err.Error() != wantErr.Error() {
			t.Errorf("store.add(%+v) = %v; want %v", e, err, wantErr)
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
		internalEntry, _ := s.Entries[e.Name]

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
	t.Run("store paths internally as relative paths", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)
		e.SourcePath = fmt.Sprintf("%s/source-file", rootPath)

		err := s.add(e)
		noErr(t, err)

		assert(t, len(s.Entries) == 1)
		internalEntry, _ := s.Entries[e.Name]

		wantPath := "source-file"

		if internalEntry.SourcePath != wantPath {
			t.Errorf("entry.sourcePath = %q; want %q", internalEntry.SourcePath, wantPath)
		}
	})
	t.Run("fail when adding entry with empty fields", func(t *testing.T) {
		var e link

		tests := []struct {
			title   string
			field   *string
			wantErr error
		}{
			{title: "SourcePath", field: &e.SourcePath},
			{title: "DestinationPath", field: &e.DestinationPath},
		}

		for _, tt := range tests {
			t.Run(tt.title, func(t *testing.T) {
				s, rootPath := newTestStore(t)

				e = newTestEntry(rootPath)
				*tt.field = ""

				if err := s.add(e); err == nil {
					t.Errorf("store.add(%+v) = nil; want err", e)
				}
			})
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

func TestGet(t *testing.T) {
	t.Run("retrieve an entry from the store", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)
		e.Name = "test-entry"

		err := s.add(e)
		noErr(t, err)

		retrievedEntry, _, err := s.get(e.Name)
		noErr(t, err)

		if retrievedEntry != e {
			t.Errorf("store.get(%q) = %+v; want %+v", e.Name, retrievedEntry, e)
		}
	})
	t.Run("return not ok for a missing entry", func(t *testing.T) {
		s, _ := newTestStore(t)

		if _, ok, _ := s.get("missing"); ok {
			t.Error("store.get(\"missing\") = true; want false")
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("update an existing entry", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)

		err := s.add(e)
		noErr(t, err)

		oldName := e.Name

		e.Name = "new-name"
		e.SourcePath = fmt.Sprintf("%s/new-source", rootPath)
		e.DestinationPath = "/new-path"

		err = s.update(oldName, e)
		noErr(t, err)

		retrievedEntry, _, err := s.get(e.Name)
		noErr(t, err)

		if retrievedEntry != e {
			t.Errorf("store.get(%q) = %+v, _, _; want %+v, _, _", e.Name, retrievedEntry, e)
		}
	})
	t.Run("remove the original entry", func(t *testing.T) {
		s, rootPath := newTestStore(t)

		e := newTestEntry(rootPath)

		err := s.add(e)
		noErr(t, err)

		oldName := e.Name

		e.Name = "new-name"
		e.SourcePath = fmt.Sprintf("%s/new-source", rootPath)
		e.DestinationPath = "/new-path"

		err = s.update(oldName, e)
		noErr(t, err)

		_, ok, err := s.get(oldName)
		noErr(t, err)

		if ok {
			t.Errorf("store.get(%q) = _, true, _; want _, false, _", oldName)
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
	fileName := fmt.Sprintf("source-file-%s", randomString(8))
	return link{
		SourcePath:      fmt.Sprintf("%s/%s", dir, fileName),
		DestinationPath: "/destination",
		Name:            fileName,
	}
}
