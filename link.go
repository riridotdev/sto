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
	state, err := l.state()
	if err != nil {
		return err
	}

	switch state {
	case linked:
		return nil
	case conflictingItem:
		return conflictingItemError(l.destinationPath)
	case conflictingLink:
		return conflictingLinkError(l.destinationPath)
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
	_, err := os.Stat(l.sourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return broken, nil
	}
	if err != nil {
		return unknown, fmt.Errorf("reading stat %q: %v", l.sourcePath, err)
	}

	state, err := os.Lstat(l.destinationPath)
	if errors.Is(err, os.ErrNotExist) {
		return unlinked, nil
	}
	if err != nil {
		return unknown, fmt.Errorf("reading stat %q: %v", l.destinationPath, err)
	}
	if state.Mode()&os.ModeSymlink != os.ModeSymlink {
		return conflictingItem, nil
	}

	resolvedPath, err := os.Readlink(l.destinationPath)
	if err != nil {
		return unknown, fmt.Errorf("resolving link %q: %v", l.destinationPath, err)
	}
	if resolvedPath != l.sourcePath {
		return conflictingLink, nil
	}

	return linked, nil
}

type linkState byte

const (
	_ linkState = iota
	linked
	unlinked
	conflictingItem
	conflictingLink
	broken
	unknown
)

func (ls linkState) String() string {
	switch ls {
	case linked:
		return "linked"
	case unlinked:
		return "unlinked"
	case conflictingItem:
		return "conflictingItem"
	case conflictingLink:
		return "conflictingLink"
	case broken:
		return "broken"
	case unknown:
		return "unknown"
	default:
		panic(fmt.Sprintf("unrecognised state: %d", ls))
	}
}

type conflictingLinkError string

func (e conflictingLinkError) Error() string {
	return fmt.Sprintf("conflicting symlink at %q", string(e))
}

type conflictingItemError string

func (e conflictingItemError) Error() string {
	return fmt.Sprintf("conflicting item at %q", string(e))
}
