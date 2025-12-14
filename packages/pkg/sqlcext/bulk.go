// Package sqlcext provides extensions to sqlc for features it cannot generate,
// particularly bulk operations and SQLite-specific functionality.
package sqlcext

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const (
	// DefaultBatchSize is conservative for older SQLite versions.
	// SQLite has a default limit of 999 variables (SQLITE_MAX_VARIABLE_NUMBER).
	// For a 3-column insert, this allows ~333 rows per batch (333 * 3 = 999).
	DefaultBatchSize = 100

	// MaxBatchSize for modern SQLite 3.32.0+ which supports up to 32766 variables.
	// Still conservative to account for other query overhead.
	MaxBatchSize = 1000
)

// DBTX is a minimal interface for database operations that matches sqlc's generated interface.
// This allows BulkInserter to work with *sql.DB, *sql.Tx, or sqlc's *Queries types.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

// BulkInserter handles multi-value INSERT statements for SQLite.
// It automatically chunks large batches to stay within SQLite's variable limits.
//
// Example usage:
//
//	inserter := sqlcext.NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 100)
//	rows := [][]any{
//	    {1, "author", "John Doe"},
//	    {1, "tags", "golang,sqlite"},
//	}
//	err := inserter.Insert(ctx, db, rows)
type BulkInserter struct {
	table      string
	columns    []string
	batchSize  int
	valueCount int // Number of values per row (len(columns))
}

// NewBulkInserter creates a bulk inserter for the given table and columns.
//
// Parameters:
//   - table: The table name (e.g., "note_meta")
//   - columns: Column names in order (e.g., []string{"note_id", "key", "value"})
//   - batchSize: Maximum rows per INSERT statement (0 uses DefaultBatchSize)
//
// The batchSize should be chosen such that: batchSize * len(columns) < SQLITE_MAX_VARIABLE_NUMBER
// For safety, DefaultBatchSize (100) works well for most cases.
func NewBulkInserter(table string, columns []string, batchSize int) *BulkInserter {
	if batchSize <= 0 || batchSize > MaxBatchSize {
		batchSize = DefaultBatchSize
	}

	return &BulkInserter{
		table:      table,
		columns:    columns,
		batchSize:  batchSize,
		valueCount: len(columns),
	}
}

// Insert executes a bulk INSERT for the given rows.
// Each row must have exactly len(columns) values, in the same order as the columns.
//
// If len(rows) > batchSize, it automatically chunks into multiple INSERT statements.
// All INSERTs are executed within the same transaction context (caller controls transaction).
//
// Returns an error if:
//   - rows is nil (empty slices are OK and return nil)
//   - any row has incorrect number of values
//   - database execution fails
//
// Example:
//
//	rows := [][]any{
//	    {1, "key1", "value1"},
//	    {1, "key2", "value2"},
//	    {2, "key1", "value3"},
//	}
//	inserter := NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 100)
//	err := inserter.Insert(ctx, tx, rows)
func (b *BulkInserter) Insert(ctx context.Context, db DBTX, rows [][]any) error {
	if len(rows) == 0 {
		return nil // Nothing to insert
	}

	// Process in chunks to stay within SQLite variable limits
	for i := 0; i < len(rows); i += b.batchSize {
		end := i + b.batchSize
		if end > len(rows) {
			end = len(rows)
		}

		chunk := rows[i:end]
		if err := b.insertChunk(ctx, db, chunk); err != nil {
			return fmt.Errorf("bulk insert chunk [%d:%d]: %w", i, end, err)
		}
	}

	return nil
}

// insertChunk executes a single multi-value INSERT statement for a chunk of rows.
func (b *BulkInserter) insertChunk(ctx context.Context, db DBTX, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}

	// Validate all rows have correct number of values
	for i, row := range rows {
		if len(row) != b.valueCount {
			return fmt.Errorf("row %d has %d values, expected %d", i, len(row), b.valueCount)
		}
	}

	// Build SQL: INSERT INTO table (col1, col2, ...) VALUES (?, ?, ...), (?, ?, ...), ...
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(b.columns, ", "))
	sb.WriteString(") VALUES ")

	// Build placeholders: (?, ?, ?), (?, ?, ?), ...
	rowPlaceholder := "(" + strings.Repeat("?, ", b.valueCount-1) + "?)"
	valueClauses := make([]string, len(rows))
	for i := range rows {
		valueClauses[i] = rowPlaceholder
	}
	sb.WriteString(strings.Join(valueClauses, ", "))

	// Flatten rows into args slice
	args := make([]any, 0, len(rows)*b.valueCount)
	for _, row := range rows {
		args = append(args, row...)
	}

	// Execute the bulk INSERT
	_, err := db.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("exec bulk insert: %w", err)
	}

	return nil
}

// BulkUpsert executes a bulk INSERT with ON CONFLICT DO UPDATE for SQLite.
// This is useful for UPSERT operations where you want to insert or update based on a conflict.
//
// Parameters:
//   - conflictColumns: Columns that define the conflict (e.g., []string{"note_id", "key"})
//   - updateColumns: Columns to update on conflict (e.g., []string{"value"})
//
// Example:
//
//	inserter := sqlcext.NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 100)
//	rows := [][]any{
//	    {1, "author", "John Doe"},
//	    {1, "tags", "golang,sqlite"},
//	}
//	err := inserter.Upsert(ctx, db, rows, []string{"note_id", "key"}, []string{"value"})
func (b *BulkInserter) Upsert(ctx context.Context, db DBTX, rows [][]any, conflictColumns []string, updateColumns []string) error {
	if len(rows) == 0 {
		return nil
	}

	// Process in chunks
	for i := 0; i < len(rows); i += b.batchSize {
		end := i + b.batchSize
		if end > len(rows) {
			end = len(rows)
		}

		chunk := rows[i:end]
		if err := b.upsertChunk(ctx, db, chunk, conflictColumns, updateColumns); err != nil {
			return fmt.Errorf("bulk upsert chunk [%d:%d]: %w", i, end, err)
		}
	}

	return nil
}

func (b *BulkInserter) upsertChunk(ctx context.Context, db DBTX, rows [][]any, conflictColumns []string, updateColumns []string) error {
	if len(rows) == 0 {
		return nil
	}

	// Validate all rows have correct number of values
	for i, row := range rows {
		if len(row) != b.valueCount {
			return fmt.Errorf("row %d has %d values, expected %d", i, len(row), b.valueCount)
		}
	}

	// Build SQL: INSERT INTO table (col1, col2, ...) VALUES (?, ?), (?, ?) ON CONFLICT(...) DO UPDATE SET ...
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(b.table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(b.columns, ", "))
	sb.WriteString(") VALUES ")

	// Build placeholders
	rowPlaceholder := "(" + strings.Repeat("?, ", b.valueCount-1) + "?)"
	valueClauses := make([]string, len(rows))
	for i := range rows {
		valueClauses[i] = rowPlaceholder
	}
	sb.WriteString(strings.Join(valueClauses, ", "))

	// Add ON CONFLICT clause
	sb.WriteString(" ON CONFLICT (")
	sb.WriteString(strings.Join(conflictColumns, ", "))
	sb.WriteString(") DO UPDATE SET ")

	// Build update assignments: col = excluded.col
	updateAssignments := make([]string, len(updateColumns))
	for i, col := range updateColumns {
		updateAssignments[i] = fmt.Sprintf("%s = excluded.%s", col, col)
	}
	sb.WriteString(strings.Join(updateAssignments, ", "))

	// Flatten rows into args
	args := make([]any, 0, len(rows)*b.valueCount)
	for _, row := range rows {
		args = append(args, row...)
	}

	// Execute
	_, err := db.ExecContext(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("exec bulk upsert: %w", err)
	}

	return nil
}
