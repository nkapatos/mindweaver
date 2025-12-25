// Package brain contains brain service migrations
package brain

import (
	"database/sql"
	"embed"
	"log/slog"

	"github.com/nkapatos/mindweaver/shared/migrator"
)

//go:embed *.sql
var migrations embed.FS

// RunMigrations runs all up migrations for the brain service.
func RunMigrations(db *sql.DB, logger *slog.Logger) error {
	return migrator.RunMigrations(db, migrations, logger)
}
