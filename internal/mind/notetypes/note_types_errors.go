// NoteTypes Domain Errors
package notetypes

import "errors"

var (
	// ErrNoteTypeNotFound is returned when a note type is not found
	ErrNoteTypeNotFound = errors.New("note type not found")

	// ErrNoteTypeAlreadyExists is returned when creating/updating a note type with a duplicate type identifier
	ErrNoteTypeAlreadyExists = errors.New("note type already exists")

	// ErrNoteTypeIsSystem is returned when attempting to delete a system note type
	ErrNoteTypeIsSystem = errors.New("cannot delete system note type")
)
