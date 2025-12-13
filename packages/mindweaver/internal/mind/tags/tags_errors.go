package tags

// Tags Domain Errors
// Domain-specific errors for the tags service layer

import (
	"errors"
)

// Domain errors for tags service
var (
	// ErrTagAlreadyExists indicates a tag with the same name already exists
	ErrTagAlreadyExists = errors.New("tag already exists")

	// ErrTagNotFound indicates a tag was not found
	ErrTagNotFound = errors.New("tag not found")
)
