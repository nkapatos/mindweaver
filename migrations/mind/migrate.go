// Package mind contains mind service migrations
package mind

import (
	"database/sql"
	"embed"
	"log/slog"

	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/migrator"
)

//go:embed *.sql
var migrations embed.FS

// RunMigrations runs all up migrations for the mind service.
func RunMigrations(db *sql.DB, logger *slog.Logger) error {
	return migrator.RunMigrations(db, migrations, logger)
}
