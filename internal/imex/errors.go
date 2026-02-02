package imex

import (
	"fmt"
	"sync"
)

// ErrorType categorizes different failure modes
type ErrorType int

const (
	ErrorTypeUnknown    ErrorType = iota
	ErrorTypeIO                   // File read/write errors
	ErrorTypePermission           // Permission denied
	ErrorTypeEncoding             // Invalid UTF-8 or encoding issues
	ErrorTypeSize                 // File too large
	ErrorTypeNotFound             // File not found
	ErrorTypeTransport            // Transport/network errors when sending to Mind
)

func (e ErrorType) String() string {
	switch e {
	case ErrorTypeIO:
		return "IO Error"
	case ErrorTypePermission:
		return "Permission Denied"
	case ErrorTypeEncoding:
		return "Encoding Error"
	case ErrorTypeSize:
		return "File Too Large"
	case ErrorTypeNotFound:
		return "File Not Found"
	case ErrorTypeTransport:
		return "Transport Error"
	default:
		return "Unknown Error"
	}
}

// ImportError wraps errors with additional context
type ImportError struct {
	Path       string
	Type       ErrorType
	Underlying error
}

func (e *ImportError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Path, e.Type, e.Underlying)
}

// ErrorTracker manages error collection during import
// Thread-safe for concurrent use
type ErrorTracker struct {
	mu     sync.RWMutex
	errors map[string]*ImportError
}

// NewErrorTracker creates a new error tracker
func NewErrorTracker() *ErrorTracker {
	return &ErrorTracker{
		errors: make(map[string]*ImportError),
	}
}

// Add records an error for a specific file path
func (et *ErrorTracker) Add(path string, errType ErrorType, err error) {
	et.mu.Lock()
	defer et.mu.Unlock()

	et.errors[path] = &ImportError{
		Path:       path,
		Type:       errType,
		Underlying: err,
	}
}

// Get retrieves the error for a specific file path
func (et *ErrorTracker) Get(path string) (*ImportError, bool) {
	et.mu.RLock()
	defer et.mu.RUnlock()

	err, ok := et.errors[path]
	return err, ok
}

// GetAll returns all recorded errors
func (et *ErrorTracker) GetAll() map[string]error {
	et.mu.RLock()
	defer et.mu.RUnlock()

	result := make(map[string]error, len(et.errors))
	for path, err := range et.errors {
		result[path] = err
	}
	return result
}

// Count returns the total number of errors
func (et *ErrorTracker) Count() int {
	et.mu.RLock()
	defer et.mu.RUnlock()

	return len(et.errors)
}

// HasErrors returns true if any errors were recorded
func (et *ErrorTracker) HasErrors() bool {
	return et.Count() > 0
}
