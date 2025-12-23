package links

// Links Domain Errors
// Domain-specific errors for the links service layer

import (
	"errors"
)

// Domain errors for links service
var (
	// ErrLinkNotFound indicates a link was not found
	ErrLinkNotFound = errors.New("link not found")
)
