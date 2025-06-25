package routes

import (
	"github.com/labstack/echo/v4"
)

// SetupStaticRoutes configures static asset routes
func SetupStaticRoutes(e *echo.Echo) {
	// Static assets
	e.Static("/assets", "dist")
}
