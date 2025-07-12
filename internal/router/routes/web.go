package routes

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/config"
	"github.com/nkapatos/mindweaver/internal/handlers/web"
)

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, authHandler *web.AuthHandler, authMiddleware echo.MiddlewareFunc, homeHandler *web.HomeHandler, promptsHandler *web.PromptsHandler, providersHandler *web.ProvidersHandler, llmServicesHandler *web.LLMServicesHandler, llmServiceConfigsHandler *web.LLMServiceConfigsHandler, settingsHandler *web.SettingsHandler, conversationHandler *web.ConversationHandler, setupHandler *web.SetupHandler) {
	// Setup routes (no authentication required)
	e.GET("/setup", setupHandler.SetupPage)
	e.POST("/setup", setupHandler.SetupApplication)

	// Auth routes (no authentication required)
	e.GET(config.RouteAuthSignIn, authHandler.SignInPage)
	e.POST(config.RouteAuthSignIn, authHandler.SignIn)
	e.POST("/auth/logout", authHandler.Logout)

	// Protected routes (require authentication)
	protected := e.Group("")
	protected.Use(authMiddleware)

	// Home (requires authentication)
	protected.GET(config.RouteHome, homeHandler.Home)

	// Prompts group (requires authentication)
	prompts := protected.Group(config.RoutePrompts)
	prompts.GET("", promptsHandler.Prompts)
	prompts.GET(fmt.Sprintf("/%s", config.RESTActionNew), promptsHandler.NewPrompt)
	prompts.POST(fmt.Sprintf("/%s", config.RESTActionCreate), promptsHandler.CreatePrompt)
	prompts.GET(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), promptsHandler.EditPrompt)
	prompts.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), promptsHandler.UpdatePrompt)
	prompts.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionDelete), promptsHandler.DeletePrompt)

	// Providers group (requires authentication)
	providers := protected.Group(config.RouteProviders)
	providers.GET("", providersHandler.Providers)
	providers.GET(fmt.Sprintf("/%s", config.RESTActionNew), providersHandler.NewProvider)
	providers.POST(fmt.Sprintf("/%s", config.RESTActionCreate), providersHandler.CreateProvider)
	providers.GET(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), providersHandler.EditProvider)
	providers.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), providersHandler.UpdateProvider)
	providers.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionDelete), providersHandler.DeleteProvider)

	// LLM Services group (requires authentication)
	llmServices := protected.Group(config.RouteLLMServices)
	llmServices.GET("", llmServicesHandler.LLMServices)
	llmServices.GET(fmt.Sprintf("/%s", config.RESTActionNew), llmServicesHandler.NewLLMService)
	llmServices.POST(fmt.Sprintf("/%s", config.RESTActionCreate), llmServicesHandler.CreateLLMService)
	llmServices.GET(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), llmServicesHandler.EditLLMService)
	llmServices.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), llmServicesHandler.UpdateLLMService)
	llmServices.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionDelete), llmServicesHandler.DeleteLLMService)
	llmServices.GET("/:service-id/models", llmServicesHandler.GetModels)
	llmServices.POST("/test-connection", llmServicesHandler.TestConnection)

	// LLM Service Configurations group (requires authentication)
	configs := protected.Group(config.RouteLLMServiceConfigs)
	configs.GET("", llmServiceConfigsHandler.LLMServiceConfigs)
	configs.GET(fmt.Sprintf("/%s", config.RESTActionNew), llmServiceConfigsHandler.NewLLMServiceConfig)
	configs.POST(fmt.Sprintf("/%s", config.RESTActionCreate), llmServiceConfigsHandler.CreateLLMServiceConfig)
	configs.GET(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), llmServiceConfigsHandler.EditLLMServiceConfig)
	configs.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionEdit), llmServiceConfigsHandler.UpdateLLMServiceConfig)
	configs.POST(fmt.Sprintf("/%s/%s", ":id", config.RESTActionDelete), llmServiceConfigsHandler.DeleteLLMServiceConfig)
	configs.GET(fmt.Sprintf("/%s/models", ":id"), llmServiceConfigsHandler.GetModelsForService)

	// Settings (requires authentication)
	protected.GET(config.RouteSettings, settingsHandler.Settings)

	// Conversations group (requires authentication)
	conversations := protected.Group(config.RouteConversations)
	conversations.GET("", conversationHandler.Conversation)
	conversations.GET(fmt.Sprintf("/%s", config.RESTActionNew), conversationHandler.NewConversation)
	conversations.POST(fmt.Sprintf("/%s", config.RESTActionCreate), conversationHandler.CreateConversation)
	conversations.GET(fmt.Sprintf("/%s", ":id"), conversationHandler.ViewConversation)
}
