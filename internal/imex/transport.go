package imex

import (
	"context"
	"sync"
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

// Operation represents a long-running operation (Google AIP-151 pattern)
// Returned by Mind service when batch import is queued
type Operation struct {
	Name       string         `json:"name"`               // "operations/uuid"
	Done       bool           `json:"done"`               // true when complete
	StatusLink string         `json:"status_link"`        // URL to poll for status (e.g., "/api/import/operations/uuid")
	Metadata   *OperationMeta `json:"metadata,omitempty"` // Progress information

	// One of these will be set when done=true:
	Result *BatchResult `json:"result,omitempty"` // Success result
	// Error is in the Response[Operation].Error field (handled by middleware)
}

// OperationMeta provides progress information for ongoing operations
type OperationMeta struct {
	BatchID        string    `json:"batch_id"`
	Status         string    `json:"status"`          // "queued" | "processing" | "completed" | "failed"
	ProgressPct    int       `json:"progress_pct"`    // 0-100
	ProcessedFiles int       `json:"processed_files"` // Files processed so far
	TotalFiles     int       `json:"total_files"`     // Total files in batch
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

// TransportStats tracks statistics for a transport session
type TransportStats struct {
	TotalBatches   int           // Total batches sent
	TotalFiles     int           // Total files sent
	TotalBytes     int64         // Total bytes sent
	SuccessCount   int           // Successfully sent batches
	FailureCount   int           // Failed batches
	RetryCount     int           // Number of retries
	Duration       time.Duration // Total time spent
	AverageLatency time.Duration // Average latency per batch
}

type TransportConfig struct {
	BaseURL      string        // Base URL for the transport (e.g., Mind service endpoint)
	APIKey       string        // API key or token for authentication
	Timeout      time.Duration // Request timeout
	MaxRetries   int           // Number of retries for transient failures
	RetryBackoff time.Duration // Backoff duration between retries
	UserAgent    string        // User-Agent header value
}

// DirectTransport is a no-op transport that simulates sending files directly
// Useful for testing or dry-run scenarios
type DirectTransport struct {
	config TransportConfig
	client any // Placeholder for HTTP client or other transport client
	mu     sync.Mutex
	stats  TransportStats
}

// NewDirectTransport creates a new DirectTransport with the given config
func NewDirectTransport(config TransportConfig) *DirectTransport {
	// Default config values
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryBackoff == 0 {
		config.RetryBackoff = 2 * time.Second
	}
	if config.UserAgent == "" {
		config.UserAgent = "imex-direct-transport/1.0"
	}

	return &DirectTransport{
		config: config,
		client: nil, // Initialize actual client if needed
		stats:  TransportStats{},
	}
}

// Send simulates sending a batch of files
func (dt *DirectTransport) Send(ctx context.Context, batch *BatchRequest) (*Operation, error) {
	dt.mu.Lock()
	defer dt.mu.Unlock()

	// Simulate sending by updating stats
	dt.stats.TotalBatches++
	dt.stats.TotalFiles += len(batch.Files)
	var batchSize int64
	for _, file := range batch.Files {
		batchSize += file.Size
	}
	dt.stats.TotalBytes += batchSize
	dt.stats.SuccessCount++

	// Simulate operation response
	op := &Operation{
		Name:       "operations/simulated-op-12345",
		Done:       true,
		StatusLink: "/api/import/operations/simulated-op-12345",
		Result: &BatchResult{
			TotalFiles: len(batch.Files),
			Processed:  len(batch.Files),
			Skipped:    0,
			Failed:     0,
			Duration:   0.1,
		},
	}
	return op, nil
}

// GetOperation simulates retrieving an operation
func (dt *DirectTransport) GetOperation(ctx context.Context, operationName string) (*Operation, error) {
	// Simulate immediate completion
	op := &Operation{
		Name: operationName,
		Done: true,
		Result: &BatchResult{
			TotalFiles: 0,
			Processed:  0,
			Skipped:    0,
			Failed:     0,
			Duration:   0.1,
		},
	}
	return op, nil
}

// WaitForCompletion simulates waiting for an operation to complete
func (dt *DirectTransport) WaitForCompletion(ctx context.Context, operationName string, pollInterval time.Duration) (*Operation, error) {
	// Simulate immediate completion
	return dt.GetOperation(ctx, operationName)
}

// Close cleans up resources (no-op for DirectTransport)
func (dt *DirectTransport) Close() error {
	return nil
}

// Stats returns transport statistics
func (dt *DirectTransport) Stats() TransportStats {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	return dt.stats
}
