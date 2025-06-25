package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
)

var (
	RouteHome            = "/"
	RoutePrompts         = "/prompts"
	RoutePromptsEdit     = "/prompts/edit/:id"
	RoutePromptsDelete   = "/prompts/delete"
	RouteProviders       = "/providers"
	RouteProvidersEdit   = "/providers/edit/:id"
	RouteProvidersDelete = "/providers/delete"
	RouteSettings        = "/settings"
	RouteChats           = "/chats"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, providersHandler *web.ProvidersHandler, settingsHandler *web.SettingsHandler, chatsHandler *web.ChatsHandler) {
	e.GET("/", homeHandler.Home)

	// Prompts
	e.GET("/prompts", promptsHandler.Prompts)
	e.POST("/prompts", promptsHandler.CreatePrompt)
	e.GET("/prompts/edit/:id", promptsHandler.EditPrompt)
	e.POST("/prompts/edit/:id", promptsHandler.UpdatePrompt)
	e.POST("/prompts/delete", promptsHandler.DeletePrompt)

	// Providers
	e.GET("/providers", providersHandler.Providers)
	e.POST("/providers", providersHandler.CreateProvider)
	e.GET("/providers/edit/:id", providersHandler.EditProvider)
	e.POST("/providers/edit/:id", providersHandler.UpdateProvider)
	e.POST("/providers/delete", providersHandler.DeleteProvider)

	// Settings
	e.GET("/settings", settingsHandler.Settings)

	// Chats
	e.GET("/chats", chatsHandler.Chats)
}
