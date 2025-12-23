package meta

// NoteMeta Domain Errors
// Domain-specific errors for the note metadata service layer

import (
	"errors"
)

// Domain errors for note metadata service
var (
	// ErrNoteMetaNotFound indicates note metadata was not found
	ErrNoteMetaNotFound = errors.New("note metadata not found")
)
