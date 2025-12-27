package templates

// Templates Domain Errors
// Domain-specific errors for the templates service layer

import (
	"errors"
)

// Domain errors for templates service
var (
	// ErrTemplateAlreadyExists indicates a template with the same name already exists
	ErrTemplateAlreadyExists = errors.New("template already exists")

	// ErrTemplateNotFound indicates a template was not found
	ErrTemplateNotFound = errors.New("template not found")
)
