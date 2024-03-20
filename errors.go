package sto

import "errors"

var errEntryAlreadyExists = errors.New("Entry already exists")
var errEntryNotFound = errors.New("Entry not found")
var ErrEntrySourceInvalid = errors.New("Entry source invalid")
var ErrExistingFileAtDestination = errors.New("Existing file at destination")
var ErrLinkAlreadyExists = errors.New("Link already exists")
var errSourceAlreadyExists = errors.New("Source path already exists")

type ErrExistingSymlinkMismatch string

func (e ErrExistingSymlinkMismatch) Error() string {
	return "Existing symlink mismatch"
}
