package notes

import "errors"

// Domain errors for Notes
var (
	// ErrNoteNotFound is returned when a note is not found.
	ErrNoteNotFound = errors.New("note not found")

	// ErrNoteAlreadyExists is returned when a note with the same UUID already exists.
	ErrNoteAlreadyExists = errors.New("note already exists")

	// ErrInvalidCollectionID is returned when collection_id references a non-existent collection.
	ErrInvalidCollectionID = errors.New("invalid collection id")

	// ErrInvalidNoteTypeID is returned when note_type_id references a non-existent note type.
	ErrInvalidNoteTypeID = errors.New("invalid note type id")

	// ErrStaleNote is returned when the note version doesn't match (optimistic locking failure).
	ErrStaleNote = errors.New("note has been modified by another request")

	// ErrInvalidTitle is returned when the title is empty or exceeds max length.
	ErrInvalidTitle = errors.New("invalid title")

	// ErrInvalidDescription is returned when the description exceeds max length.
	ErrInvalidDescription = errors.New("invalid description")
)
