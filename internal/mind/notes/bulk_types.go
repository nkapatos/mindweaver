// NOTE: @ai exteract for agent doc
// ============================================================================
// Bulk Operations API Types
// ----------------------------------------------------------------------------
// References: backend/REFAC_GUIDELINES.md, backend/docs/BULK_OPERATIONS_GUIDE.md
//
// This file defines request/response types for bulk operations that affect
// multiple notes at once (e.g., rename tag across all notes, update metadata key).
//
// These operations are processed asynchronously using BadgerDB for tracking.
// Clients receive an operation_id immediately and can poll for status or
// use SSE streaming for real-time progress updates.
//
// Architecture:
// - BadgerDB stores operation state (like imex importer)
// - Worker processes operations in batches (100 notes per commit)
// - SSE endpoint streams progress events
// - Cleanup after 24 hours
// ============================================================================

package notes

// ============================================================================
// Metadata Bulk Operations
// ============================================================================

// BulkRenameMetaKeyReq is the request to rename a metadata key across all notes.
// Example: Rename "authot" → "author" across 50 notes.
type BulkRenameMetaKeyReq struct {
	OldKey string `json:"old_key"` // Required: current key name
	NewKey string `json:"new_key"` // Required: new key name
}

// BulkUpdateMetaValueReq is the request to update a metadata value across all notes.
// Example: Change project="Q1 Launch" → "Q1 Product Launch" across all notes.
type BulkUpdateMetaValueReq struct {
	Key      string `json:"key"`       // Required: metadata key to filter
	OldValue string `json:"old_value"` // Required: current value to match
	NewValue string `json:"new_value"` // Required: new value to set
}

// BulkDeleteMetaKeyReq is the request to delete a metadata key from all notes.
// Example: Remove deprecated "status" key from all notes.
type BulkDeleteMetaKeyReq struct {
	Key string `json:"key"` // Required: metadata key to delete
}

// ============================================================================
// Tag Bulk Operations
// ============================================================================

// BulkRenameTagReq is the request to rename a tag across all notes.
// Example: Rename #ai → #artificial-intelligence across 100 notes.
type BulkRenameTagReq struct {
	OldName string `json:"old_name"` // Required: current tag name
	NewName string `json:"new_name"` // Required: new tag name
}

// ============================================================================
// Bulk Operation Responses
// ============================================================================

// BulkOperationRes is the immediate response after queueing a bulk operation.
// Client uses operation_id to poll status or subscribe to SSE stream.
type BulkOperationRes struct {
	OperationID string `json:"operation_id"` // UUID for tracking
	Status      string `json:"status"`       // "pending" initially
	TotalItems  int    `json:"total_items"`  // Estimated number of notes to process
	Message     string `json:"message"`      // Human-readable description
}

// BulkOperationStatusRes is the response when querying operation status.
// Used for polling: GET /api/bulk-operations/:id/status
type BulkOperationStatusRes struct {
	OperationID     string  `json:"operation_id"`
	Type            string  `json:"type"`             // "metadata_rename", "tag_rename", etc.
	Status          string  `json:"status"`           // "pending", "in_progress", "completed", "failed"
	TotalItems      int     `json:"total_items"`      // Total notes to process
	ProcessedItems  int     `json:"processed_items"`  // Notes processed so far
	FailedItems     int     `json:"failed_items"`     // Notes that failed
	ProgressPercent float64 `json:"progress_percent"` // 0-100
	ErrorMessage    *string `json:"error_message,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	CompletedAt     *string `json:"completed_at,omitempty"`
}

// BulkOperationProgressEvent is the SSE event structure.
// Streamed via GET /api/bulk-operations/:id/stream
type BulkOperationProgressEvent struct {
	OperationID     string  `json:"operation_id"`
	Status          string  `json:"status"`
	ProcessedItems  int     `json:"processed_items"`
	TotalItems      int     `json:"total_items"`
	ProgressPercent float64 `json:"progress_percent"`
	Message         string  `json:"message"`
	Timestamp       string  `json:"timestamp"`
}
