// Search V3 Route Registration (Connect-RPC)
package search

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

// RegisterSearchV3Routes registers V3 search routes (Connect-RPC with both gRPC and HTTP/JSON support)
func RegisterSearchV3Routes(e *echo.Echo, handler *SearchHandlerV3, logger *slog.Logger) error {
	// Initialize protovalidate validator
	validator, err := protovalidate.New()
	if err != nil {
		return err
	}

	// Create validation interceptor
	validationInterceptor := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Validate request message
			if msg, ok := req.Any().(interface{ ProtoReflect() protoreflect.Message }); ok {
				if err := validator.Validate(msg); err != nil {
					return nil, connect.NewError(connect.CodeInvalidArgument, err)
				}
			}
			return next(ctx, req)
		}
	})

	// Create handler with validation interceptor
	path, connectHandler := mindv3connect.NewSearchServiceHandler(
		handler,
		connect.WithInterceptors(validationInterceptor),
	)

	// Wrap Connect handler for Echo
	h2cHandler := h2c.NewHandler(connectHandler, &http2.Server{})

	// Register Connect handler - it handles its own routing
	e.Match([]string{"GET", "POST"}, path+"*", echo.WrapHandler(h2cHandler))

	logger.Info("Registered V3 Search routes with automatic validation", "path", path)
	return nil
}
