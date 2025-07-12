package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/handlers/api"
)

// SetupAPIRoutes configures all API routes
func SetupAPIRoutes(
	e *echo.Echo,
	actorHandler *api.ActorHandler,
	promptHandler *api.PromptHandler,
	llmHandler *api.LLMHandler,
	conversationHandler *api.ConversationHandler,
	providerHandler *api.ProvidersHandler,
	llmServiceHandler *api.LLMServicesHandler,
	llmServiceConfigHandler *api.LLMServiceConfigsHandler,
	modelsHandler *api.ModelsHandler,
) {
	// API routes
	e.POST("/api/actors", func(c echo.Context) error {
		actorHandler.CreateActor(c.Response().Writer, c.Request())
		return nil
	})
	e.POST("/api/prompts", func(c echo.Context) error {
		promptHandler.CreatePrompt(c.Response().Writer, c.Request())
		return nil
	})

	// Provider routes - only if handler is provided
	if providerHandler != nil {
		e.POST("/api/providers", providerHandler.CreateProvider)
		e.GET("/api/providers", providerHandler.GetAllProviders)
		e.GET("/api/providers/:id", providerHandler.GetProvider)
		e.PUT("/api/providers/:id", providerHandler.UpdateProvider)
		e.DELETE("/api/providers/:id", providerHandler.DeleteProvider)
		e.GET("/api/providers/by-llm-service-config/:config_id", providerHandler.GetProvidersByLLMServiceConfig)
		e.GET("/api/providers/by-llm-service/:service_id", providerHandler.GetProvidersByLLMService)
		e.GET("/api/providers/by-system-prompt/:prompt_id", providerHandler.GetProvidersBySystemPrompt)
	}

	// LLM Service routes - only if handler is provided
	if llmServiceHandler != nil {
		e.POST("/api/llm-services", llmServiceHandler.CreateLLMService)
		e.GET("/api/llm-services", llmServiceHandler.GetAllLLMServices)
		e.GET("/api/llm-services/:id", llmServiceHandler.GetLLMService)
		e.PUT("/api/llm-services/:id", llmServiceHandler.UpdateLLMService)
		e.DELETE("/api/llm-services/:id", llmServiceHandler.DeleteLLMService)
		e.GET("/api/llm-services/:id/configs", llmServiceHandler.GetLLMServiceConfigs)
	}

	// LLM Service Config routes - only if handler is provided
	if llmServiceConfigHandler != nil {
		e.POST("/api/llm-service-configs", llmServiceConfigHandler.CreateLLMServiceConfig)
		e.GET("/api/llm-service-configs", llmServiceConfigHandler.GetAllLLMServiceConfigs)
		e.GET("/api/llm-service-configs/:id", llmServiceConfigHandler.GetLLMServiceConfig)
		e.PUT("/api/llm-service-configs/:id", llmServiceConfigHandler.UpdateLLMServiceConfig)
		e.DELETE("/api/llm-service-configs/:id", llmServiceConfigHandler.DeleteLLMServiceConfig)
	}

	// Model routes - only if handler is provided
	if modelsHandler != nil {
		e.GET("/api/llm-services/:service_id/models", modelsHandler.GetModels)
		e.GET("/api/llm-services/:service_id/models/:model_id", modelsHandler.GetModel)
		e.POST("/api/llm-services/:service_id/models", modelsHandler.CreateModel)
		e.PUT("/api/llm-services/:service_id/models/:model_id", modelsHandler.UpdateModel)
		e.DELETE("/api/llm-services/:service_id/models/:model_id", modelsHandler.DeleteModel)
		e.POST("/api/llm-services/:service_id/models/refresh", modelsHandler.RefreshModels)
	}

	// LLM routes - only if handler is provided
	if llmHandler != nil {
		e.POST("/api/generate", llmHandler.Generate)
		// e.POST("/api/chat", llmHandler.Chat)
	}

	// Conversation routes - only if handler is provided
	if conversationHandler != nil {
		// Conversation CRUD
		e.POST("/api/conversations", conversationHandler.CreateConversation)
		e.GET("/api/conversations", conversationHandler.GetConversations)
		e.GET("/api/conversations/:id", conversationHandler.GetConversation)

		// Message CRUD
		e.POST("/api/conversations/:id/messages", conversationHandler.CreateMessage)
		e.GET("/api/conversations/:id/messages", conversationHandler.GetMessages)
		e.PUT("/api/messages/:id", conversationHandler.UpdateMessage)
		e.DELETE("/api/messages/:id", conversationHandler.DeleteMessage)
	}
}
