package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
)

// SetupAPIRoutes configures all API routes
func SetupAPIRoutes(e *echo.Echo, actorHandler *api.ActorHandler, promptHandler *api.PromptHandler) {
	// API routes
	e.POST("/api/actors", func(c echo.Context) error {
		actorHandler.CreateActor(c.Response().Writer, c.Request())
		return nil
	})
	e.POST("/api/prompts", func(c echo.Context) error {
		promptHandler.CreatePrompt(c.Response().Writer, c.Request())
		return nil
	})
}
