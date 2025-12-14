// Package sqlcext provides extensions to sqlc for features it cannot generate,
// particularly FTS5 virtual table queries and other SQLite-specific functionality.
//
// SECURITY: All queries use parameterized statements with ? placeholders.
// User input is NEVER concatenated into SQL strings.
package sqlcext

import (
	"database/sql"
	"time"
)

// FTSConfig defines the table configuration for FTS5 queries.
// This allows the same querier to work with different table schemas.
type FTSConfig struct {
	// ContentTable is the main content table (e.g., "notes", "assistant_notes")
	ContentTable string
	// FTSTable is the FTS5 virtual table (e.g., "notes_fts", "assistant_notes_fts")
	FTSTable string
	// IDColumn is the primary key column name (usually "id")
	IDColumn string
	// ContentRowID is the column that links FTS to content (usually "id")
	ContentRowID string
}

// FTSSearchParams contains the parameters for an FTS search query.
type FTSSearchParams struct {
	Query       string `json:"query"`        // Search query text (will be sanitized)
	LimitCount  int64  `json:"limit_count"`  // Maximum results to return
	OffsetCount int64  `json:"offset_count"` // Pagination offset
}

// FTSResult is a generic interface that FTS result types must implement.
// This allows the querier to work with different result structures.
type FTSResult interface {
	GetID() int64
	GetTitle() string
	GetScore() float64
	GetCreatedAt() time.Time
}

type CollectionTreeRow struct {
	ID          int64
	Name        string
	ParentID    sql.NullInt64
	Path        string
	Description sql.NullString
	Position    sql.NullInt64
	IsSystem    bool
	Depth       int
}
