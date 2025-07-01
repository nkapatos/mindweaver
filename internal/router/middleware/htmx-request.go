package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	IsHtmxRequestKey contextKey = "is_htmx_request"
)

func HTMXMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isHtmx := c.Request().Header.Get("HX-Request") == "true"
			ctx := context.WithValue(c.Request().Context(), IsHtmxRequestKey, isHtmx)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func IsHtmxRequest(ctx context.Context) bool {
	if isHtmx, ok := ctx.Value(IsHtmxRequestKey).(bool); ok {
		return isHtmx
	}
	return false
}
