package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// ParseIDParam Unit Tests
// ============================================================================
// ParseIDParam wraps strconv.ParseInt(s, 10, 64).
// These tests document its behavior for handler authors.

func TestParseIDParam(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		// Valid cases
		{name: "valid positive", input: "1", want: 1, wantErr: false},
		{name: "valid large", input: "9223372036854775807", want: 9223372036854775807, wantErr: false},
		{name: "valid zero", input: "0", want: 0, wantErr: false},
		{name: "valid negative", input: "-1", want: -1, wantErr: false},
		{name: "valid negative large", input: "-9223372036854775808", want: -9223372036854775808, wantErr: false},

		// Invalid cases - parse errors
		{name: "empty string", input: "", wantErr: true},
		{name: "non-numeric", input: "abc", wantErr: true},
		{name: "mixed alphanumeric", input: "123abc", wantErr: true},
		{name: "float", input: "1.5", wantErr: true},
		{name: "whitespace", input: " 1", wantErr: true},
		{name: "overflow positive", input: "9223372036854775808", wantErr: true},
		{name: "overflow negative", input: "-9223372036854775809", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIDParam(tt.input)

			if tt.wantErr {
				require.Error(t, err, "ParseIDParam(%q) should return error", tt.input)
				return
			}

			require.NoError(t, err, "ParseIDParam(%q) should not return error", tt.input)
			assert.Equal(t, tt.want, got, "ParseIDParam(%q) = %d, want %d", tt.input, got, tt.want)
		})
	}
}

// TestParseIDParam_DocumentsBehavior documents that ParseIDParam does NOT
// validate business rules (positive IDs). It only parses strings to int64.
// Handlers must add their own validation if they need positive-only IDs.
func TestParseIDParam_DocumentsBehavior(t *testing.T) {
	// Zero parses successfully - handlers must validate if zero is invalid
	id, err := ParseIDParam("0")
	require.NoError(t, err)
	assert.Equal(t, int64(0), id)

	// Negative parses successfully - handlers must validate if negative is invalid
	id, err = ParseIDParam("-1")
	require.NoError(t, err)
	assert.Equal(t, int64(-1), id)
}
