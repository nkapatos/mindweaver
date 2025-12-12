// Package testdb provides shared test database utilities.
package testdb

import (
	"database/sql"
	"log/slog"
	"testing"

	_ "modernc.org/sqlite"
)

// MigrationRunner defines the interface for running service-specific migrations.
type MigrationRunner func(*sql.DB, *slog.Logger) error

// SetupTestDB creates an in-memory SQLite database with migrations applied.
// Returns a configured *sql.DB ready for testing.
//
// Usage:
//
//	import mindmigrations "github.com/nkapatos/mindweaver/migrations/mind"
//	db := testdb.SetupTestDB(t, mindmigrations.RunMigrations)
//	defer db.Close()
func SetupTestDB(t *testing.T, runMigrations MigrationRunner) *sql.DB {
	t.Helper()

	// Create in-memory database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Configure SQLite for testing (fast in-memory mode)
	if _, err := db.Exec("PRAGMA journal_mode=MEMORY;"); err != nil {
		t.Fatalf("Failed to set journal mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=OFF;"); err != nil {
		t.Fatalf("Failed to set synchronous mode: %v", err)
	}

	// Run service-specific migrations
	logger := NewTestLogger(t)
	if err := runMigrations(db, logger); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// NewTestLogger creates a logger suitable for testing.
// Logs are written to t.Log() and only shown on test failure or with -v flag.
// This is more useful than io.Discard for debugging failing tests.
func NewTestLogger(t *testing.T) *slog.Logger {
	t.Helper()

	// Create a custom handler that writes to t.Log()
	handler := slog.NewTextHandler(&testLogWriter{t: t}, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors by default in tests
	})

	return slog.New(handler)
}

// testLogWriter adapts testing.T to io.Writer for slog.
type testLogWriter struct {
	t *testing.T
}

func (w *testLogWriter) Write(p []byte) (n int, err error) {
	w.t.Helper()
	w.t.Log(string(p))
	return len(p), nil
}
