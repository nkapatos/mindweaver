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
	currentPath := c.Path()

	// Get all LLM services for display
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		llmServices = []store.LlmService{}
	}

	// Get available adapters for the form dropdown
	adapters := []string{"openai", "anthropic", "openrouter", "lmstudio", "ollama"}

	return views.LLMServicesPage(llmServices, nil, adapters, currentPath).Render(c.Request().Context(), c.Response().Writer)
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
	configuration := c.FormValue("configuration")

	// Validate required fields
	if name == "" || adapter == "" || apiKey == "" || baseURL == "" || configuration == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Name, adapter, API key, base URL and configuration are required")
	}

	// Create the LLM service
	_, err := h.llmService.CreateLLMService(c.Request().Context(), name, description, adapter, apiKey, baseURL, organization, configuration)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services?error=Failed to create LLM service")
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
	currentPath := c.Path()
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

	return views.LLMServicesPage(llmServices, llmService, adapters, currentPath).Render(c.Request().Context(), c.Response().Writer)
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
	configuration := c.FormValue("configuration")

	// Validate required fields
	if name == "" || adapter == "" || apiKey == "" || baseURL == "" || configuration == "" {
		return c.Redirect(http.StatusSeeOther, "/llm-services/edit/"+idStr+"?error=Name, adapter, API key, base URL and configuration are required")
	}

	// Update the LLM service
	if err := h.llmService.UpdateLLMService(c.Request().Context(), id, name, description, adapter, apiKey, baseURL, organization, configuration); err != nil {
		return c.Redirect(http.StatusSeeOther, "/llm-services/edit/"+idStr+"?error=Failed to update LLM service")
	}

	// Redirect back to LLM services page with success message
	return c.Redirect(http.StatusSeeOther, "/llm-services?success=LLM service updated successfully")
}
