// Package types provides the wrapping response types following Google's API Improvement Proposals (AIP).
// We use a consistent response envelope (data/error wrapper) for better client developer experience.
//
// DEPRECATED: This package is for legacy REST APIs only (V1, V2).
// For new APIs, use Protocol Buffers definitions in proto/ directory.
//
// Migration path:
// - V1/V2 APIs: Continue using pkg/types (handwritten REST)
// - V3+ APIs: Use proto definitions with Connect-RPC (generated types, automatic validation)
//
// Proto-based APIs (V3+) provide:
// - Single source of truth (like SQL+sqlc)
// - Automatic validation via protovalidate
// - Both gRPC and REST/JSON via Connect-RPC
// - AIP compliance enforced at proto level
//
// Key AIPs followed:
// - AIP-121: Resource-oriented design - https://google.aip.dev/121
// - AIP-122: Resource names - https://google.aip.dev/122
// - AIP-131-135: Standard methods - https://google.aip.dev/131
// - AIP-140: Field names (snake_case) - https://google.aip.dev/140
// - AIP-158: Pagination (cursor-based) - https://google.aip.dev/158
// - AIP-193: Errors - https://google.aip.dev/193
//
// Documentation: docs/DEVELOPMENT.md (API Design Guidelines section)
package types

// Response is the top-level API response wrapper providing consistent structure for all endpoints.
// This wrapper is not strictly required by AIP but improves client developer experience by:
//   - Providing consistent top-level structure across all responses
//   - Enabling type-safe generic handling in clients
//   - Supporting partial success scenarios (both data and error present)
//
// Can contain either data (success), error (failure), or both (partial success in batch operations).
//
// Usage patterns:
//   - Single resource (Get, Create, Update): Response[Note] where T is the resource
//   - Collection (List): Response[ListResult[Note]] where T is ListResult
//   - Operation result: Response[OperationResult] for minimal responses
type Response[T any] struct {
	Data  *T             `json:"data,omitempty"`  // The response data (resource or collection)
	Error *ErrorResponse `json:"error,omitempty"` // Error information if request failed
}

// ListResult wraps a collection of items with pagination metadata.
// Use this for List operations that return multiple resources.
// Follows AIP-158 cursor-based pagination for scalability.
//
// Example:
//
//	Response[ListResult[Note]] for GET /api/notes (list)
type ListResult[T any] struct {
	Kind          string `json:"kind,omitempty"`            // Resource type with #list suffix (e.g., "note#list")
	Items         []T    `json:"items"`                     // Array of resources (never null, empty array if no items)
	NextPageToken string `json:"next_page_token,omitempty"` // Token for next page (AIP-158 cursor-based pagination)
	TotalSize     int    `json:"total_size,omitempty"`      // Total number of items (optional, may be expensive to compute)
}

// SingleResult wraps a single resource with metadata.
// DEPRECATED: This wrapper is unnecessary. Add kind, name fields directly to your resource type.
//
// Instead of:
//
//	Response[SingleResult[Note]]
//
// Use:
//
//	Response[Note] where Note has kind, name fields
//
// This type is kept for legacy compatibility only.
type SingleResult[T any] struct {
	Kind     string `json:"kind,omitempty"`     // Resource type (e.g., "note")
	Name     string `json:"name,omitempty"`     // Resource name (e.g., "notes/123") - AIP-122
	Resource T      `json:"resource,omitempty"` // The actual resource
}

// ErrorResponse is a standard error response following AIP-193 HTTP/1.1+JSON representation.
// https://google.aip.dev/193#http11json-representation
//
// For partial success in batch operations, this contains failed items while Data contains successes.
type ErrorResponse struct {
	Code    int           `json:"code"`              // HTTP status code (e.g., 404)
	Message string        `json:"message"`           // Human-readable error message
	Status  string        `json:"status,omitempty"`  // Canonical status name (e.g., "NOT_FOUND", "INVALID_ARGUMENT") - AIP-193
	Details []ErrorDetail `json:"details,omitempty"` // Array of detailed error information (AIP-193 requires this be called "details", not "errors")
}

// ErrorDetail provides machine-readable error information following AIP-193 ErrorInfo.
// https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto
//
// Per AIP-193: All error responses MUST include an ErrorInfo within details.
// The (reason, domain) pair forms a machine-readable identifier for the error.
type ErrorDetail struct {
	Type     string            `json:"@type"`              // Type URL, always "type.googleapis.com/google.rpc.ErrorInfo" for ErrorInfo
	Reason   string            `json:"reason"`             // UPPER_SNAKE_CASE error reason (e.g., "RESOURCE_NOT_FOUND", "FIELD_REQUIRED")
	Domain   string            `json:"domain"`             // Globally unique domain (e.g., "mind.mindweaver.com", "brain.mindweaver.com")
	Metadata map[string]string `json:"metadata,omitempty"` // Contextual information (e.g., {"resource": "tags/123", "field": "name"})
}

// ============================================================================
// Standard Method Responses (AIP-133, AIP-134, AIP-135)
// ============================================================================
//
// Per AIP guidance:
// - CREATE (AIP-133): Return the full created resource with server-generated fields
// - UPDATE (AIP-134): Return the full updated resource
// - DELETE (AIP-135): Return 204 No Content (empty body) or the deleted resource
//
// Most handlers should return the full resource type, not this minimal type.
// This type is kept for legacy compatibility or cases where full resource return is impractical.

// OperationResult is a minimal response for operations where returning the full resource is impractical.
// DEPRECATED: Prefer returning the full resource per AIP-133/134/135.
//
// For DELETE: Use c.NoContent(204) instead of returning this.
// For CREATE/UPDATE: Return Response[YourResource] with the full resource.
type OperationResult struct {
	Name string `json:"name,omitempty"` // Resource name in AIP-122 format (e.g., "notes/123")
}

// ============================================================================
// Helper Types for Common Patterns
// ============================================================================

// RequestMetadata contains metadata about the request/response cycle.
// Can be included in response types for debugging and tracing.
// Follows AIP-155 for request identification.
type RequestMetadata struct {
	RequestID string `json:"request_id,omitempty"` // From X-Request-ID header (AIP-155)
	Timestamp string `json:"timestamp,omitempty"`  // Server timestamp RFC3339 (AIP-142)
}

// ============================================================================
// Migration Notes
// ============================================================================
//
// PAGINATION MIGRATION (AIP-158 Compliance):
//
// OLD PATTERN (deprecated - offset-based):
//   type Pagination struct {
//       Page       int    `json:"page"`
//       PerPage    int    `json:"per_page"`
//       TotalPages int    `json:"total_pages"`
//   }
//
// NEW PATTERN (current - cursor-based per AIP-158):
//   type ListResult[T] struct {
//       Items         []T    `json:"items"`
//       NextPageToken string `json:"next_page_token,omitempty"`
//       TotalSize     int    `json:"total_size,omitempty"`
//   }
//
// Benefits of cursor-based pagination (AIP-158):
//   - More scalable for large datasets
//   - Handles concurrent modifications gracefully
//   - No skipped or duplicate items when data changes
//   - Works with any sorting/filtering
//
// RESPONSE STRUCTURE MIGRATION:
//
// OLD PATTERN (deprecated):
//   type Data[T any] struct {
//       Items []T `json:"items"` // Always array, even for single resources
//   }
//   Response[Note] meant: data.items[0]
//
// NEW PATTERN (current):
//   Single resource: Response[Note] means: data is the Note directly
//   Collection: Response[ListResult[Note]] means: data.items[] is array
//
// This aligns with:
//   - Google Cloud API actual behavior
//   - Google AIP-121 (resource-oriented design)
//   - Better type safety and developer experience
//
// ETAG MIGRATION (AIP-154):
//
// ETags are for conditional requests (If-Match, If-None-Match).
// They should be exchanged via HTTP headers, not in response bodies.
//
// OLD: ETag in ListResult body, SingleResult body
// NEW: ETag in HTTP header only (set via c.Response().Header().Set("ETag", value))
//
// Client flow:
//   1. GET /notes/123 → receives ETag: "abc123" in header
//   2. PUT /notes/123 with If-Match: "abc123" → server validates
//   3. If mismatch → 412 Precondition Failed (conflict detected)
//
// STANDARD METHOD RESPONSES (AIP-133, AIP-134, AIP-135):
//
// OLD: All methods returned OperationResult with Kind, ID, Done, Deleted, etc.
// NEW: Follow AIP standard method patterns:
//   - CREATE (AIP-133): Return Response[Resource] with full created resource
//   - UPDATE (AIP-134): Return Response[Resource] with full updated resource
//   - DELETE (AIP-135): Return 204 No Content (c.NoContent(204))
//   - GET (AIP-131): Return Response[Resource]
//   - LIST (AIP-132): Return Response[ListResult[Resource]]
//
// AIP ALIGNMENT SUMMARY:
//   - AIP-121: Resource-oriented design ✓
//   - AIP-122: Resource names (notes/123) ✓
//   - AIP-131-135: Standard methods (proper return types) ✓
//   - AIP-140: snake_case field names ✓
//   - AIP-142: RFC3339 timestamps ✓
//   - AIP-154: ETags in headers only ✓
//   - AIP-155: Request identification ✓
//   - AIP-158: Cursor-based pagination ✓
//   - AIP-193: Standard error format ✓
//
// ============================================================================
