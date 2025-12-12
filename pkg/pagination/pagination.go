// Package pagination provides utilities for cursor-based pagination across the API.
// Page tokens are opaque strings that encode offset information (base64-encoded).
// This design allows future migration to cursor-based pagination without API changes.
package pagination

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

// DefaultPageSize is the default number of items per page.
const DefaultPageSize = 50

// MaxPageSize is the maximum allowed page size to prevent abuse.
const MaxPageSize = 100

// Request contains the pagination parameters from an API request.
type Request struct {
	PageSize  int32
	PageToken string
}

// Response contains pagination metadata for the API response.
type Response struct {
	NextPageToken string
	TotalCount    int64 // Only populated on first page (when no page_token provided)
}

// Params contains the resolved pagination parameters for database queries.
type Params struct {
	Limit  int32
	Offset int32
}

// ParseRequest parses pagination parameters from an API request.
// If pageSize is 0, DefaultPageSize is used.
// If pageSize exceeds MaxPageSize, MaxPageSize is used.
func ParseRequest(pageSize int32, pageToken string) Request {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Request{
		PageSize:  pageSize,
		PageToken: pageToken,
	}
}

// ToParams converts a pagination request to database query parameters.
// Returns the limit (pageSize + 1 for next page detection) and offset.
func (r Request) ToParams() Params {
	offset := int32(0)
	if r.PageToken != "" {
		decoded, err := DecodePageToken(r.PageToken)
		if err == nil {
			offset = decoded
		}
	}
	return Params{
		Limit:  r.PageSize + 1, // Fetch one extra to detect if there's a next page
		Offset: offset,
	}
}

// BuildResponse builds a pagination response from query results.
// Pass the actual number of items fetched (before trimming the extra item).
// If fetchedCount > pageSize, there's a next page.
func (r Request) BuildResponse(fetchedCount int, totalCount int64) Response {
	resp := Response{}

	// Only include total count on first page (no page_token)
	if r.PageToken == "" {
		resp.TotalCount = totalCount
	}

	// Check if there's a next page
	if fetchedCount > int(r.PageSize) {
		// Calculate next offset
		currentOffset := int32(0)
		if r.PageToken != "" {
			decoded, err := DecodePageToken(r.PageToken)
			if err == nil {
				currentOffset = decoded
			}
		}
		nextOffset := currentOffset + r.PageSize
		resp.NextPageToken = EncodePageToken(nextOffset)
	}

	return resp
}

// EncodePageToken encodes an offset into an opaque page token.
func EncodePageToken(offset int32) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("offset:%d", offset)))
}

// DecodePageToken decodes a page token back into an offset.
func DecodePageToken(token string) (int32, error) {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("invalid page token: %w", err)
	}

	var offset int64
	_, err = fmt.Sscanf(string(decoded), "offset:%d", &offset)
	if err != nil {
		return 0, fmt.Errorf("invalid page token format: %w", err)
	}

	return int32(offset), nil
}

// TrimResults trims the extra item fetched for next page detection.
// Returns a slice of at most pageSize items.
func TrimResults[T any](results []T, pageSize int32) []T {
	if int32(len(results)) > pageSize {
		return results[:pageSize]
	}
	return results
}

// IsFirstPage returns true if this is the first page (no page_token).
func (r Request) IsFirstPage() bool {
	return r.PageToken == ""
}

// ParseInt32 safely converts a string to int32 for pagination.
// Returns 0 if the string is empty or invalid.
func ParseInt32(s string) int32 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0
	}
	return int32(v)
}
