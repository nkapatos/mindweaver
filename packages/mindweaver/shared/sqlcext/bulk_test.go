package sqlcext

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// setupBulkTestDB creates an in-memory SQLite database for bulk insert testing
func setupBulkTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create test table
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value TEXT NOT NULL,
			count INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	return db
}

// setupMetaTestDB creates a test database for note_meta simulation
func setupMetaTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE note_meta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			note_id INTEGER NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			UNIQUE(note_id, key)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create note_meta table: %v", err)
	}

	return db
}

func TestNewBulkInserter(t *testing.T) {
	tests := []struct {
		name          string
		table         string
		columns       []string
		batchSize     int
		wantBatchSize int
	}{
		{
			name:          "default batch size (zero)",
			table:         "test_table",
			columns:       []string{"name", "value", "count"},
			batchSize:     0,
			wantBatchSize: DefaultBatchSize,
		},
		{
			name:          "default batch size (negative)",
			table:         "test_table",
			columns:       []string{"name", "value"},
			batchSize:     -1,
			wantBatchSize: DefaultBatchSize,
		},
		{
			name:          "custom batch size",
			table:         "test_table",
			columns:       []string{"name", "value"},
			batchSize:     50,
			wantBatchSize: 50,
		},
		{
			name:          "exceeds max batch size",
			table:         "test_table",
			columns:       []string{"name"},
			batchSize:     2000,
			wantBatchSize: DefaultBatchSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inserter := NewBulkInserter(tt.table, tt.columns, tt.batchSize)

			if inserter.table != tt.table {
				t.Errorf("table = %q, want %q", inserter.table, tt.table)
			}
			if len(inserter.columns) != len(tt.columns) {
				t.Errorf("len(columns) = %d, want %d", len(inserter.columns), len(tt.columns))
			}
			if inserter.batchSize != tt.wantBatchSize {
				t.Errorf("batchSize = %d, want %d", inserter.batchSize, tt.wantBatchSize)
			}
			if inserter.valueCount != len(tt.columns) {
				t.Errorf("valueCount = %d, want %d", inserter.valueCount, len(tt.columns))
			}
		})
	}
}

func TestBulkInserter_Insert_EmptyRows(t *testing.T) {
	db := setupBulkTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)

	// Empty slice should not error
	err := inserter.Insert(context.Background(), db, nil)
	if err != nil {
		t.Errorf("Insert(nil) returned error: %v", err)
	}

	err = inserter.Insert(context.Background(), db, [][]any{})
	if err != nil {
		t.Errorf("Insert(empty) returned error: %v", err)
	}

	// Verify no rows inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

func TestBulkInserter_Insert_SingleRow(t *testing.T) {
	db := setupBulkTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)

	rows := [][]any{
		{"test1", "value1", 42},
	}

	err := inserter.Insert(context.Background(), db, rows)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify inserted
	var name, value string
	var count int
	err = db.QueryRow("SELECT name, value, count FROM test_table").Scan(&name, &value, &count)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if name != "test1" || value != "value1" || count != 42 {
		t.Errorf("got (%q, %q, %d), want (test1, value1, 42)", name, value, count)
	}
}

func TestBulkInserter_Insert_MultipleRows(t *testing.T) {
	db := setupBulkTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)

	rows := [][]any{
		{"test1", "value1", 1},
		{"test2", "value2", 2},
		{"test3", "value3", 3},
		{"test4", "value4", 4},
		{"test5", "value5", 5},
	}

	err := inserter.Insert(context.Background(), db, rows)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify count
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 rows, got %d", count)
	}

	// Verify data
	rows_result, err := db.Query("SELECT name, value, count FROM test_table ORDER BY count")
	if err != nil {
		t.Fatalf("failed to query rows: %v", err)
	}
	defer rows_result.Close()

	i := 0
	for rows_result.Next() {
		var name, value string
		var c int
		err := rows_result.Scan(&name, &value, &c)
		if err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		wantName := fmt.Sprintf("test%d", i+1)
		wantValue := fmt.Sprintf("value%d", i+1)
		if name != wantName || value != wantValue || c != i+1 {
			t.Errorf("row %d: got (%q, %q, %d), want (%q, %q, %d)", i, name, value, c, wantName, wantValue, i+1)
		}
		i++
	}
}

func TestBulkInserter_Insert_Chunking(t *testing.T) {
	db := setupBulkTestDB(t)
	defer db.Close()

	// Use small batch size to test chunking
	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 3)

	// Insert 10 rows with batch size 3 (should chunk into 4 batches: 3+3+3+1)
	rows := make([][]any, 10)
	for i := 0; i < 10; i++ {
		rows[i] = []any{fmt.Sprintf("test%d", i), fmt.Sprintf("value%d", i), i}
	}

	err := inserter.Insert(context.Background(), db, rows)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify all 10 rows inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 10 {
		t.Errorf("expected 10 rows, got %d", count)
	}
}

func TestBulkInserter_Insert_InvalidRowLength(t *testing.T) {
	db := setupBulkTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)

	// Row with wrong number of values
	rows := [][]any{
		{"test1", "value1", 1},
		{"test2", "value2"}, // Missing count
	}

	err := inserter.Insert(context.Background(), db, rows)
	if err == nil {
		t.Fatal("expected error for invalid row length, got nil")
	}

	if !strings.Contains(err.Error(), "has 2 values, expected 3") {
		t.Errorf("error message should mention incorrect values, got: %v", err)
	}
}

func TestBulkInserter_Insert_Transaction(t *testing.T) {
	db := setupBulkTestDB(t)
	defer db.Close()

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)

	rows := [][]any{
		{"test1", "value1", 1},
		{"test2", "value2", 2},
	}

	err = inserter.Insert(context.Background(), tx, rows)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Verify within transaction
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows in transaction, got %d", count)
	}

	// Rollback
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback: %v", err)
	}

	// Verify rollback worked
	err = db.QueryRow("SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count after rollback: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

func TestBulkInserter_Upsert_Insert(t *testing.T) {
	db := setupMetaTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 100)

	rows := [][]any{
		{1, "author", "John Doe"},
		{1, "tags", "golang,sqlite"},
		{2, "author", "Jane Smith"},
	}

	err := inserter.Upsert(context.Background(), db, rows, []string{"note_id", "key"}, []string{"value"})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Verify inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM note_meta").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

func TestBulkInserter_Upsert_Update(t *testing.T) {
	db := setupMetaTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 100)

	// Insert initial data
	rows1 := [][]any{
		{1, "author", "John Doe"},
		{1, "tags", "golang"},
	}
	err := inserter.Upsert(context.Background(), db, rows1, []string{"note_id", "key"}, []string{"value"})
	if err != nil {
		t.Fatalf("initial Upsert failed: %v", err)
	}

	// Update existing rows
	rows2 := [][]any{
		{1, "author", "John Smith"},  // Update author
		{1, "tags", "golang,sqlite"}, // Update tags
		{1, "created", "2025-01-01"}, // Insert new key
	}
	err = inserter.Upsert(context.Background(), db, rows2, []string{"note_id", "key"}, []string{"value"})
	if err != nil {
		t.Fatalf("second Upsert failed: %v", err)
	}

	// Verify count (should be 3 total)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM note_meta").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}

	// Verify updated values
	var author string
	err = db.QueryRow("SELECT value FROM note_meta WHERE note_id = 1 AND key = 'author'").Scan(&author)
	if err != nil {
		t.Fatalf("failed to query author: %v", err)
	}
	if author != "John Smith" {
		t.Errorf("author = %q, want %q", author, "John Smith")
	}

	var tags string
	err = db.QueryRow("SELECT value FROM note_meta WHERE note_id = 1 AND key = 'tags'").Scan(&tags)
	if err != nil {
		t.Fatalf("failed to query tags: %v", err)
	}
	if tags != "golang,sqlite" {
		t.Errorf("tags = %q, want %q", tags, "golang,sqlite")
	}
}

func TestBulkInserter_Upsert_Chunking(t *testing.T) {
	db := setupMetaTestDB(t)
	defer db.Close()

	// Small batch size for chunking test
	inserter := NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 2)

	rows := [][]any{
		{1, "key1", "value1"},
		{1, "key2", "value2"},
		{1, "key3", "value3"},
		{1, "key4", "value4"},
		{1, "key5", "value5"},
	}

	err := inserter.Upsert(context.Background(), db, rows, []string{"note_id", "key"}, []string{"value"})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Verify all inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM note_meta").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query count: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 rows, got %d", count)
	}
}

func TestBulkInserter_Upsert_EmptyRows(t *testing.T) {
	db := setupMetaTestDB(t)
	defer db.Close()

	inserter := NewBulkInserter("note_meta", []string{"note_id", "key", "value"}, 100)

	// Empty slice should not error
	err := inserter.Upsert(context.Background(), db, nil, []string{"note_id", "key"}, []string{"value"})
	if err != nil {
		t.Errorf("Upsert(nil) returned error: %v", err)
	}

	err = inserter.Upsert(context.Background(), db, [][]any{}, []string{"note_id", "key"}, []string{"value"})
	if err != nil {
		t.Errorf("Upsert(empty) returned error: %v", err)
	}
}

// Benchmark tests
func BenchmarkBulkInserter_Insert_Small(b *testing.B) {
	db := setupBulkTestDB(&testing.T{})
	defer db.Close()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)
	rows := [][]any{
		{"test1", "value1", 1},
		{"test2", "value2", 2},
		{"test3", "value3", 3},
		{"test4", "value4", 4},
		{"test5", "value5", 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear table
		db.Exec("DELETE FROM test_table")

		err := inserter.Insert(context.Background(), db, rows)
		if err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

func BenchmarkBulkInserter_Insert_Large(b *testing.B) {
	db := setupBulkTestDB(&testing.T{})
	defer db.Close()

	inserter := NewBulkInserter("test_table", []string{"name", "value", "count"}, 100)

	// 1000 rows
	rows := make([][]any, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = []any{fmt.Sprintf("test%d", i), fmt.Sprintf("value%d", i), i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Exec("DELETE FROM test_table")

		err := inserter.Insert(context.Background(), db, rows)
		if err != nil {
			b.Fatalf("Insert failed: %v", err)
		}
	}
}

func BenchmarkLoopInsert_Small(b *testing.B) {
	db := setupBulkTestDB(&testing.T{})
	defer db.Close()

	rows := [][]any{
		{"test1", "value1", 1},
		{"test2", "value2", 2},
		{"test3", "value3", 3},
		{"test4", "value4", 4},
		{"test5", "value5", 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Exec("DELETE FROM test_table")

		for _, row := range rows {
			_, err := db.Exec("INSERT INTO test_table (name, value, count) VALUES (?, ?, ?)", row...)
			if err != nil {
				b.Fatalf("Insert failed: %v", err)
			}
		}
	}
}

func BenchmarkLoopInsert_Large(b *testing.B) {
	db := setupBulkTestDB(&testing.T{})
	defer db.Close()

	rows := make([][]any, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = []any{fmt.Sprintf("test%d", i), fmt.Sprintf("value%d", i), i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Exec("DELETE FROM test_table")

		for _, row := range rows {
			_, err := db.Exec("INSERT INTO test_table (name, value, count) VALUES (?, ?, ?)", row...)
			if err != nil {
				b.Fatalf("Insert failed: %v", err)
			}
		}
	}
}
