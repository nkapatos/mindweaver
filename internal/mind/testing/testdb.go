// Package testing provides test utilities for the Mind service.
package testing

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/migrations/mind"
	"github.com/nkapatos/mindweaver/shared/testdb"
)

// SetupTest creates a test database with Mind migrations and returns
// everything needed to test Mind services.
//
// Returns:
//   - *sql.DB: The test database (caller must defer db.Close())
//   - store.Querier: Mind store querier
//   - *slog.Logger: Test logger
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    db, querier, logger := testing.SetupTest(t)
//	    defer db.Close()
//	    // ... use querier and logger
//	}
func SetupTest(t *testing.T) (*sql.DB, store.Querier, *slog.Logger) {
	t.Helper()

	db := testdb.SetupTestDB(t, mind.RunMigrations)
	querier := store.New(db)
	logger := testdb.NewTestLogger(t)

	return db, querier, logger
}
