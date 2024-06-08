package main

import (
	"errors"
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
		s := newTestStore(t)

		entries := s.entries()

		if len(entries) != 0 {
			t.Errorf("len(entries) = %d; want 0", len(entries))
		}
	})
}

func TestAdd(t *testing.T) {
	t.Run("add a new link", func(t *testing.T) {
		s := newTestStore(t)

		l := link{
			sourcePath:      "test",
			destinationPath: "test",
		}

		s.add(l)

		entries := s.entries()

		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d; want 1", len(entries))
		}
		if entries[0] != l {
			t.Errorf("entries[0] = %+v; want %+v", l, entries[0])
		}
	})
	t.Run("behave idempotently when adding links", func(t *testing.T) {
		s := newTestStore(t)

		l := link{
			sourcePath:      "test",
			destinationPath: "test",
		}

		s.add(l)
		s.add(l)

		entries := s.entries()

		if len(entries) != 1 {
			t.Fatalf("len(entries) = %d; want 1", len(entries))
		}
		if entries[0] != l {
			t.Errorf("entries[0] = %+v; want %+v", l, entries[0])
		}
	})
}

func newTestStore(t *testing.T) store {
	dir := t.TempDir()

	s, err := initStore(dir)
	noErr(t, err)

	return s
}
