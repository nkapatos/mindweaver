package integration

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

func SetupTestDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	provider := SetupGooseProvider(db)
	err = RunUpMigrations(provider)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// Setup goose provider for the test database
func SetupGooseProvider(db *sql.DB) *goose.Provider {
	provider, err := goose.NewProvider(
		goose.DialectSQLite3,
		db,
		os.DirFS("../../migrations"),
		goose.WithVerbose(false),
	)
	if err != nil {
		log.Fatal(err)
	}

	return provider
}

// db migrations setup
func RunUpMigrations(provider *goose.Provider) error {
	// run the migrations
	_, err := provider.Up(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func RunDownMigrations() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	provider := SetupGooseProvider(db)

	_, err = provider.Down(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
