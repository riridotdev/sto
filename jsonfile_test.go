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
