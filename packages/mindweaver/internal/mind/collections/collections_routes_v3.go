package collections

import (
	"context"
	"log/slog"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/pkg/gen/proto/mind/v3/mindv3connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// RegisterCollectionsV3Routes registers V3 collections routes (Connect-RPC with both gRPC and HTTP/JSON support)
func RegisterCollectionsV3Routes(e *echo.Echo, handler *CollectionsHandlerV3, logger *slog.Logger) error {
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
	path, connectHandler := mindv3connect.NewCollectionsServiceHandler(
		handler,
		connect.WithInterceptors(validationInterceptor),
	)

	// Wrap Connect handler for Echo
	// Connect needs HTTP/2 for gRPC, h2c allows HTTP/2 without TLS
	h2cHandler := h2c.NewHandler(connectHandler, &http2.Server{})

	// Register Connect handler directly - it handles its own routing
	// Use Match to catch all methods and let Connect handle routing
	e.Match([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, path+"*", echo.WrapHandler(h2cHandler))

	logger.Info("Registered V3 Collections routes with automatic validation", "path", path)
	return nil
}
