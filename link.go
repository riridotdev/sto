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

func (l link) state() (linkState, error) {
	_, err := os.Lstat(l.destinationPath)
	if errors.Is(err, os.ErrNotExist) {
		return unlinked, nil
	}
	if err != nil {
		return unknown, fmt.Errorf("reading stat %q: %v", l.destinationPath, err)
	}

	resolvedPath, err := os.Readlink(l.destinationPath)
	if err != nil {
		return unknown, fmt.Errorf("resolving link %q: %v", l.destinationPath, err)
	}
	if resolvedPath != l.sourcePath {
		return conflict, nil
	}

	return linked, nil
}

type linkState byte

const (
	_ linkState = iota
	linked
	unlinked
	conflict
	unknown
)

func (ls linkState) String() string {
	switch ls {
	case linked:
		return "linked"
	case unlinked:
		return "unlinked"
	case conflict:
		return "conflict"
	case unknown:
		return "unknown"
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
