package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestLink(t *testing.T) {
	t.Run("create a symlink as defined by the link", func(t *testing.T) {
		l := newTestLink(t)

		err := l.link()
		noErr(t, err)

		resolvedPath, err := os.Readlink(l.DestinationPath)
		noErr(t, err)

		if resolvedPath != l.SourcePath {
			t.Errorf("os.Readlink(%q) = %q, _; want %q, _", l.DestinationPath, resolvedPath, l.SourcePath)
		}
	})
	t.Run("behave idempotently when linking", func(t *testing.T) {
		l := newTestLink(t)

		err := l.link()
		noErr(t, err)

		if err := l.link(); err != nil {
			t.Errorf("link.link() = %q; want nil", err)
		}
	})
	t.Run("fail when existing symlink at destination resolves to a different source", func(t *testing.T) {
		l := newTestLink(t)

		err := l.link()
		noErr(t, err)

		conflictingLink := newTestLink(t)
		conflictingLink.DestinationPath = l.DestinationPath

		wantErr := conflictingLinkError(l.DestinationPath)

		if err := conflictingLink.link(); err != wantErr {
			t.Errorf("link.link = %q; want %q", err, wantErr)
		}
	})
	t.Run("fail when file exists at destination", func(t *testing.T) {
		l := newTestLink(t)

		f, err := os.Create(l.DestinationPath)
		noErr(t, err)
		f.Close()
		defer os.Remove(l.DestinationPath)

		wantErr := conflictingItemError(l.DestinationPath)

		if err := l.link(); err != wantErr {
			t.Errorf("link.link = %q; want %q", err, wantErr)
		}
	})
	t.Run("fail when source file does not exist", func(t *testing.T) {
		l := newTestLink(t)
		removeFile(t, l.SourcePath)

		wantErr := sourceMissingError(l.SourcePath)

		if err := l.link(); !errors.Is(err, wantErr) {
			t.Errorf("link.link = %q; want %q", err, wantErr)
		}
	})
}

func TestUnlink(t *testing.T) {
	t.Run("remove an existing symlink at the destination", func(t *testing.T) {
		l := newTestLink(t)

		err := l.link()
		noErr(t, err)

		err = l.unlink()
		noErr(t, err)

		_, err = os.Readlink(l.DestinationPath)

		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("os.ReadLink(%q) = _, %q; want _, %q", l.DestinationPath, err, os.ErrNotExist)
		}
	})
	t.Run("behave idempotently when unlinking", func(t *testing.T) {
		l := newTestLink(t)

		err := l.link()
		noErr(t, err)

		err = l.unlink()
		noErr(t, err)

		if err := l.unlink(); err != nil {
			t.Errorf("link.unlink() = %q; want nil", err)
		}
	})
}

func TestState(t *testing.T) {
	t.Run("returns unlinked when unlinked", func(t *testing.T) {
		l := newTestLink(t)

		state, err := l.state()
		noErr(t, err)

		if state != unlinked {
			t.Errorf("link.state() = %s; want %s", state, unlinked)
		}
	})
	t.Run("returns linked when linked", func(t *testing.T) {
		l := newTestLink(t)

		err := l.link()
		noErr(t, err)

		state, err := l.state()
		noErr(t, err)

		if state != linked {
			t.Errorf("link.state() = %s; want %s", state, linked)
		}
	})
	t.Run("returns conflict when another link exists at the destination", func(t *testing.T) {
		conflictLink := newTestLink(t)

		err := conflictLink.link()
		noErr(t, err)

		l := newTestLink(t)
		l.DestinationPath = conflictLink.DestinationPath

		state, err := l.state()
		noErr(t, err)

		if state != conflictingLink {
			t.Errorf("link.state() = %s; want %s", state, conflictingLink)
		}
	})
	t.Run("returns conflict when a file exists at the destination", func(t *testing.T) {
		l := newTestLink(t)

		f, err := os.Create(l.DestinationPath)
		noErr(t, err)
		f.Close()
		defer removeFile(t, l.DestinationPath)

		state, err := l.state()
		noErr(t, err)

		if state != conflictingItem {
			t.Errorf("link.state() = %s; want %s", state, conflictingItem)
		}
	})
	t.Run("return broken when source file is missing", func(t *testing.T) {
		l := newTestLink(t)
		removeFile(t, l.SourcePath)

		state, err := l.state()
		noErr(t, err)

		if state != sourceMissing {
			t.Errorf("link.state() = %s; want %s", state, sourceMissing)
		}
	})
}

func newTestLink(t *testing.T) link {
	t.Helper()

	sourcePath := newTestFile(t, "")

	dir := t.TempDir()
	destinationPath := fmt.Sprintf("%s/test-link", dir)

	l := link{
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
	}

	t.Cleanup(func() {
		if err := l.unlink(); err != nil {
			t.Fatalf("cleaning up: unlinking link: %v", err)
		}
	})

	return link{
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
	}
}
