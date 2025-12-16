// Templates V3 Route Registration (Connect-RPC)
package templates

import (
	"context"
	"log/slog"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/packages/mindweaver/gen/proto/mind/v3/mindv3connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// RegisterTemplatesV3Routes registers V3 templates routes (Connect-RPC with both gRPC and HTTP/JSON support)
func RegisterTemplatesV3Routes(e *echo.Echo, handler *TemplatesHandlerV3, logger *slog.Logger) error {
	// Connect-RPC automatically supports:
	// - gRPC (binary protobuf over HTTP/2)
	// - gRPC-Web (for browsers)
	// - Connect protocol (JSON or binary over HTTP/1.1 or HTTP/2)

	// Initialize protovalidate validator
	validator, err := protovalidate.New()
	if err != nil {
		return err
	}

	// Create validation interceptor
	validationInterceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Validate request message (cast to proto.Message)
			if msg, ok := req.Any().(interface{ ProtoReflect() protoreflect.Message }); ok {
				if err := validator.Validate(msg); err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
			}
			return next(ctx, req)
		}
	})

	// Create handler with validation interceptor
	path, connectHandler := mindv3connect.NewTemplatesServiceHandler(
		handler,
		connect.WithInterceptors(validationInterceptor),
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
