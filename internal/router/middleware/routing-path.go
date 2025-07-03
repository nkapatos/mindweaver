package middleware

import (
	"context"

	"github.com/labstack/echo/v4"
)

type routerPathCtxKey string

const (
	RouterPath routerPathCtxKey = "router_path"
)

func RouterPathMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			ctx := context.WithValue(c.Request().Context(), RouterPath, path)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func ActivedRoute(ctx context.Context) string {
	if ap, ok := ctx.Value(RouterPath).(string); ok {
		return ap
	}
	return "/"
}
