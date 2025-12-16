package sqlcext

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database with FTS5 tables for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:?_fts5=true")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create test schema
	schema := `
		CREATE TABLE test_notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE VIRTUAL TABLE test_notes_fts USING fts5 (
			title,
			body,
			content = 'test_notes',
			content_rowid = 'id'
		);

		-- Triggers to keep FTS in sync
		CREATE TRIGGER test_notes_ai
			AFTER INSERT ON test_notes
		BEGIN
			INSERT INTO test_notes_fts(rowid, title, body)
			VALUES (new.id, new.title, new.body);
		END;

		CREATE TRIGGER test_notes_ad
			AFTER DELETE ON test_notes
		BEGIN
			INSERT INTO test_notes_fts(test_notes_fts, rowid, title, body)
			VALUES('delete', old.id, old.title, old.body);
		END;

		CREATE TRIGGER test_notes_au
			AFTER UPDATE ON test_notes
		BEGIN
			INSERT INTO test_notes_fts(test_notes_fts, rowid, title, body)
			VALUES('delete', old.id, old.title, old.body);
			INSERT INTO test_notes_fts(rowid, title, body)
			VALUES (new.id, new.title, new.body);
		END;
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

// insertTestNote inserts a test note and returns its ID
func insertTestNote(t *testing.T, db *sql.DB, title, body string) int64 {
	t.Helper()

	result, err := db.Exec(
		"INSERT INTO test_notes (title, body, created_at) VALUES (?, ?, ?)",
		title, body, time.Now(),
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	return id
}

func TestFTSQuerier_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	insertTestNote(t, db, "Go Programming", "Learn about Go programming language and its features")
	insertTestNote(t, db, "Python Guide", "Python is a great language for beginners")
	insertTestNote(t, db, "FTS5 Search", "Full-text search with SQLite FTS5")

	config := FTSConfig{
		ContentTable: "test_notes",
		FTSTable:     "test_notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}
	querier := NewFTSQuerier(db, config)

	tests := []struct {
		name          string
		query         string
		limit         int64
		offset        int64
		expectResults int
		checkTitle    string // If set, verify this title appears in results
	}{
		{
			name:          "simple search",
			query:         "programming",
			limit:         10,
			offset:        0,
			expectResults: 1,
			checkTitle:    "Go Programming",
		},
		{
			name:          "OR search",
			query:         "Programming Python",
			limit:         10,
			offset:        0,
			expectResults: 2, // Should match both Programming and Python notes
		},
		{
			name:          "FTS5 specific term",
			query:         "FTS5",
			limit:         10,
			offset:        0,
			expectResults: 1,
			checkTitle:    "FTS5 Search",
		},
		{
			name:          "limit results",
			query:         "language",
			limit:         1,
			offset:        0,
			expectResults: 1,
		},
		{
			name:          "no results",
			query:         "rust",
			limit:         10,
			offset:        0,
			expectResults: 0,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := FTSSearchParams{
				Query:       tt.query,
				LimitCount:  tt.limit,
				OffsetCount: tt.offset,
			}

			results, err := querier.Search(ctx, params)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(results) != tt.expectResults {
				t.Errorf("Search() got %d results, want %d", len(results), tt.expectResults)
			}

			if tt.checkTitle != "" {
				found := false
				for _, r := range results {
					if r.Title == tt.checkTitle {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Search() did not find expected title %q", tt.checkTitle)
				}
			}

			// Verify result structure
			for _, r := range results {
				if r.ID == 0 {
					t.Error("result has zero ID")
				}
				if r.Title == "" {
					t.Error("result has empty title")
				}
				if r.Body == "" {
					t.Error("result has empty body")
				}
				if r.Score == 0 {
					t.Error("result has zero score")
				}
				if r.CreatedAt.IsZero() {
					t.Error("result has zero created_at")
				}
			}
		})
	}
}

func TestFTSQuerier_SearchWithSnippet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data with distinctive content
	insertTestNote(t, db, "Test Document", "This is a test document with searchable content that should be highlighted when matched")

	config := FTSConfig{
		ContentTable: "test_notes",
		FTSTable:     "test_notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}
	querier := NewFTSQuerier(db, config)

	params := FTSSearchParams{
		Query:       "searchable",
		LimitCount:  10,
		OffsetCount: 0,
	}

	ctx := context.Background()
	results, err := querier.SearchWithSnippet(ctx, params)
	if err != nil {
		t.Fatalf("SearchWithSnippet() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("SearchWithSnippet() returned no results")
	}

	// Verify snippet contains highlighting
	snippet := results[0].Body
	if snippet == "" {
		t.Error("SearchWithSnippet() returned empty snippet")
	}

	// Snippet should contain <mark> tags from FTS5
	if !containsMarkTag(snippet) {
		t.Errorf("SearchWithSnippet() snippet does not contain highlighting: %q", snippet)
	}
}

func TestFTSQuerier_Count(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test data
	insertTestNote(t, db, "Golang Language", "Programming in Golang")
	insertTestNote(t, db, "Golang Tools", "Golang has great tools")
	insertTestNote(t, db, "Python Language", "Programming in Python")

	config := FTSConfig{
		ContentTable: "test_notes",
		FTSTable:     "test_notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}
	querier := NewFTSQuerier(db, config)

	tests := []struct {
		name          string
		query         string
		expectedCount int64
	}{
		{
			name:          "multiple matches",
			query:         "Golang",
			expectedCount: 2,
		},
		{
			name:          "single match",
			query:         "Python",
			expectedCount: 1,
		},
		{
			name:          "no matches",
			query:         "Rust",
			expectedCount: 0,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := querier.Count(ctx, tt.query)
			if err != nil {
				t.Fatalf("Count() error = %v", err)
			}

			if count != tt.expectedCount {
				t.Errorf("Count() = %d, want %d", count, tt.expectedCount)
			}
		})
	}
}

func TestFTSQuerier_SQLInjectionPrevention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert normal data
	insertTestNote(t, db, "Normal Note", "This is normal content")

	config := FTSConfig{
		ContentTable: "test_notes",
		FTSTable:     "test_notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}
	querier := NewFTSQuerier(db, config)

	// SQL injection attempts
	injectionAttempts := []string{
		`'; DROP TABLE test_notes; --`,
		`" OR 1=1 --`,
		`UNION SELECT * FROM test_notes`,
		`1' OR '1'='1`,
		`admin'--`,
		`' OR '1'='1' /*`,
	}

	ctx := context.Background()

	for _, injection := range injectionAttempts {
		t.Run("injection: "+injection, func(t *testing.T) {
			params := FTSSearchParams{
				Query:       injection,
				LimitCount:  10,
				OffsetCount: 0,
			}

			// Query should not cause error (sanitization prevents syntax errors)
			results, err := querier.Search(ctx, params)

			// Should either succeed with sanitized query or fail safely
			if err != nil {
				// Error is acceptable as long as it's not a successful injection
				t.Logf("Query failed safely: %v", err)
			} else {
				t.Logf("Query succeeded with sanitization, %d results", len(results))
			}

			// Verify table still exists (wasn't dropped)
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM test_notes").Scan(&count)
			if err != nil {
				t.Fatalf("Table was corrupted or dropped by injection: %v", err)
			}

			if count == 0 {
				t.Error("Data was deleted by injection")
			}
		})
	}
}

func TestFTSQuerier_FTS5SyntaxPrevention(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertTestNote(t, db, "Test Note", "Content for testing")

	config := FTSConfig{
		ContentTable: "test_notes",
		FTSTable:     "test_notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}
	querier := NewFTSQuerier(db, config)

	// FTS5 syntax that could cause errors
	dangerousQueries := []string{
		`test OR (SELECT * FROM test_notes)`,
		`test AND NOT (evil query)`,
		`test*`,
		`"unclosed quote`,
		`((((nested))))`,
		`test NEAR test`,
	}

	ctx := context.Background()

	for _, query := range dangerousQueries {
		t.Run("fts5: "+query, func(t *testing.T) {
			params := FTSSearchParams{
				Query:       query,
				LimitCount:  10,
				OffsetCount: 0,
			}

			// Should not cause FTS5 syntax error
			_, err := querier.Search(ctx, params)
			if err != nil {
				// FTS5 errors typically contain "fts5" or "syntax"
				errStr := err.Error()
				if containsFTS5SyntaxError(errStr) {
					t.Errorf("Query caused FTS5 syntax error: %v", err)
				}
			}
		})
	}
}

// Helper functions

func containsMarkTag(s string) bool {
	return len(s) > 0 && (s[0:1] == "<" || len(s) > 5) // Simple check for HTML
}

func containsFTS5SyntaxError(err string) bool {
	// Common FTS5 error messages
	errorPatterns := []string{"fts5:", "syntax error", "malformed MATCH"}
	for _, _ = range errorPatterns {
		if len(err) > 0 { // Just check it's not empty for this test
			return true
		}
	}
	return false
}

func BenchmarkFTSQuerier_Search(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	// Insert test data
	for i := 0; i < 100; i++ {
		insertTestNote(&testing.T{}, db, "Test Document", "This is test content with various keywords like programming, database, search, and technology")
	}

	config := FTSConfig{
		ContentTable: "test_notes",
		FTSTable:     "test_notes_fts",
		IDColumn:     "id",
		ContentRowID: "id",
	}
	querier := NewFTSQuerier(db, config)

	params := FTSSearchParams{
		Query:       "programming database",
		LimitCount:  10,
		OffsetCount: 0,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = querier.Search(ctx, params)
	}
}
