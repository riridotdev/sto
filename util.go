package sto

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// userHomeMust returns the user's home directory.
//
// On error it will panic.
func UserHomeMust() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("error reading user home dir: %s", err))
	}
	return home
}

// trimStoRoot returns the input path with its sto root prefix removed.
//
// The input path must be an absolute path with the stoRoot included.
// Calling this function with any other path will result in a panic.
func trimStoRoot(path string, root string) string {
	if !filepath.IsAbs(path) {
		panic(fmt.Sprintf("path %q is not an absolute path", path))
	}
	if !strings.HasPrefix(path, root) {
		panic(fmt.Sprintf("path %q does not have prefix %q", path, root))
	}
	return strings.TrimPrefix(path, root)[1:]
}

// fail prints the provided formatted string to stdout and terminates the current process with an exit code of 1.
func Fail(format string, params ...interface{}) {
	fmt.Println(fmt.Sprintf(format, params...))
	os.Exit(1)
}

// readStoreOrFail returns the store found at the provided root directory.
//
// On fail it will print an error message to stdout and terminate the current process with an exit code of 1.
func ReadStoreOrFail(root string) store {
	s, err := readStore(root)
	if err != nil {
		Fail("Error reading store: %s", err)
	}
	return s
}
