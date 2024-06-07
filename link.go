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
	stat, err := os.Lstat(l.destinationPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading link stat for %q: %v", l.destinationPath, err)
	}
	if err == nil {
		if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
			return conflictingItemError(l.destinationPath)
		}
		resolvedLink, err := os.Readlink(l.destinationPath)
		if err != nil {
			return fmt.Errorf("resolving symlink at %q: %v", l.destinationPath, err)
		}
		if resolvedLink != l.sourcePath {
			return conflictingLinkError(link{
				sourcePath:      resolvedLink,
				destinationPath: l.destinationPath,
			})
		}
		return nil
	}

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

func (l link) state() linkState {
	return unlinked
}

type linkState byte

const (
	_ linkState = iota
	unlinked
)

func (ls linkState) String() string {
	switch ls {
	case unlinked:
		return "unlinked"
	default:
		panic(fmt.Sprintf("unrecognised state: %d", ls))
	}
}

type conflictingLinkError link

func (e conflictingLinkError) Error() string {
	return fmt.Sprintf("conflicting symlink %q -> %q", e.destinationPath, e.sourcePath)
}

type conflictingItemError string

func (e conflictingItemError) Error() string {
	return fmt.Sprintf("conflicting item at %q", string(e))
}
