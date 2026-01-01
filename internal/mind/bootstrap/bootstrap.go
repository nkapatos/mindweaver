// Package bootstrap provides initialization for the Mind service (Notes/PKM/Knowledge).
// This allows the mind service to be run standalone or combined with other services.
package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	_ "modernc.org/sqlite"

	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/internal/mind/collections"
	"github.com/nkapatos/mindweaver/internal/mind/events"
	"github.com/nkapatos/mindweaver/internal/mind/gen/store"
	"github.com/nkapatos/mindweaver/internal/mind/links"
	"github.com/nkapatos/mindweaver/internal/mind/meta"
	"github.com/nkapatos/mindweaver/internal/mind/notes"
	"github.com/nkapatos/mindweaver/internal/mind/notetypes"
	"github.com/nkapatos/mindweaver/internal/mind/search"
	"github.com/nkapatos/mindweaver/internal/mind/tags"
	"github.com/nkapatos/mindweaver/internal/mind/templates"
	mindmigrations "github.com/nkapatos/mindweaver/migrations/mind"
	"github.com/nkapatos/mindweaver/shared/interceptors"
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
// Returns the database connection, notes service, event hub, and error if initialization fails.
// The caller is responsible for closing the returned database connection and event hub.
// The notes service is returned for scheduler integration in combined mode.
// The event hub is returned for graceful shutdown and can be used by other services to publish events.
func Initialize(e *echo.Echo, apiGroup *echo.Group, dbPath string, logger *slog.Logger) (*sql.DB, *notes.NotesService, events.Hub, error) {
	logger.Info("🧠 Initializing Mind service (Notes/PKM)")

	// Open database connection
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared&mode=rwc", dbPath))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open notes database: %w", err)
	}

	// Configure WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to enable WAL mode for notes: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to enable WAL synchronous mode for notes: %w", err)
	}
	if _, err := db.Exec("PRAGMA wal_autocheckpoint=100;"); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to enable WAL checkpoint for notes: %w", err)
	}

	// Run migrations
	if err := mindmigrations.RunMigrations(db, logger); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to run notes DB migrations: %w", err)
	}

	logger.Info("Mind database initialized", "path", dbPath)

	// Initialize store and ensure default data exists
	querier := store.New(db)
	ctx := context.Background()

	// Ensure default data exists (idempotent)
	if err := notetypes.EnsureDefaultNoteTypes(ctx, querier, logger); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to ensure default note types: %w", err)
	}
	if err := collections.EnsureDefaultCollections(ctx, querier, logger); err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("failed to ensure default collections: %w", err)
	}

	// Note: titleindex initialization removed - See issue #37 and #43

	noteMetaService := meta.NewNoteMetaService(querier, db, logger, "Notes Meta Service")
	notesService := notes.NewNotesService(db, querier, logger, "Notes Service")
	tagService := tags.NewTagsService(querier, logger, "Tags Service")
	templateService := templates.NewTemplatesService(querier, logger, "Templates Service")
	linksService := links.NewLinksService(querier, logger, "Links Service")
	noteTypesService := notetypes.NewNoteTypesService(querier, logger, "NoteTypes Service")
	collectionsService := collections.NewCollectionsService(db, querier, logger, "Collections Service")
	searchService := search.NewSearchService(db, querier, logger)

	// Initialize handlers
	tagsHandler := tags.NewTagsHandler(tagService)
	templatesHandler := templates.NewTemplatesHandler(templateService)
	noteTypesHandler := notetypes.NewNoteTypesHandler(noteTypesService)
	collectionsHandler := collections.NewCollectionsHandler(collectionsService)
	notesHandler := notes.NewNotesHandler(notesService, noteMetaService, linksService, tagService)
	noteMetaHandler := meta.NewNoteMetaHandler(noteMetaService)
	searchHandlerV3 := search.NewSearchHandlerV3(searchService)

	// Register V3 routes (Connect-RPC with protobuf) - supports gRPC + HTTP/JSON
	// Connect-RPC requires registration at Echo root level (not in a group)
	validationOpt := connect.WithInterceptors(interceptors.ValidationInterceptor)

	type serviceReg struct {
		name    string
		path    string
		handler http.Handler
	}

	tagsPath, tagsConnHandler := mindv3connect.NewTagsServiceHandler(tagsHandler, validationOpt)
	templatesPath, templatesConnHandler := mindv3connect.NewTemplatesServiceHandler(templatesHandler, validationOpt)
	noteTypesPath, noteTypesConnHandler := mindv3connect.NewNoteTypesServiceHandler(noteTypesHandler, validationOpt)
	collectionsPath, collectionsConnHandler := mindv3connect.NewCollectionsServiceHandler(collectionsHandler, validationOpt)
	notesPath, notesConnHandler := mindv3connect.NewNotesServiceHandler(notesHandler, validationOpt)
	noteMetaPath, noteMetaConnHandler := mindv3connect.NewNoteMetaServiceHandler(noteMetaHandler, validationOpt)
	searchPath, searchConnHandler := mindv3connect.NewSearchServiceHandler(searchHandlerV3, validationOpt)

	services := []serviceReg{
		{"Tags", tagsPath, tagsConnHandler},
		{"Templates", templatesPath, templatesConnHandler},
		{"NoteTypes", noteTypesPath, noteTypesConnHandler},
		{"Collections", collectionsPath, collectionsConnHandler},
		{"Notes", notesPath, notesConnHandler},
		{"NoteMeta", noteMetaPath, noteMetaConnHandler},
		{"Search", searchPath, searchConnHandler},
	}

	for _, svc := range services {
		registerConnectService(e, logger, svc.name, svc.path, svc.handler)
	}

	// Initialize event hub and SSE handler
	eventHub := events.NewHub(logger)
	sseHandler := events.NewSSEHandler(eventHub, logger)

	// Register SSE endpoint for real-time events
	e.GET("/events/stream", sseHandler.HandleStream)
	logger.Info("Registered SSE endpoint", "path", "/events/stream")

	// Note: Import service registration removed - See issue #37 for decision on restoration

	logger.Info("✅ Mind service ready")

	return db, notesService, eventHub, nil
}

// registerConnectService registers a Connect-RPC service handler with Echo.
// Connect-RPC supports gRPC (binary protobuf over HTTP/2), gRPC-Web (for browsers),
// and Connect protocol (JSON or binary over HTTP/1.1 or HTTP/2).
func registerConnectService(e *echo.Echo, logger *slog.Logger, serviceName, path string, handler http.Handler) {
	// Wrap in h2c handler for HTTP/2 without TLS (needed for gRPC)
	h2cHandler := h2c.NewHandler(handler, &http2.Server{})

	// Register with Echo - Match all methods and let Connect handle routing
	e.Match([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, path+"*", echo.WrapHandler(h2cHandler))

	logger.Info("Registered V3 routes", "service", serviceName, "path", path)
}
