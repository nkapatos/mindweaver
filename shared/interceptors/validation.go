package interceptors

import (
	"context"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValidationInterceptor is a singleton Connect-RPC interceptor that validates
// incoming proto messages using protovalidate.
// It validates all requests before they reach the service handler.
var ValidationInterceptor = mustNewValidationInterceptor()

// mustNewValidationInterceptor creates a validation interceptor.
// Panics if the validator cannot be initialized (fail-fast at startup).
func mustNewValidationInterceptor() connect.UnaryInterceptorFunc {
	validator, err := protovalidate.New()
	if err != nil {
		panic("failed to initialize protovalidate validator: " + err.Error())
	}

	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
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
}
