package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"testing"
)

func noErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func assert(t testing.TB, b bool) {
	t.Helper()
	if !b {
		t.Fatalf("failed assertion")
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
		if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("cleaning up: removing file %q: %v", filePath, err)
		}
	})

	return filePath
}

func removeFile(t testing.TB, path string) {
	t.Helper()

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleting file %q: %v", path, err)
	}
}

func randomString(length int) string {
	s := make([]byte, length)

	n, err := rand.Read(s)
	if err != nil {
		panic(fmt.Sprintf("creating random string"))
	}
	if n != length {
		panic(fmt.Sprintf("read %d bytes; expected %d", length, n))
	}

	for i := range s {
		s[i] = s[i]%26 + 'a'
	}

	return string(s)
}
