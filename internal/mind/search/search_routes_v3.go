// Search V3 Route Registration (Connect-RPC)
package search

import (
	"log/slog"

	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/shared/interceptors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// RegisterSearchV3Routes registers V3 search routes (Connect-RPC with both gRPC and HTTP/JSON support)
func RegisterSearchV3Routes(e *echo.Echo, handler *SearchHandlerV3, logger *slog.Logger) error {
	// Create handler with validation interceptor
	path, connectHandler := mindv3connect.NewSearchServiceHandler(
		handler,
		connect.WithInterceptors(interceptors.ValidationInterceptor),
	)

	// Wrap Connect handler for Echo
	h2cHandler := h2c.NewHandler(connectHandler, &http2.Server{})

	// Register Connect handler - it handles its own routing
	e.Match([]string{"GET", "POST"}, path+"*", echo.WrapHandler(h2cHandler))

	logger.Info("Registered V3 Search routes with automatic validation", "path", path)
	return nil
}
