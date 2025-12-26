// Templates V3 Route Registration (Connect-RPC)
package templates

import (
	"log/slog"

	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/gen/proto/mind/v3/mindv3connect"
	"github.com/nkapatos/mindweaver/shared/interceptors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// RegisterTemplatesRoutes registers V3 templates routes (Connect-RPC with both gRPC and HTTP/JSON support)
func RegisterTemplatesRoutes(e *echo.Echo, handler *TemplatesHandler, logger *slog.Logger) error {
	// Connect-RPC automatically supports:
	// - gRPC (binary protobuf over HTTP/2)
	// - gRPC-Web (for browsers)
	// - Connect protocol (JSON or binary over HTTP/1.1 or HTTP/2)

	// Create handler with validation interceptor
	path, connectHandler := mindv3connect.NewTemplatesServiceHandler(
		handler,
		connect.WithInterceptors(interceptors.ValidationInterceptor),
	)

	// Wrap Connect handler for Echo
	// Connect needs HTTP/2 for gRPC, h2c allows HTTP/2 without TLS
	h2cHandler := h2c.NewHandler(connectHandler, &http2.Server{})

	// Register Connect handler directly - it handles its own routing
	// Use Match to catch all methods and let Connect handle routing
	e.Match([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, path+"*", echo.WrapHandler(h2cHandler))

	// NOTE: Connect-RPC error format is close to AIP-193 but not exact:
	// - Connect: {code: "invalid_argument", message: "..."}
	// - AIP-193: {error: {code: 400, message: "...", status: "INVALID_ARGUMENT", details: [...]}}
	// For full AIP-193, would need custom error interceptor (deferred post-PoC)

	logger.Info("Registered V3 Templates routes with automatic validation", "path", path)
	return nil
}
