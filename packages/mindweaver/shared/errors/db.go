// Package errors provides utilities for detecting and handling errors across the application.
//
// This package consolidates error handling logic that is shared across services:
// - Database error detection (SQLC/SQLite errors)
// - Connect-RPC error builders (AIP-193 compliant)
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

// IsNotFoundError checks if an error indicates a record was not found.
// SQLC returns "sql: no rows in result set" for sql.ErrNoRows.
//
// Example usage in service layer:
//
//	if errors.IsNotFoundError(err) {
//		return store.Tag{}, ErrTagNotFound
//	}
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// SQLC returns this exact string for sql.ErrNoRows
	return err.Error() == "sql: no rows in result set"
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
