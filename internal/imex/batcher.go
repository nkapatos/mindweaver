package imex

import (
	"github.com/google/uuid"
)

// BatchRequest represents a batch of files being sent to Mind
// This is NOT wrapped - it's the request payload
type BatchRequest struct {
	BatchID  string         `json:"batch_id"`         // Unique identifier for this batch
	Files    []FileData     `json:"files"`            // Files in this batch
	Metadata map[string]any `json:"metadata"`         // Optional metadata (source, version, etc.)
	Parent   string         `json:"parent,omitempty"` // Parent resource (e.g., "users/123")
}

// BatchResult represents the final result of a completed batch import
type BatchResult struct {
	BatchID    string       `json:"batch_id"`
	TotalFiles int          `json:"total_files"`
	Processed  int          `json:"processed"`
	Skipped    int          `json:"skipped"` // Files skipped (duplicates)
	Failed     int          `json:"failed"`  // Files that failed to process
	Errors     []BatchError `json:"errors,omitempty"`
	Duration   float64      `json:"duration_seconds"`
}

// BatchError represents an error for a specific file during batch processing
// This aligns with types.ErrorDetail but is specific to file errors
type BatchError struct {
	File         string `json:"file"`                    // File path that failed
	Reason       string `json:"reason"`                  // Machine-readable reason (e.g., "invalid_utf8")
	Message      string `json:"message"`                 // Human-readable error message
	LocationType string `json:"location_type,omitempty"` // "file"
}

// SimpleBatcher implements Batcher interface with basic size/count limits
type SimpleBatcher struct {
	config       BatcherConfig
	currentBatch *BatchRequest
	currentSize  int64
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

// NewSimpleBatcher creates a new simple batcher
func NewSimpleBatcher(config BatcherConfig) *SimpleBatcher {
	// Set defaults
	if config.MaxFiles == 0 {
		config.MaxFiles = 50
	}
	if config.MaxSizeBytes == 0 {
		config.MaxSizeBytes = 10 * 1024 * 1024 // 10MB
	}

	return &SimpleBatcher{
		config: config,
		currentBatch: &BatchRequest{
			BatchID: uuid.New().String(),
			Files:   make([]FileData, 0, config.MaxFiles),
		},
		currentSize: 0,
	}
}

// Add implements Batcher.Add
func (b *SimpleBatcher) Add(file FileData) bool {
	b.currentBatch.Files = append(b.currentBatch.Files, file)
	b.currentSize += file.Size

	return b.IsFull()
}

// Flush implements Batcher.Flush
func (b *SimpleBatcher) Flush() *BatchRequest {
	if len(b.currentBatch.Files) == 0 {
		return nil
	}

	batch := b.currentBatch

	// Reset for next batch
	b.currentBatch = &BatchRequest{
		BatchID: uuid.New().String(),
		Files:   make([]FileData, 0, b.config.MaxFiles),
	}
	b.currentSize = 0

	return batch
}

// IsFull implements Batcher.IsFull
func (b *SimpleBatcher) IsFull() bool {
	return len(b.currentBatch.Files) >= b.config.MaxFiles ||
		b.currentSize >= b.config.MaxSizeBytes
}

// Count implements Batcher.Count
func (b *SimpleBatcher) Count() int {
	return len(b.currentBatch.Files)
}
