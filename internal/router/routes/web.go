package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/config"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, providersHandler *web.ProvidersHandler, llmServicesHandler *web.LLMServicesHandler, llmServiceConfigsHandler *web.LLMServiceConfigsHandler, settingsHandler *web.SettingsHandler, conversationHandler *web.ConversationHandler) {
	e.GET(config.RouteHome, homeHandler.Home)

	// Prompts
	e.GET(config.RoutePrompts, promptsHandler.Prompts)
	e.POST(config.RoutePrompts, promptsHandler.CreatePrompt)
	e.GET(config.RoutePromptsEdit, promptsHandler.EditPrompt)
	e.POST(config.RoutePromptsEdit, promptsHandler.UpdatePrompt)
	e.POST(config.RoutePromptsDelete, promptsHandler.DeletePrompt)

	// Providers
	e.GET(config.RouteProviders, providersHandler.Providers)
	e.POST(config.RouteProviders, providersHandler.CreateProvider)
	e.GET(config.RouteProvidersEdit, providersHandler.EditProvider)
	e.POST(config.RouteProvidersEdit, providersHandler.UpdateProvider)
	e.POST(config.RouteProvidersDelete, providersHandler.DeleteProvider)

	// LLM Services
	e.GET(config.RouteLLMServices, llmServicesHandler.LLMServices)
	e.POST(config.RouteLLMServices, llmServicesHandler.CreateLLMService)
	e.GET(config.RouteLLMServicesEdit, llmServicesHandler.EditLLMService)
	e.POST(config.RouteLLMServicesEdit, llmServicesHandler.UpdateLLMService)
	e.POST(config.RouteLLMServicesDelete, llmServicesHandler.DeleteLLMService)
	e.GET(config.RouteLLMServicesModels, llmServicesHandler.GetModels)

	// LLM Service Configurations
	e.GET(config.RouteLLMServiceConfigs, llmServiceConfigsHandler.LLMServiceConfigs)
	e.POST(config.RouteLLMServiceConfigs, llmServiceConfigsHandler.CreateLLMServiceConfig)
	e.GET(config.RouteLLMServiceConfigsEdit, llmServiceConfigsHandler.EditLLMServiceConfig)
	e.POST(config.RouteLLMServiceConfigsEdit, llmServiceConfigsHandler.UpdateLLMServiceConfig)
	e.POST(config.RouteLLMServiceConfigsDelete, llmServiceConfigsHandler.DeleteLLMServiceConfig)
	e.GET(config.RouteLLMServiceConfigsModels, llmServiceConfigsHandler.GetModelsForService)

	// Settings
	e.GET(config.RouteSettings, settingsHandler.Settings)

	// Conversations
	e.GET(config.RouteConversations, conversationHandler.Conversation)
	e.GET(config.RouteConversationsNew, conversationHandler.NewConversation)
	e.POST(config.RouteConversationsCreate, conversationHandler.CreateConversation)
	e.GET(config.RouteConversationsView, conversationHandler.ViewConversation)
}
