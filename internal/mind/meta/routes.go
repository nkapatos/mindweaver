// NoteMeta V3 Routes Registration
// Registers Connect-RPC routes for note metadata sub-resource
package meta

import (
	"log/slog"

	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/shared/interceptors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// RegisterNoteMetaRoutes registers Connect-RPC routes for NoteMeta V3
// Routes: GET /v3/notes/{note_id}/meta
func RegisterNoteMetaRoutes(e *echo.Echo, handler *NoteMetaHandler, logger *slog.Logger) error {
	// Create Connect-RPC handler with validation
	path, connectHandler := mindv3connect.NewNoteMetaServiceHandler(
		handler,
		connect.WithInterceptors(interceptors.ValidationInterceptor),
	)

	// Wrap in h2c handler (HTTP/2 without TLS)
	h2cHandler := h2c.NewHandler(connectHandler, &http2.Server{})

	// Register with Echo - Connect-RPC uses POST by default
	// Although proto defines GET, Connect protocol requires POST for RPC calls
	e.Match([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, path+"*", echo.WrapHandler(h2cHandler))

	logger.Info("Registered V3 NoteMeta routes with automatic validation",
		"path", path)

	return nil
}
