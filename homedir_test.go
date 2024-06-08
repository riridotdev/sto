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
}
