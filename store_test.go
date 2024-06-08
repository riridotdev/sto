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
