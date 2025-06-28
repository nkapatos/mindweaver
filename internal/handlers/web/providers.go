package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type ProvidersHandler struct {
	providerService *services.ProviderService
	llmService      *services.LLMService
	promptService   *services.PromptService
}

func NewProvidersHandler(providerService *services.ProviderService, llmService *services.LLMService, promptService *services.PromptService) *ProvidersHandler {
	return &ProvidersHandler{
		providerService: providerService,
		llmService:      llmService,
		promptService:   promptService,
	}
}

// Providers handles GET /providers - displays the providers page with form and list
func (h *ProvidersHandler) Providers(c echo.Context) error {
	currentPath := c.Path()

	// Get all providers with relations for display
	providersWithRelations, err := h.providerService.GetAllProvidersWithRelations(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		providersWithRelations = []struct {
			Provider     store.Provider
			LLMService   store.LlmService
			SystemPrompt *store.Prompt
		}{}
	}

	// Get all LLM services for the form dropdown
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get all system prompts for the form dropdown
	systemPrompts, err := h.promptService.GetSystemPrompts(c.Request().Context())
	if err != nil {
		systemPrompts = []store.Prompt{}
	}

	// Convert to template format
	var templateProviders []views.ProviderWithRelations
	for _, p := range providersWithRelations {
		templateProviders = append(templateProviders, views.ProviderWithRelations{
			Provider:     p.Provider,
			LLMService:   p.LLMService,
			SystemPrompt: p.SystemPrompt,
		})
	}

	return views.ProvidersPage(templateProviders, nil, llmServices, systemPrompts, currentPath).Render(c.Request().Context(), c.Response().Writer)
}

// CreateProvider handles POST /providers - processes form submission
func (h *ProvidersHandler) CreateProvider(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	description := c.FormValue("description")
	llmServiceIDStr := c.FormValue("llm_service_id")
	systemPromptIDStr := c.FormValue("system_prompt_id")

	// Validate required fields
	if name == "" || description == "" || llmServiceIDStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Name, description and LLM service are required")
	}

	// Parse LLM service ID
	llmServiceID, err := strconv.ParseInt(llmServiceIDStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Invalid LLM service ID")
	}

	// Parse optional system prompt ID
	var systemPromptID *int64
	if systemPromptIDStr != "" {
		id, err := strconv.ParseInt(systemPromptIDStr, 10, 64)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/providers?error=Invalid system prompt ID")
		}
		systemPromptID = &id
	}

	// Create the provider
	_, err = h.providerService.CreateProvider(c.Request().Context(), name, description, llmServiceID, systemPromptID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Failed to create provider")
	}

	// Redirect back to providers page with success message
	return c.Redirect(http.StatusSeeOther, "/providers?success=Provider created successfully")
}

// DeleteProvider handles POST /providers/delete - deletes a provider
func (h *ProvidersHandler) DeleteProvider(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract provider ID
	idStr := c.FormValue("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Provider ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Invalid provider ID")
	}

	// Delete the provider
	if err := h.providerService.DeleteProvider(c.Request().Context(), id); err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Failed to delete provider")
	}

	// Redirect back to providers page with success message
	return c.Redirect(http.StatusSeeOther, "/providers?success=Provider deleted successfully")
}

// EditProvider handles GET /providers/edit/{id} - shows edit form
func (h *ProvidersHandler) EditProvider(c echo.Context) error {
	currentPath := c.Path()
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Provider ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Invalid provider ID")
	}

	// Get the provider to edit
	provider, err := h.providerService.GetProviderByID(c.Request().Context(), id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Provider not found")
	}

	// Get all providers with relations for display
	providersWithRelations, err := h.providerService.GetAllProvidersWithRelations(c.Request().Context())
	if err != nil {
		providersWithRelations = []struct {
			Provider     store.Provider
			LLMService   store.LlmService
			SystemPrompt *store.Prompt
		}{}
	}

	// Get all LLM services for the form dropdown
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get all system prompts for the form dropdown
	systemPrompts, err := h.promptService.GetSystemPrompts(c.Request().Context())
	if err != nil {
		systemPrompts = []store.Prompt{}
	}

	// Convert to template format
	var templateProviders []views.ProviderWithRelations
	for _, p := range providersWithRelations {
		templateProviders = append(templateProviders, views.ProviderWithRelations{
			Provider:     p.Provider,
			LLMService:   p.LLMService,
			SystemPrompt: p.SystemPrompt,
		})
	}

	return views.ProvidersPage(templateProviders, provider, llmServices, systemPrompts, currentPath).Render(c.Request().Context(), c.Response().Writer)
}

// UpdateProvider handles POST /providers/edit/{id} - processes edit form submission
func (h *ProvidersHandler) UpdateProvider(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Provider ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers?error=Invalid provider ID")
	}

	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	description := c.FormValue("description")
	llmServiceIDStr := c.FormValue("llm_service_id")
	systemPromptIDStr := c.FormValue("system_prompt_id")

	// Validate required fields
	if name == "" || description == "" || llmServiceIDStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Name, description and LLM service are required")
	}

	// Parse LLM service ID
	llmServiceID, err := strconv.ParseInt(llmServiceIDStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Invalid LLM service ID")
	}

	// Parse optional system prompt ID
	var systemPromptID *int64
	if systemPromptIDStr != "" {
		promptID, err := strconv.ParseInt(systemPromptIDStr, 10, 64)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Invalid system prompt ID")
		}
		systemPromptID = &promptID
	}

	// Update the provider
	if err := h.providerService.UpdateProvider(c.Request().Context(), id, name, description, llmServiceID, systemPromptID); err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Failed to update provider")
	}

	// Redirect back to providers page with success message
	return c.Redirect(http.StatusSeeOther, "/providers?success=Provider updated successfully")
}
