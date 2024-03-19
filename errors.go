package main

import "errors"

var errEntryAlreadyExists = errors.New("Entry already exists")
var errEntryNotFound = errors.New("Entry not found")
var errEntrySourceInvalid = errors.New("Entry source invalid")
var errExistingFileAtDestination = errors.New("Existing file at destination")
var errLinkAlreadyExists = errors.New("Link already exists")
var errSourceAlreadyExists = errors.New("Source path already exists")

type errExistingSymlinkMismatch string

func (e errExistingSymlinkMismatch) Error() string {
	return "Existing symlink mismatch"
}
