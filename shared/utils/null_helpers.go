// Package utils provides utilities for working with SQL null types and common operations.
//
// This package consolidates null type conversion helpers that are shared across
// the mind and brain services, following consistent naming conventions and patterns.
package utils

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
)

// ============================================================================
// Null Type Conversions - Convert between SQL null types and Go pointers
// ============================================================================

// ToNullString converts a string pointer to sql.NullString.
func ToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullString creates a sql.NullString from a string value.
// Use this for literal values: utils.NullString("value")
func NullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// NullStringEmpty creates an empty/invalid sql.NullString.
// Use this instead of sql.NullString{Valid: false}
func NullStringEmpty() sql.NullString {
	return sql.NullString{Valid: false}
}

// NullStringFrom creates a sql.NullString from an optional string value.
// Handles empty strings as NULL.
func NullStringFrom(s string, nullIfEmpty bool) sql.NullString {
	if nullIfEmpty && s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// FromNullString converts sql.NullString to a string pointer.
func FromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// ToNullInt64 converts an int64 pointer to sql.NullInt64.
func ToNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// NullInt64 creates a sql.NullInt64 from an int64 value.
// Use this for literal values: utils.NullInt64(123)
func NullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}

// NullInt64Empty creates an empty/invalid sql.NullInt64.
// Use this instead of sql.NullInt64{Valid: false}
func NullInt64Empty() sql.NullInt64 {
	return sql.NullInt64{Valid: false}
}

// FromNullInt64 converts sql.NullInt64 to an int64 pointer.
func FromNullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

// ToNullFloat64 converts a float64 pointer to sql.NullFloat64.
func ToNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

// NullFloat64 creates a sql.NullFloat64 from a float64 value.
// Use this for literal values: utils.NullFloat64(3.14)
func NullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}

// NullFloat64Empty creates an empty/invalid sql.NullFloat64.
// Use this instead of sql.NullFloat64{Valid: false}
func NullFloat64Empty() sql.NullFloat64 {
	return sql.NullFloat64{Valid: false}
}

// FromNullFloat64 converts sql.NullFloat64 to a float64 pointer.
func FromNullFloat64(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

// ToNullBool converts a bool pointer to sql.NullBool.
func ToNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

// NullBool creates a sql.NullBool from a bool value.
// Use this for literal values: utils.NullBool(true)
func NullBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

// FromNullBool converts sql.NullBool to a bool pointer.
func FromNullBool(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}

// BoolFromNull converts sql.NullBool to bool (returns false if null).
func BoolFromNull(nb sql.NullBool) bool {
	if !nb.Valid {
		return false
	}
	return nb.Bool
}

// ============================================================================
// Time Formatting
// ============================================================================

// FormatNullTime converts sql.NullTime to RFC3339 formatted string pointer.
func FormatNullTime(t sql.NullTime) *string {
	if !t.Valid {
		return nil
	}
	formatted := t.Time.Format("2006-01-02T15:04:05Z")
	return &formatted
}

// FormatNullTimeOrDefault converts sql.NullTime to RFC3339 string, returns empty string if null.
func FormatNullTimeOrDefault(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02T15:04:05Z")
}

// ============================================================================
// Default Value Helpers
// ============================================================================

// StringOrDefault returns the string value or a default if pointer is nil.
func StringOrDefault(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// Int64OrDefault returns the int64 value or a default if pointer is nil.
func Int64OrDefault(i *int64, def int64) int64 {
	if i == nil {
		return def
	}
	return *i
}

// BoolOrDefault returns the bool value or a default if pointer is nil.
func BoolOrDefault(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// JSONOrDefault returns the JSON value or a default if nil or empty.
func JSONOrDefault(j json.RawMessage, def json.RawMessage) json.RawMessage {
	if len(j) == 0 {
		return def
	}
	return j
}

// JSONPtrOrDefault returns the JSON value or a default if pointer is nil or empty.
func JSONPtrOrDefault(j *json.RawMessage, def json.RawMessage) json.RawMessage {
	if j == nil || len(*j) == 0 {
		return def
	}
	return *j
}

// ============================================================================
// Merge Helpers - Used for PATCH/UPDATE operations
// ============================================================================

// MergeNullString returns new value if non-nil, otherwise returns existing value.
func MergeNullString(newVal *string, existing sql.NullString) sql.NullString {
	if newVal != nil {
		return sql.NullString{String: *newVal, Valid: true}
	}
	return existing
}

// MergeNullInt64 returns new value if non-nil, otherwise returns existing value.
func MergeNullInt64(newVal *int64, existing sql.NullInt64) sql.NullInt64 {
	if newVal != nil {
		return sql.NullInt64{Int64: *newVal, Valid: true}
	}
	return existing
}

// MergeNullBool returns new value if non-nil, otherwise returns existing value.
func MergeNullBool(newVal *bool, existing sql.NullBool) sql.NullBool {
	if newVal != nil {
		return sql.NullBool{Bool: *newVal, Valid: true}
	}
	return existing
}

// ============================================================================
// String Helpers
// ============================================================================

// NilIfEmpty returns nil if the string is empty, otherwise a pointer to the string.
func NilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ============================================================================
// Parsing Helpers - Used in API layers for param parsing
// ============================================================================

// ParseIDParam parses a string ID parameter to int64.
func ParseIDParam(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// ParseUUID parses a string to uuid.UUID, returns uuid.Nil if invalid.
func ParseUUID(s string) uuid.UUID {
	u, _ := uuid.Parse(s)
	return u
}

// FromInterface converts interface{} (from SQLite nullable fields in CTEs) to *int64.
// SQLite driver returns interface{} for nullable integers in some query results.
func FromInterface(v any) *int64 {
	if v == nil {
		return nil
	}
	// SQLite returns int64 for INTEGER columns
	if i64, ok := v.(int64); ok {
		return &i64
	}
	return nil
}
