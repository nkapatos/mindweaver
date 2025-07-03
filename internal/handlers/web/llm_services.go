package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type LLMServicesHandler struct {
	llmService *services.LLMService
}

func NewLLMServicesHandler(llmService *services.LLMService) *LLMServicesHandler {
	return &LLMServicesHandler{
		llmService: llmService,
	}
}

// LLMServices handles GET /llm-services - displays the LLM services page with form and list
func (h *LLMServicesHandler) LLMServices(c echo.Context) error {

	// Get all LLM services for display
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		llmServices = []store.LlmService{}
	}

	// Get available adapters for the form dropdown
	adapters := []string{"openai", "anthropic", "openrouter", "lmstudio", "ollama"}

	return views.LLMServicesPage(llmServices, nil, adapters).Render(c.Request().Context(), c.Response().Writer)
}

// CreateLLMService handles POST /llm-services - processes form submission
func (h *LLMServicesHandler) CreateLLMService(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	description := c.FormValue("description")
	adapter := c.FormValue("adapter")
	apiKey := c.FormValue("api_key")
	baseURL := c.FormValue("base_url")
	organization := c.FormValue("organization")
	configName := c.FormValue("config_name")
	configDescription := c.FormValue("config_description")
	model := c.FormValue("model")

	// Validate required fields
	if name == "" || adapter == "" || apiKey == "" || baseURL == "" || model == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Name, adapter, API key, base URL and model are required")
	}

	// Use default config name if not provided
	if configName == "" {
		configName = "Default Configuration"
	}

	// Create the LLM service with a default configuration
	_, _, err := h.llmService.CreateLLMServiceWithConfig(c.Request().Context(), name, description, adapter, apiKey, baseURL, organization, configName, configDescription, model)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Failed to create LLM service: "+err.Error())
	}

	// Redirect back to LLM services page with success message
	return c.Redirect(http.StatusSeeOther, "/llm-services?success=LLM service created successfully")
}

// DeleteLLMService handles POST /llm-services/delete - deletes an LLM service
func (h *LLMServicesHandler) DeleteLLMService(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract LLM service ID
	idStr := c.FormValue("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=LLM service ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Invalid LLM service ID")
	}

	// Delete the LLM service
	if err := h.llmService.DeleteLLMService(c.Request().Context(), id); err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Failed to delete LLM service")
	}

	// Redirect back to LLM services page with success message
	return c.Redirect(http.StatusSeeOther, "/llm-services?success=LLM service deleted successfully")
}

// EditLLMService handles GET /llm-services/edit/{id} - shows edit form
func (h *LLMServicesHandler) EditLLMService(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=LLM service ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Invalid LLM service ID")
	}

	// Get the LLM service to edit
	llmService, err := h.llmService.GetLLMServiceByID(c.Request().Context(), id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=LLM service not found")
	}

	// Get all LLM services for display
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get available adapters for the form dropdown
	adapters := []string{"openai", "anthropic", "openrouter", "lmstudio", "ollama"}

	return views.LLMServicesPage(llmServices, llmService, adapters).Render(c.Request().Context(), c.Response().Writer)
}

// GetModels handles GET /llm-services/models - fetches available models for an adapter
func (h *LLMServicesHandler) GetModels(c echo.Context) error {
	// Get query parameters
	adapter := c.QueryParam("adapter")
	apiKey := c.QueryParam("api_key")
	baseURL := c.QueryParam("base_url")

	// Validate required parameters
	if adapter == "" || apiKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "adapter and api_key are required",
		})
	}

	// Fetch available models
	models, err := h.llmService.GetAvailableModels(c.Request().Context(), adapter, apiKey, baseURL)
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

// UpdateLLMService handles POST /llm-services/edit/{id} - processes edit form submission
func (h *LLMServicesHandler) UpdateLLMService(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=LLM service ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Invalid LLM service ID")
	}

	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	description := c.FormValue("description")
	adapter := c.FormValue("adapter")
	apiKey := c.FormValue("api_key")
	baseURL := c.FormValue("base_url")
	organization := c.FormValue("organization")

	// Validate required fields
	if name == "" || adapter == "" || apiKey == "" || baseURL == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services/edit/"+idStr+"?error=Name, adapter, API key, and base URL are required")
	}

	// Update the LLM service
	if err := h.llmService.UpdateLLMService(c.Request().Context(), id, name, description, adapter, apiKey, baseURL, organization); err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services/edit/"+idStr+"?error=Failed to update LLM service: "+err.Error())
	}

	// Redirect back to LLM services page with success message
	return c.Redirect(http.StatusSeeOther, "/llm-services?success=LLM service updated successfully")
}
