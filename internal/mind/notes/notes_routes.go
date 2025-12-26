// Notes V3 Route Registration (Connect-RPC)
package notes

import (
	"log/slog"

	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/shared/interceptors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// RegisterNotesRoutes registers V3 notes routes (Connect-RPC with both gRPC and HTTP/JSON support)
func RegisterNotesRoutes(e *echo.Echo, handler *NotesHandler, logger *slog.Logger) error {
	// Create handler with validation interceptor
	path, connectHandler := mindv3connect.NewNotesServiceHandler(
		handler,
		connect.WithInterceptors(interceptors.ValidationInterceptor),
	)

	// Wrap Connect handler for Echo
	// Connect needs HTTP/2 for gRPC, h2c allows HTTP/2 without TLS
	h2cHandler := h2c.NewHandler(connectHandler, &http2.Server{})

	// Register Connect handler directly - it handles its own routing
	// Use Match to catch all methods and let Connect handle routing
	e.Match([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, path+"*", echo.WrapHandler(h2cHandler))

	logger.Info("Registered V3 Notes routes with automatic validation", "path", path)
	return nil
}
