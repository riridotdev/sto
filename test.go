package main

import (
	"os"
	"testing"
)

func noErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func newTestFile(t testing.TB, dir string) string {
	t.Helper()

	if dir == "" {
		dir = t.TempDir()
	}

	file, err := os.CreateTemp(dir, "test-file")
	if err != nil {
		t.Fatalf("creating test file in %q: %v", dir, err)
	}
	file.Close()

	filePath := file.Name()

	t.Cleanup(func() {
		if err := os.Remove(filePath); err != nil {
			t.Fatalf("cleaning up: removing file %q: %v", filePath, err)
		}
	})

	return filePath
}
