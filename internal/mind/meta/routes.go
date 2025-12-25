// NoteMeta V3 Routes Registration
// Registers Connect-RPC routes for note metadata sub-resource
package meta

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

// RegisterNoteMetaRoutes registers Connect-RPC routes for NoteMeta V3
// Routes: GET /v3/notes/{note_id}/meta
func RegisterNoteMetaRoutes(e *echo.Echo, handler *NoteMetaHandler, logger *slog.Logger) error {
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

	// Create Connect-RPC handler with validation
	path, connectHandler := mindv3connect.NewNoteMetaServiceHandler(
		handler,
		connect.WithInterceptors(validationInterceptor),
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
