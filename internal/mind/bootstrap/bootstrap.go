// Package bootstrap provides initialization for the Mind service (Notes/PKM/Knowledge).
// This allows the mind service to be run standalone or combined with other services.
package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/nkapatos/mindweaver/internal/mind/collections"
	"github.com/nkapatos/mindweaver/internal/mind/links"
	"github.com/nkapatos/mindweaver/internal/mind/meta"
	"github.com/nkapatos/mindweaver/internal/mind/notes"
	"github.com/nkapatos/mindweaver/internal/mind/notetypes"
	"github.com/nkapatos/mindweaver/internal/mind/search"
	"github.com/nkapatos/mindweaver/internal/mind/store"
	"github.com/nkapatos/mindweaver/internal/mind/tags"
	"github.com/nkapatos/mindweaver/internal/mind/templates"
	"github.com/nkapatos/mindweaver/internal/mind/titleindex"
	mindmigrations "github.com/nkapatos/mindweaver/migrations/mind"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"
)

// Initialize sets up the Mind service on the given API group.
// It handles database initialization, migration, service setup, and route registration.
//
// Parameters:
//   - e: Echo instance (needed for Connect-RPC V3 routes)
//   - apiGroup: Echo API group to register routes under (will create /mind subgroup)
//   - dbPath: Path to the SQLite database file
//   - logger: Structured logger
//
// Returns the database connection, notes service, and error if initialization fails.
// The caller is responsible for closing the returned database connection.
// The notes service is returned for scheduler integration in combined mode.
func Initialize(e *echo.Echo, apiGroup *echo.Group, dbPath string, logger *slog.Logger) (*sql.DB, *notes.NotesService, error) {
	logger.Info("🧠 Initializing Mind service (Notes/PKM)")

	// Open database connection
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open notes database: %w", err)
	}

	// Configure WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to enable WAL mode for notes: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to enable WAL synchronous mode for notes: %w", err)
	}
	if _, err := db.Exec("PRAGMA wal_autocheckpoint=100;"); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to enable WAL checkpoint for notes: %w", err)
	}

	// Run migrations
	if err := mindmigrations.RunMigrations(db, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to run notes DB migrations: %w", err)
	}

	logger.Info("Mind database initialized", "path", dbPath)

	// Initialize store and ensure default data exists
	querier := store.New(db)
	ctx := context.Background()

	// Ensure default data exists (idempotent)
	if err := notetypes.EnsureDefaultNoteTypes(ctx, querier, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to ensure default note types: %w", err)
	}
	if err := collections.EnsureDefaultCollections(ctx, querier, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to ensure default collections: %w", err)
	}

	// Extract directory from database path for title index
	dbDir := filepath.Dir(dbPath)

	// Initialize title index (BadgerDB for title→uuid lookup)
	titleIndexPath := filepath.Join(dbDir, "titleindex")
	titleIndex, err := titleindex.NewTitleIndex(titleIndexPath, logger)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to initialize title index: %w", err)
	}

	noteMetaService := meta.NewNoteMetaService(querier, db, logger, "Notes Meta Service")
	notesService := notes.NewNotesService(db, querier, logger, "Notes Service")
	tagService := tags.NewTagsService(querier, logger, "Tags Service")
	templateService := templates.NewTemplatesService(querier, logger, "Templates Service")
	linksService := links.NewLinksService(querier, logger, "Links Service")
	noteTypesService := notetypes.NewNoteTypesService(querier, logger, "NoteTypes Service")
	collectionsService := collections.NewCollectionsService(db, querier, logger, "Collections Service")
	searchService := search.NewSearchService(db, querier, logger)

	// Initialize handlers
	tagsHandlerV3 := tags.NewTagsHandlerV3(tagService)
	templatesHandlerV3 := templates.NewTemplatesHandlerV3(templateService)
	noteTypesHandlerV3 := notetypes.NewNoteTypesHandlerV3(noteTypesService)
	collectionsHandlerV3 := collections.NewCollectionsHandlerV3(collectionsService)
	notesHandlerV3 := notes.NewNotesHandlerV3(notesService, noteMetaService, linksService, tagService)
	noteMetaHandlerV3 := meta.NewNoteMetaHandlerV3(noteMetaService)
	searchHandlerV3 := search.NewSearchHandlerV3(searchService)

	// Create /mind subgroup under the provided API group
	// This creates routes like: /api/mind/import
	// mindGroup := apiGroup.Group("/mind")

	// Register V3 routes (Connect-RPC with protobuf) - supports gRPC + HTTP/JSON
	// Note: Connect-RPC requires registration at Echo root level (not in a group)
	if err := tags.RegisterTagsV3Routes(e, tagsHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register tags V3 routes: %w", err)
	}

	if err := templates.RegisterTemplatesV3Routes(e, templatesHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register templates V3 routes: %w", err)
	}

	if err := notetypes.RegisterNoteTypesV3Routes(e, noteTypesHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register note types V3 routes: %w", err)
	}

	if err := collections.RegisterCollectionsV3Routes(e, collectionsHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register collections V3 routes: %w", err)
	}

	if err := notes.RegisterNotesV3Routes(e, notesHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register notes V3 routes: %w", err)
	}

	if err := meta.RegisterNoteMetaV3Routes(e, noteMetaHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register note meta V3 routes: %w", err)
	}

	if err := search.RegisterSearchV3Routes(e, searchHandlerV3, logger); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to register search V3 routes: %w", err)
	}

	// FIXME: after we migrate the routes to v3
	// Register import routes (V1 REST - still used for file imports)
	// opStore := importservice.NewOperationStore()
	// batchProcessor := importservice.NewBatchProcessor(db, titleIndex, logger)
	// linkResolver := importservice.NewLinkResolver(db, titleIndex, logger)

	// importHandlers := importservice.NewImportHandlers(opStore, batchProcessor, linkResolver, logger)
	// importservice.RegisterImportRoutes(mindGroup, importHandlers)

	logger.Info("✅ Mind service ready")

	return db, notesService, nil
}
