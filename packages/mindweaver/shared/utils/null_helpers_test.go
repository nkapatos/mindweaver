package utils

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// String Conversions
// ============================================================================

func TestNullString(t *testing.T) {
	result := NullString("test")
	if !result.Valid || result.String != "test" {
		t.Errorf("NullString failed: got Valid=%v, String=%s", result.Valid, result.String)
	}
}

func TestNullStringEmpty(t *testing.T) {
	result := NullStringEmpty()
	if result.Valid {
		t.Errorf("NullStringEmpty should be invalid")
	}
}

func TestNullStringFrom(t *testing.T) {
	// Non-empty string
	result := NullStringFrom("test", true)
	if !result.Valid || result.String != "test" {
		t.Errorf("NullStringFrom with value failed")
	}

	// Empty string with nullIfEmpty=true
	result = NullStringFrom("", true)
	if result.Valid {
		t.Errorf("NullStringFrom with empty string should be invalid when nullIfEmpty=true")
	}

	// Empty string with nullIfEmpty=false
	result = NullStringFrom("", false)
	if !result.Valid {
		t.Errorf("NullStringFrom with empty string should be valid when nullIfEmpty=false")
	}
}

func TestToNullString(t *testing.T) {
	// Nil pointer
	result := ToNullString(nil)
	if result.Valid {
		t.Errorf("ToNullString with nil should be invalid")
	}

	// Valid pointer
	str := "test"
	result = ToNullString(&str)
	if !result.Valid || result.String != "test" {
		t.Errorf("ToNullString with pointer failed")
	}
}

func TestFromNullString(t *testing.T) {
	// Invalid NullString
	result := FromNullString(sql.NullString{Valid: false})
	if result != nil {
		t.Errorf("FromNullString with invalid should return nil")
	}

	// Valid NullString
	result = FromNullString(sql.NullString{String: "test", Valid: true})
	if result == nil || *result != "test" {
		t.Errorf("FromNullString with valid failed")
	}
}

// ============================================================================
// Int64 Conversions
// ============================================================================

func TestNullInt64(t *testing.T) {
	result := NullInt64(42)
	if !result.Valid || result.Int64 != 42 {
		t.Errorf("NullInt64 failed: got Valid=%v, Int64=%d", result.Valid, result.Int64)
	}
}

func TestNullInt64Empty(t *testing.T) {
	result := NullInt64Empty()
	if result.Valid {
		t.Errorf("NullInt64Empty should be invalid")
	}
}

func TestToNullInt64(t *testing.T) {
	// Nil pointer
	result := ToNullInt64(nil)
	if result.Valid {
		t.Errorf("ToNullInt64 with nil should be invalid")
	}

	// Valid pointer
	val := int64(42)
	result = ToNullInt64(&val)
	if !result.Valid || result.Int64 != 42 {
		t.Errorf("ToNullInt64 with pointer failed")
	}
}

func TestFromNullInt64(t *testing.T) {
	// Invalid NullInt64
	result := FromNullInt64(sql.NullInt64{Valid: false})
	if result != nil {
		t.Errorf("FromNullInt64 with invalid should return nil")
	}

	// Valid NullInt64
	result = FromNullInt64(sql.NullInt64{Int64: 42, Valid: true})
	if result == nil || *result != 42 {
		t.Errorf("FromNullInt64 with valid failed")
	}
}

// ============================================================================
// Float64 Conversions
// ============================================================================

func TestNullFloat64(t *testing.T) {
	result := NullFloat64(3.14)
	if !result.Valid || result.Float64 != 3.14 {
		t.Errorf("NullFloat64 failed")
	}
}

func TestNullFloat64Empty(t *testing.T) {
	result := NullFloat64Empty()
	if result.Valid {
		t.Errorf("NullFloat64Empty should be invalid")
	}
}

func TestToNullFloat64(t *testing.T) {
	// Nil pointer
	result := ToNullFloat64(nil)
	if result.Valid {
		t.Errorf("ToNullFloat64 with nil should be invalid")
	}

	// Valid pointer
	val := 3.14
	result = ToNullFloat64(&val)
	if !result.Valid || result.Float64 != 3.14 {
		t.Errorf("ToNullFloat64 with pointer failed")
	}
}

func TestFromNullFloat64(t *testing.T) {
	// Invalid NullFloat64
	result := FromNullFloat64(sql.NullFloat64{Valid: false})
	if result != nil {
		t.Errorf("FromNullFloat64 with invalid should return nil")
	}

	// Valid NullFloat64
	result = FromNullFloat64(sql.NullFloat64{Float64: 3.14, Valid: true})
	if result == nil || *result != 3.14 {
		t.Errorf("FromNullFloat64 with valid failed")
	}
}

// ============================================================================
// Bool Conversions
// ============================================================================

func TestNullBool(t *testing.T) {
	result := NullBool(true)
	if !result.Valid || !result.Bool {
		t.Errorf("NullBool failed")
	}
}

func TestToNullBool(t *testing.T) {
	// Nil pointer
	result := ToNullBool(nil)
	if result.Valid {
		t.Errorf("ToNullBool with nil should be invalid")
	}

	// Valid pointer
	val := true
	result = ToNullBool(&val)
	if !result.Valid || !result.Bool {
		t.Errorf("ToNullBool with pointer failed")
	}
}

func TestFromNullBool(t *testing.T) {
	// Invalid NullBool
	result := FromNullBool(sql.NullBool{Valid: false})
	if result != nil {
		t.Errorf("FromNullBool with invalid should return nil")
	}

	// Valid NullBool
	result = FromNullBool(sql.NullBool{Bool: true, Valid: true})
	if result == nil || !*result {
		t.Errorf("FromNullBool with valid failed")
	}
}

func TestBoolFromNull(t *testing.T) {
	// Invalid NullBool returns false
	result := BoolFromNull(sql.NullBool{Valid: false})
	if result {
		t.Errorf("BoolFromNull with invalid should return false")
	}

	// Valid NullBool returns actual value
	result = BoolFromNull(sql.NullBool{Bool: true, Valid: true})
	if !result {
		t.Errorf("BoolFromNull with valid true failed")
	}

	result = BoolFromNull(sql.NullBool{Bool: false, Valid: true})
	if result {
		t.Errorf("BoolFromNull with valid false failed")
	}
}

// ============================================================================
// Time Formatting
// ============================================================================

func TestFormatNullTime(t *testing.T) {
	// Invalid NullTime
	result := FormatNullTime(sql.NullTime{Valid: false})
	if result != nil {
		t.Errorf("FormatNullTime with invalid should return nil")
	}

	// Valid NullTime
	tm := time.Date(2023, 12, 1, 10, 30, 0, 0, time.UTC)
	result = FormatNullTime(sql.NullTime{Time: tm, Valid: true})
	if result == nil || *result != "2023-12-01T10:30:00Z" {
		t.Errorf("FormatNullTime failed: got %v", result)
	}
}

func TestFormatNullTimeOrDefault(t *testing.T) {
	// Invalid NullTime
	result := FormatNullTimeOrDefault(sql.NullTime{Valid: false})
	if result != "" {
		t.Errorf("FormatNullTimeOrDefault with invalid should return empty string")
	}

	// Valid NullTime
	tm := time.Date(2023, 12, 1, 10, 30, 0, 0, time.UTC)
	result = FormatNullTimeOrDefault(sql.NullTime{Time: tm, Valid: true})
	if result != "2023-12-01T10:30:00Z" {
		t.Errorf("FormatNullTimeOrDefault failed: got %s", result)
	}
}

// ============================================================================
// Default Value Helpers
// ============================================================================

func TestStringOrDefault(t *testing.T) {
	// Nil pointer
	result := StringOrDefault(nil, "default")
	if result != "default" {
		t.Errorf("StringOrDefault with nil should return default")
	}

	// Valid pointer
	str := "value"
	result = StringOrDefault(&str, "default")
	if result != "value" {
		t.Errorf("StringOrDefault with pointer should return value")
	}
}

func TestInt64OrDefault(t *testing.T) {
	// Nil pointer
	result := Int64OrDefault(nil, 99)
	if result != 99 {
		t.Errorf("Int64OrDefault with nil should return default")
	}

	// Valid pointer
	val := int64(42)
	result = Int64OrDefault(&val, 99)
	if result != 42 {
		t.Errorf("Int64OrDefault with pointer should return value")
	}
}

func TestBoolOrDefault(t *testing.T) {
	// Nil pointer
	result := BoolOrDefault(nil, true)
	if !result {
		t.Errorf("BoolOrDefault with nil should return default")
	}

	// Valid pointer
	val := false
	result = BoolOrDefault(&val, true)
	if result {
		t.Errorf("BoolOrDefault with pointer should return value")
	}
}

func TestJSONOrDefault(t *testing.T) {
	defaultJSON := json.RawMessage(`{"default":true}`)

	// Nil JSON
	result := JSONOrDefault(nil, defaultJSON)
	if string(result) != string(defaultJSON) {
		t.Errorf("JSONOrDefault with nil should return default")
	}

	// Empty JSON
	result = JSONOrDefault(json.RawMessage{}, defaultJSON)
	if string(result) != string(defaultJSON) {
		t.Errorf("JSONOrDefault with empty should return default")
	}

	// Valid JSON
	validJSON := json.RawMessage(`{"valid":true}`)
	result = JSONOrDefault(validJSON, defaultJSON)
	if string(result) != string(validJSON) {
		t.Errorf("JSONOrDefault with value should return value")
	}
}

func TestJSONPtrOrDefault(t *testing.T) {
	defaultJSON := json.RawMessage(`{"default":true}`)

	// Nil pointer
	result := JSONPtrOrDefault(nil, defaultJSON)
	if string(result) != string(defaultJSON) {
		t.Errorf("JSONPtrOrDefault with nil should return default")
	}

	// Empty JSON
	empty := json.RawMessage{}
	result = JSONPtrOrDefault(&empty, defaultJSON)
	if string(result) != string(defaultJSON) {
		t.Errorf("JSONPtrOrDefault with empty should return default")
	}

	// Valid JSON
	validJSON := json.RawMessage(`{"valid":true}`)
	result = JSONPtrOrDefault(&validJSON, defaultJSON)
	if string(result) != string(validJSON) {
		t.Errorf("JSONPtrOrDefault with value should return value")
	}
}

// ============================================================================
// Merge Helpers
// ============================================================================

func TestMergeNullString(t *testing.T) {
	existing := sql.NullString{String: "old", Valid: true}

	// Nil new value keeps existing
	result := MergeNullString(nil, existing)
	if !result.Valid || result.String != "old" {
		t.Errorf("MergeNullString with nil should keep existing")
	}

	// New value replaces existing
	newVal := "new"
	result = MergeNullString(&newVal, existing)
	if !result.Valid || result.String != "new" {
		t.Errorf("MergeNullString with new value should replace")
	}
}

func TestMergeNullInt64(t *testing.T) {
	existing := sql.NullInt64{Int64: 10, Valid: true}

	// Nil new value keeps existing
	result := MergeNullInt64(nil, existing)
	if !result.Valid || result.Int64 != 10 {
		t.Errorf("MergeNullInt64 with nil should keep existing")
	}

	// New value replaces existing
	newVal := int64(20)
	result = MergeNullInt64(&newVal, existing)
	if !result.Valid || result.Int64 != 20 {
		t.Errorf("MergeNullInt64 with new value should replace")
	}
}

func TestMergeNullBool(t *testing.T) {
	existing := sql.NullBool{Bool: false, Valid: true}

	// Nil new value keeps existing
	result := MergeNullBool(nil, existing)
	if !result.Valid || result.Bool {
		t.Errorf("MergeNullBool with nil should keep existing")
	}

	// New value replaces existing
	newVal := true
	result = MergeNullBool(&newVal, existing)
	if !result.Valid || !result.Bool {
		t.Errorf("MergeNullBool with new value should replace")
	}
}

// ============================================================================
// String Helpers
// ============================================================================

func TestNilIfEmpty(t *testing.T) {
	// Empty string
	result := NilIfEmpty("")
	if result != nil {
		t.Errorf("NilIfEmpty with empty string should return nil")
	}

	// Non-empty string
	result = NilIfEmpty("value")
	if result == nil || *result != "value" {
		t.Errorf("NilIfEmpty with value should return pointer")
	}
}

// ============================================================================
// Parsing Helpers
// ============================================================================

func TestParseUUID(t *testing.T) {
	// Valid UUID
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	result := ParseUUID(validUUID)
	if result == uuid.Nil {
		t.Errorf("ParseUUID with valid UUID should not return Nil")
	}

	// Invalid UUID
	result = ParseUUID("invalid-uuid")
	if result != uuid.Nil {
		t.Errorf("ParseUUID with invalid UUID should return Nil")
	}
}

func TestFromInterface(t *testing.T) {
	// Nil interface
	result := FromInterface(nil)
	if result != nil {
		t.Errorf("FromInterface with nil should return nil")
	}

	// Valid int64
	var val interface{} = int64(42)
	result = FromInterface(val)
	if result == nil || *result != 42 {
		t.Errorf("FromInterface with int64 failed")
	}

	// Non-int64 type
	result = FromInterface("not an int")
	if result != nil {
		t.Errorf("FromInterface with non-int64 should return nil")
	}
}
