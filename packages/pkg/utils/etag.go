package utils

import (
	"crypto/sha256"
	"fmt"
)

var etagSalt string

// InitETagSalt initializes the salt used for ETag hashing.
// This should be called once during application startup with the salt from config.
func InitETagSalt(salt string) {
	etagSalt = salt
}

// ComputeHashedETag generates a secure, hashed ETag from a version number.
// Uses SHA256(version + salt) to create unpredictable ETags that don't expose version numbers.
// Returns a weak ETag in the format: W/"<hash>" where hash is first 16 hex chars of SHA256.
//
// Example: version=5 -> W/"a7f3bc9d8e2f1a4c"
//
// Google AIP-154 compliant: Uses weak ETag prefix W/
func ComputeHashedETag(version int64) string {
	// Hash: sha256(version + salt)
	data := fmt.Sprintf("%d:%s", version, etagSalt)
	hash := sha256.Sum256([]byte(data))
	// Use first 16 hex chars (8 bytes) for brevity while maintaining uniqueness
	hashStr := fmt.Sprintf("%x", hash[:8])
	return fmt.Sprintf(`W/"%s"`, hashStr)
}

// ComputeListETag generates an ETag for a collection of items.
// Aggregates all item versions to create a collective ETag that changes when:
//   - Any item is added, updated, or deleted
//   - Any item's version changes
//
// Algorithm: SHA256(sum_of_versions + count + salt)
// Returns: W/"<hash>" where hash is first 16 hex chars
//
// Google AIP-154 pattern: List resources should have ETags representing collective state
func ComputeListETag(versionSum int64, count int) string {
	if count == 0 {
		return `W/"empty"`
	}

	// Build string: "sum:count:salt"
	data := fmt.Sprintf("%d:%d:%s", versionSum, count, etagSalt)

	// Hash the concatenated string
	hash := sha256.Sum256([]byte(data))
	hashStr := fmt.Sprintf("%x", hash[:8])
	return fmt.Sprintf(`W/"%s"`, hashStr)
}

// ParseETagVersion extracts the version number from a hashed ETag for debugging.
// This is NOT cryptographically secure - it's only for logging/debugging purposes.
// Returns -1 if the ETag cannot be parsed.
//
// Note: You cannot reverse a hashed ETag back to a version number.
// This function is a placeholder for future debugging needs.
func ParseETagVersion(etag string) int64 {
	// Hashed ETags cannot be reversed to version numbers
	// This is a security feature, not a bug
	return -1
}

// ComputeSimpleETag generates a simple ETag for non-versioned resources.
// Uses a hash of the count for unpredictability (better than exposing count directly).
// This is for resources that don't have version columns (collections, tags, templates, etc.)
//
// Returns: W/"<hash>" where hash is first 8 hex chars of SHA256(count + salt)
func ComputeSimpleETag(count int) string {
	if count == 0 {
		return `W/"empty"`
	}
	data := fmt.Sprintf("count:%d:%s", count, etagSalt)
	hash := sha256.Sum256([]byte(data))
	hashStr := fmt.Sprintf("%x", hash[:4]) // Use 4 bytes (8 hex chars) for simple resources
	return fmt.Sprintf(`W/"%s"`, hashStr)
}
