// Package transport provides abstractions for sending imported file data
// to various destinations (HTTP, cloud storage, etc.)
package transport

import (
	"context"
	"time"
)

// Transport defines the interface for sending batches of files to a destination
// Implementations include HTTP (Mind service), S3, GCS, Azure Blob, etc.
//
// Note: All responses from Mind are wrapped in types.Response[T].
// Transport implementations handle unwrapping and error extraction.
type Transport interface {
	// Send sends a batch of files to the destination
	// Returns an Operation (long-running operation handle)
	// The implementation should handle retries for transient failures
	// On error, returns Go error (not the wrapped response error)
	Send(ctx context.Context, batch *BatchRequest) (*Operation, error)

	// GetOperation retrieves the current state of a long-running operation
	// Returns nil if the operation is not found
	// On error, returns Go error
	GetOperation(ctx context.Context, operationName string) (*Operation, error)

	// WaitForCompletion polls an operation until it completes
	// Returns the final Operation with done=true
	// pollInterval is how often to check (e.g., 2s)
	// On error or timeout, returns Go error
	WaitForCompletion(ctx context.Context, operationName string, pollInterval time.Duration) (*Operation, error)

	// Close cleans up any resources used by the transport
	Close() error

	// Stats returns transport statistics (bytes sent, latency, etc.)
	Stats() TransportStats
}

// Batcher handles batching logic for efficient transport
// It groups files into optimally-sized batches based on size/count limits
type Batcher interface {
	// Add adds a file to the current batch
	// Returns true if the batch is full and should be sent
	Add(file FileData) bool

	// Flush returns the current batch and resets for the next batch
	// Returns nil if the batch is empty
	Flush() *BatchRequest

	// IsFull returns true if the current batch is full
	IsFull() bool

	// Count returns the number of files in the current batch
	Count() int
}

// BatcherConfig configures batch size limits
type BatcherConfig struct {
	MaxFiles     int   // Maximum files per batch (default: 50)
	MaxSizeBytes int64 // Maximum total size per batch (default: 10MB)
}
