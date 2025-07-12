package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type LLMServiceConfigsHandler struct {
	llmService *services.LLMService
}

func NewLLMServiceConfigsHandler(llmService *services.LLMService) *LLMServiceConfigsHandler {
	return &LLMServiceConfigsHandler{
		llmService: llmService,
	}
}

// LLMServiceConfigs handles GET /llm-service-configs - displays the configurations page
func (h *LLMServiceConfigsHandler) LLMServiceConfigs(c echo.Context) error {

	// Get all LLM services for the service selection dropdown
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get all configurations with their service info for display
	var configsWithServices []views.LLMServiceConfigWithService
	for _, service := range llmServices {
		configs, err := h.llmService.GetLLMServiceConfigsByServiceID(c.Request().Context(), service.ID)
		if err != nil {
			continue // Skip this service if we can't get its configs
		}
		for _, config := range configs {
			configsWithServices = append(configsWithServices, views.LLMServiceConfigWithService{
				LLMServiceConfig: config,
				LLMService:       service,
			})
		}
	}

	return views.LLMServiceConfigsList(configsWithServices).Render(c.Request().Context(), c.Response().Writer)
}

// NewLLMServiceConfig handles GET /llm-service-configs/new - shows create form
func (h *LLMServiceConfigsHandler) NewLLMServiceConfig(c echo.Context) error {
	// Get all LLM services for the service selection dropdown
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get all configurations with their service info for display
	var configsWithServices []views.LLMServiceConfigWithService
	for _, service := range llmServices {
		configs, err := h.llmService.GetLLMServiceConfigsByServiceID(c.Request().Context(), service.ID)
		if err != nil {
			continue // Skip this service if we can't get its configs
		}
		for _, config := range configs {
			configsWithServices = append(configsWithServices, views.LLMServiceConfigWithService{
				LLMServiceConfig: config,
				LLMService:       service,
			})
		}
	}

	return views.LLMServiceConfigsList(configsWithServices).Render(c.Request().Context(), c.Response().Writer)
}

// CreateLLMServiceConfig handles POST /llm-service-configs/create - creates a new configuration
func (h *LLMServiceConfigsHandler) CreateLLMServiceConfig(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	llmServiceIDStr := c.FormValue("llm_service_id")
	name := c.FormValue("name")
	description := c.FormValue("description")
	model := c.FormValue("model")
	temperatureStr := c.FormValue("temperature")
	maxTokensStr := c.FormValue("max_tokens")

	// If this is just a service selection (no name/model), redirect back with service ID
	if llmServiceIDStr != "" && (name == "" || model == "") {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs/new?llm_service_id="+llmServiceIDStr)
	}

	// Validate required fields
	if llmServiceIDStr == "" || name == "" || model == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs/new?error=Service, name, and model are required")
	}

	// Parse LLM service ID
	llmServiceID, err := strconv.ParseInt(llmServiceIDStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs/new?error=Invalid LLM service ID")
	}

	// Parse temperature (optional, default to 0.7)
	temperature := 0.7
	if temperatureStr != "" {
		if temp, err := strconv.ParseFloat(temperatureStr, 64); err == nil {
			temperature = temp
		}
	}

	// Parse max tokens (optional, default to 2000)
	maxTokens := 2000
	if maxTokensStr != "" {
		if tokens, err := strconv.ParseInt(maxTokensStr, 10, 64); err == nil {
			maxTokens = int(tokens)
		}
	}

	// Create configuration with custom values
	config := services.DefaultConfiguration(model)
	if temperature != 0.7 {
		config.Temperature = &temperature
	}
	if maxTokens != 2000 {
		config.MaxTokens = &maxTokens
	}

	// Create the LLM service configuration
	_, err = h.llmService.CreateLLMServiceConfig(c.Request().Context(), llmServiceID, name, description, config, 1, 1) // TODO: Use real actor ID from session
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs/new?error=Failed to create configuration: "+err.Error())
	}

	// Redirect back with success message
	return c.Redirect(http.StatusSeeOther, "/llm-service-configs?success=Configuration created successfully")
}

// GetModelsForService handles GET /llm-service-configs/models - fetches models for a service
func (h *LLMServiceConfigsHandler) GetModelsForService(c echo.Context) error {
	// Get query parameters
	llmServiceIDStr := c.QueryParam("llm_service_id")

	// Validate required parameters
	if llmServiceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "llm_service_id is required",
		})
	}

	// Parse LLM service ID
	llmServiceID, err := strconv.ParseInt(llmServiceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid LLM service ID",
		})
	}

	// Get the LLM service to get its connection details
	llmService, err := h.llmService.GetLLMServiceByID(c.Request().Context(), llmServiceID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "LLM service not found",
		})
	}

	// Fetch available models
	models, err := h.llmService.GetAvailableModels(c.Request().Context(), llmService.Adapter, llmService.ApiKey, llmService.BaseUrl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch models: " + err.Error(),
		})
	}

	// Return models as JSON
	return c.JSON(http.StatusOK, map[string]interface{}{
		"models": models,
	})
}

// EditLLMServiceConfig handles GET /llm-service-configs/{id}/edit - shows edit form
func (h *LLMServiceConfigsHandler) EditLLMServiceConfig(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Configuration ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Invalid configuration ID")
	}

	// Get the configuration to edit
	config, err := h.llmService.GetLLMServiceConfigByID(c.Request().Context(), id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Configuration not found")
	}

	// Get all LLM services for the service selection dropdown
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get models for the service this config belongs to
	var availableModels []views.Model
	models, err := h.llmService.GetAvailableModelsForService(c.Request().Context(), config.LlmServiceID)
	if err == nil {
		// Convert to views.Model format
		for _, model := range models {
			availableModels = append(availableModels, views.Model{
				ID:   model.ID,
				Name: model.Name,
			})
		}
	}

	// Get all configurations with their service info for display
	var configsWithServices []views.LLMServiceConfigWithService
	for _, service := range llmServices {
		configs, err := h.llmService.GetLLMServiceConfigsByServiceID(c.Request().Context(), service.ID)
		if err != nil {
			continue
		}
		for _, cfg := range configs {
			configsWithServices = append(configsWithServices, views.LLMServiceConfigWithService{
				LLMServiceConfig: cfg,
				LLMService:       service,
			})
		}
	}

	return views.LLMServiceConfigDetailsForm(config, llmServices, config.LlmServiceID, availableModels).Render(c.Request().Context(), c.Response().Writer)
}

// UpdateLLMServiceConfig handles POST /llm-service-configs/{id}/edit - updates a configuration
func (h *LLMServiceConfigsHandler) UpdateLLMServiceConfig(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Configuration ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Invalid configuration ID")
	}

	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	description := c.FormValue("description")
	model := c.FormValue("model")
	temperatureStr := c.FormValue("temperature")
	maxTokensStr := c.FormValue("max_tokens")

	// Validate required fields
	if name == "" || model == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs/"+idStr+"/edit?error=Name and model are required")
	}

	// Parse temperature (optional, default to 0.7)
	temperature := 0.7
	if temperatureStr != "" {
		if temp, err := strconv.ParseFloat(temperatureStr, 64); err == nil {
			temperature = temp
		}
	}

	// Parse max tokens (optional, default to 2000)
	maxTokens := 2000
	if maxTokensStr != "" {
		if tokens, err := strconv.ParseInt(maxTokensStr, 10, 64); err == nil {
			maxTokens = int(tokens)
		}
	}

	// Create configuration with custom values
	config := services.DefaultConfiguration(model)
	if temperature != 0.7 {
		config.Temperature = &temperature
	}
	if maxTokens != 2000 {
		config.MaxTokens = &maxTokens
	}

	// Update the configuration
	if err := h.llmService.UpdateLLMServiceConfig(c.Request().Context(), id, name, description, config, 1); err != nil { // TODO: Use real actor ID from session
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs/"+idStr+"/edit?error=Failed to update configuration: "+err.Error())
	}

	// Redirect back with success message
	return c.Redirect(http.StatusSeeOther, "/llm-service-configs?success=Configuration updated successfully")
}

// DeleteLLMServiceConfig handles POST /llm-service-configs/delete - deletes a configuration
func (h *LLMServiceConfigsHandler) DeleteLLMServiceConfig(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract configuration ID
	idStr := c.FormValue("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Configuration ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Invalid configuration ID")
	}

	// Delete the configuration
	if err := h.llmService.DeleteLLMServiceConfig(c.Request().Context(), id); err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-service-configs?error=Failed to delete configuration")
	}

	// Redirect back with success message
	return c.Redirect(http.StatusSeeOther, "/llm-service-configs?success=Configuration deleted successfully")
}
