package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, providersHandler *web.ProvidersHandler, settingsHandler *web.SettingsHandler) {
	e.GET("/", homeHandler.Home)

	// Prompts
	e.GET("/prompts", promptsHandler.Prompts)
	e.POST("/prompts", promptsHandler.CreatePrompt)
	e.GET("/prompts/edit/:id", promptsHandler.EditPrompt)
	e.POST("/prompts/edit/:id", promptsHandler.UpdatePrompt)
	e.POST("/prompts/delete", promptsHandler.DeletePrompt)

	// Providers
	e.GET("/providers", providersHandler.Providers)

	// Settings
	e.GET("/settings", settingsHandler.Settings)
}
