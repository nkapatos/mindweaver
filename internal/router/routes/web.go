package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, settingsHandler *web.SettingsHandler) {
	e.GET("/", homeHandler.Home)

	// Prompts
	e.GET("/prompts", promptsHandler.Prompts)
	e.POST("/prompts", promptsHandler.CreatePrompt)
	e.POST("/prompts/delete", promptsHandler.DeletePrompt)

	// Settings
	e.GET("/settings", settingsHandler.Settings)
}
