// Package migrator provides shared database migration functionality.
package migrator

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// RunMigrations runs all up migrations using embedded migration files.
// It will create the DB schema if not present.
// The migrations embed.FS should contain *.sql files in the root.
func RunMigrations(db *sql.DB, migrations embed.FS, logger *slog.Logger) error {
	goose.SetBaseFS(migrations)
	goose.SetLogger(gooseLogger{logger})
	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// gooseLogger adapts slog.Logger to goose's logger interface.
type gooseLogger struct {
	logger *slog.Logger
}

func (g gooseLogger) Fatal(v ...any) {
	g.logger.Error("goose fatal", "msg", fmt.Sprint(v...))
}

func (g gooseLogger) Fatalf(format string, v ...any) {
	g.logger.Error("goose fatal", "msg", fmt.Sprintf(format, v...))
}

func (g gooseLogger) Print(v ...any) {
	g.logger.Info("goose", "msg", fmt.Sprint(v...))
}

func (g gooseLogger) Printf(format string, v ...any) {
	g.logger.Info("goose", "msg", fmt.Sprintf(format, v...))
}
