package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
)

var (
	RouteHome                    = "/"
	RoutePrompts                 = "/prompts"
	RoutePromptsEdit             = "/prompts/edit/:id"
	RoutePromptsDelete           = "/prompts/delete"
	RouteProviders               = "/providers"
	RouteProvidersEdit           = "/providers/edit/:id"
	RouteProvidersDelete         = "/providers/delete"
	RouteLLMServices             = "/llm-services"
	RouteLLMServicesEdit         = "/llm-services/edit/:id"
	RouteLLMServicesDelete       = "/llm-services/delete"
	RouteLLMServicesModels       = "/llm-services/models"
	RouteLLMServiceConfigs       = "/llm-service-configs"
	RouteLLMServiceConfigsEdit   = "/llm-service-configs/edit/:id"
	RouteLLMServiceConfigsDelete = "/llm-service-configs/delete"
	RouteLLMServiceConfigsModels = "/llm-service-configs/models"
	RouteSettings                = "/settings"
	RouteConversations           = "/conversations"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, providersHandler *web.ProvidersHandler, llmServicesHandler *web.LLMServicesHandler, llmServiceConfigsHandler *web.LLMServiceConfigsHandler, settingsHandler *web.SettingsHandler, conversationHandler *web.ConversationHandler) {
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

	// LLM Services
	e.GET(RouteLLMServices, llmServicesHandler.LLMServices)
	e.POST(RouteLLMServices, llmServicesHandler.CreateLLMService)
	e.GET(RouteLLMServicesEdit, llmServicesHandler.EditLLMService)
	e.POST(RouteLLMServicesEdit, llmServicesHandler.UpdateLLMService)
	e.POST(RouteLLMServicesDelete, llmServicesHandler.DeleteLLMService)
	e.GET(RouteLLMServicesModels, llmServicesHandler.GetModels)

	// LLM Service Configurations
	e.GET(RouteLLMServiceConfigs, llmServiceConfigsHandler.LLMServiceConfigs)
	e.POST(RouteLLMServiceConfigs, llmServiceConfigsHandler.CreateLLMServiceConfig)
	e.GET(RouteLLMServiceConfigsEdit, llmServiceConfigsHandler.EditLLMServiceConfig)
	e.POST(RouteLLMServiceConfigsEdit, llmServiceConfigsHandler.UpdateLLMServiceConfig)
	e.POST(RouteLLMServiceConfigsDelete, llmServiceConfigsHandler.DeleteLLMServiceConfig)
	e.GET(RouteLLMServiceConfigsModels, llmServiceConfigsHandler.GetModelsForService)

	// Settings
	e.GET(RouteSettings, settingsHandler.Settings)

	// Conversation
	e.GET(RouteConversations, conversationHandler.Conversation)
}
