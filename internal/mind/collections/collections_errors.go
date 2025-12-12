package collections

import "errors"

// Domain errors for Collections
var (
	// ErrCollectionNotFound is returned when a collection is not found.
	ErrCollectionNotFound = errors.New("collection not found")

	// ErrCollectionAlreadyExists is returned when a collection with the same path already exists.
	ErrCollectionAlreadyExists = errors.New("collection already exists")

	// ErrCollectionIsSystem is returned when attempting to delete a system collection.
	ErrCollectionIsSystem = errors.New("cannot delete system collection")

	// ErrInvalidParentCollection is returned when parent_id references a non-existent collection.
	ErrInvalidParentCollection = errors.New("invalid parent collection")
)
