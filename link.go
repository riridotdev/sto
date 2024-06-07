package main

import (
	"errors"
	"fmt"
	"os"
)

type link struct {
	sourcePath      string
	destinationPath string
}

func (l link) link() error {
	if err := os.Symlink(l.sourcePath, l.destinationPath); err != nil {
		return fmt.Errorf("creating symlink %q -> %q: %v", l.destinationPath, l.sourcePath, err)
	}
	return nil
}

func (l link) unlink() error {
	if err := os.Remove(l.destinationPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing symlink at %q: %v", l.destinationPath, err)
	}
	return nil
}
