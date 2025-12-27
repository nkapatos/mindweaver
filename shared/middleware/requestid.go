package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// contextKey is a private type for context keys defined in this package.
type contextKey string

const (
	requestIDKey    contextKey = "requestID"
	RequestIDHeader string     = "X-Request-ID" // HTTP header name, not a context key
)

// RequestIDMiddleware injects a request ID into the context and response header.
// If X-Request-ID header is present in the request, it will be used; otherwise a new UUID is generated.
func RequestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		rid := req.Header.Get(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		// Store in context for downstream handlers/services
		ctx := context.WithValue(req.Context(), requestIDKey, rid)
		c.SetRequest(req.WithContext(ctx))
		// Always set header in response
		c.Response().Header().Set(RequestIDHeader, rid)
		return next(c)
	}
}

// GetRequestID extracts the request ID from context, or returns empty string if not found.
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if rid, ok := v.(string); ok {
			return rid
		}
	}
	return ""
}
