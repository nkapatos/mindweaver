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
	RouteConversations   = "/conversations"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, providersHandler *web.ProvidersHandler, settingsHandler *web.SettingsHandler, conversationHandler *web.ConversationHandler) {
	e.GET(RouteHome, homeHandler.Home)

	// Prompts
	e.GET(RoutePrompts, promptsHandler.Prompts)
	e.POST(RoutePrompts, promptsHandler.CreatePrompt)
	e.GET(RoutePromptsEdit, promptsHandler.EditPrompt)
	e.POST(RoutePromptsEdit, promptsHandler.UpdatePrompt)
	e.POST(RoutePromptsDelete, promptsHandler.DeletePrompt)

	// Providers
	e.GET(RouteProviders, providersHandler.Providers)
	e.POST(RouteProviders, providersHandler.CreateProvider)
	e.GET(RouteProvidersEdit, providersHandler.EditProvider)
	e.POST(RouteProvidersEdit, providersHandler.UpdateProvider)
	e.POST(RouteProvidersDelete, providersHandler.DeleteProvider)

	// Settings
	e.GET(RouteSettings, settingsHandler.Settings)

	// Conversation
	e.GET(RouteConversations, conversationHandler.Conversation)
}
