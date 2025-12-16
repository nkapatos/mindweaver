// Package transport provides abstractions for sending imported file data
// to various destinations (HTTP, cloud storage, etc.)
//
// Response Handling (Updated 2025-11-29):
// - Mind service returns single resources directly: types.Response[Operation]
// - HTTP transport parses: response.data is the Operation directly
// - Errors are in response.error field
// - NO items[] wrapping for single resources (aligned with Google AIP-151)
package transport

import (
	"time"
)

// FileData represents a file to be transported
// This is the core data structure shared between IMEX and Mind
type FileData struct {
	Path    string    `json:"path"`     // Original file path (relative to import root)
	Content []byte    `json:"content"`  // Raw file content
	Hash    string    `json:"hash"`     // xxHash64 hex string of content
	Size    int64     `json:"size"`     // File size in bytes
	ModTime time.Time `json:"mod_time"` // File modification time
}

// BatchRequest represents a batch of files being sent to Mind
// This is NOT wrapped - it's the request payload
type BatchRequest struct {
	BatchID  string         `json:"batch_id"`         // Unique identifier for this batch
	Files    []FileData     `json:"files"`            // Files in this batch
	Metadata map[string]any `json:"metadata"`         // Optional metadata (source, version, etc.)
	Parent   string         `json:"parent,omitempty"` // Parent resource (e.g., "users/123")
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
