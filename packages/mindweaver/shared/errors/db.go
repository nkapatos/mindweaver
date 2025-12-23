// Package errors provides utilities for detecting and handling errors across the application.
//
// This package consolidates error handling logic that is shared across services:
// - Database error detection (SQLC/SQLite constraint errors)
// - Connect-RPC error builders (AIP-193 compliant)
//
// Note: For sql.ErrNoRows, use errors.Is(err, sql.ErrNoRows) directly rather than
// wrapping it in a helper. This is the idiomatic Go pattern and is more reliable.
package errors

import (
	"errors"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"
)

// IsUniqueConstraintError checks if an error is a SQLite UNIQUE constraint violation.
// SQLC returns raw driver errors which need to be unwrapped with errors.As.
//
// Example usage in service layer:
//
//	if errors.IsUniqueConstraintError(err) {
//		return 0, ErrResourceAlreadyExists
//	}
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	// Use errors.As to unwrap SQLC-returned errors
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		// SQLITE_CONSTRAINT_UNIQUE = 2067
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_UNIQUE
	}

	return false
}

// IsForeignKeyConstraintError checks if an error is a SQLite FOREIGN KEY constraint violation.
// Useful for detecting cases where deleting a resource would violate referential integrity.
//
// Example usage in service layer:
//
//	if errors.IsForeignKeyConstraintError(err) {
//		return ErrResourceInUse
//	}
func IsForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		// SQLITE_CONSTRAINT_FOREIGNKEY = 787
		return sqliteErr.Code() == sqlitelib.SQLITE_CONSTRAINT_FOREIGNKEY
	}

	return false
}
