package main

import (
	"fmt"
	"os"
	"testing"
)

func TestCompress(t *testing.T) {
	t.Run("replace homedir segment with '~'", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		noErr(t, err)

		path := fmt.Sprintf("%s/testpath", homeDir)

		compressedPath, err := compress(path)
		noErr(t, err)

		wantPath := "~/testpath"

		if compressedPath != wantPath {
			t.Errorf("compress(%q) = %q; want %q", path, compressedPath, wantPath)
		}
	})
	t.Run("return original path when no homedir segment", func(t *testing.T) {
		path := "/test-dir/test-file"

		compressedPath, err := compress(path)
		noErr(t, err)

		if compressedPath != path {
			t.Errorf("compress(%q) = %q; want %q", path, compressedPath, path)
		}
	})
}

func TestExpand(t *testing.T) {
	t.Run("expand '~' to full homedir", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		noErr(t, err)

		path := "~/testpath"

		expandedPath, err := expand(path)
		noErr(t, err)

		wantPath := fmt.Sprintf("%s/testpath", homeDir)

		if expandedPath != wantPath {
			t.Errorf("expand(%q) = %q; want %q", path, expandedPath, wantPath)
		}
	})
	t.Run("return original path when no '~'", func(t *testing.T) {
		path := "/test-dir/test-file"

		expandedPath, err := expand(path)
		noErr(t, err)

		if expandedPath != path {
			t.Errorf("expand(%q) = %q; want %q", path, expandedPath, path)
		}
	})
	t.Run("return input when given an empty string", func(t *testing.T) {
		expandedPath, err := expand("")
		noErr(t, err)

		if expandedPath != "" {
			t.Errorf("expand(\"\") = %q; want \"\"", expandedPath)
		}
	})
}
