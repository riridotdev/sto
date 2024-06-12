package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewJsonFile(t *testing.T) {
	t.Run("create non-existing file", func(t *testing.T) {
		dir := t.TempDir()

		filePath := fmt.Sprintf("%s/test-file", dir)

		_, err := newJsonFile(filePath)
		noErr(t, err)

		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("os.Stat(%q) = _, %v; want _, nil", filePath, err)
		}
	})
	t.Run("leave an existing file untouched", func(t *testing.T) {
		dir := t.TempDir()

		filePath := fmt.Sprintf("%s/test-file", dir)

		f, err := os.Create(filePath)
		noErr(t, err)

		wantBytes := []byte("hello world")

		_, err = f.Write(wantBytes)
		noErr(t, err)

		err = f.Close()
		noErr(t, err)

		_, err = newJsonFile(filePath)
		noErr(t, err)

		fileBytes, err := os.ReadFile(filePath)
		noErr(t, err)

		if !cmp.Equal(fileBytes, wantBytes) {
			t.Errorf("fileBytes != wantBytes\n%s", cmp.Diff(fileBytes, wantBytes))
		}
	})
}

func TestRead(t *testing.T) {
	t.Run("read an empty file", func(t *testing.T) {
		jf := newTestJsonFile(t)

		var items []string

		err := jf.read(&items)
		noErr(t, err)

		if len(items) != 0 {
			t.Errorf("len(items) = %d; want 0", len(items))
		}
	})
}

func newTestJsonFile(t *testing.T) jsonFile {
	dir := t.TempDir()

	filePath := fmt.Sprintf("%s/test-file", dir)

	jf, err := newJsonFile(filePath)
	noErr(t, err)

	return jf
}
