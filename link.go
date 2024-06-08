package main

import (
	"errors"
	"fmt"
	"os"
)

type link struct {
	SourcePath      string
	DestinationPath string
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
		return conflictingItemError(l.DestinationPath)
	case conflictingLink:
		return conflictingLinkError(l.DestinationPath)
	case sourceMissing:
		return sourceMissingError(l.SourcePath)
	}

	if err := os.Symlink(l.SourcePath, l.DestinationPath); err != nil {
		return fmt.Errorf("creating symlink %q -> %q: %v", l.DestinationPath, l.SourcePath, err)
	}
	return nil
}

func (l link) unlink() error {
	if err := os.Remove(l.DestinationPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing symlink at %q: %v", l.DestinationPath, err)
	}
	return nil
}

func (l link) state() (linkState, error) {
	_, err := os.Stat(l.SourcePath)
	if errors.Is(err, os.ErrNotExist) {
		return sourceMissing, nil
	}
	if err != nil {
		return unknown, fmt.Errorf("reading stat %q: %v", l.SourcePath, err)
	}

	state, err := os.Lstat(l.DestinationPath)
	if errors.Is(err, os.ErrNotExist) {
		return unlinked, nil
	}
	if err != nil {
		return unknown, fmt.Errorf("reading stat %q: %v", l.DestinationPath, err)
	}
	if state.Mode()&os.ModeSymlink != os.ModeSymlink {
		return conflictingItem, nil
	}

	resolvedPath, err := os.Readlink(l.DestinationPath)
	if err != nil {
		return unknown, fmt.Errorf("resolving link %q: %v", l.DestinationPath, err)
	}
	if resolvedPath != l.SourcePath {
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
	sourceMissing
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
	case sourceMissing:
		return "sourceMissing"
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

type sourceMissingError string

func (e sourceMissingError) Error() string {
	return fmt.Sprintf("source %q is missing", string(e))
}
