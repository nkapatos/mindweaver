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

// getProvidersWithRelations fetches all providers with their relations
func (h *ProvidersHandler) getProvidersWithRelations(ctx echo.Context) []views.ProviderWithRelations {
	providersWithRelations, err := h.providerService.GetAllProvidersWithRelations(ctx.Request().Context())
	if err != nil {
		return []views.ProviderWithRelations{}
	}

	var templateProviders []views.ProviderWithRelations
	for _, p := range providersWithRelations {
		templateProviders = append(templateProviders, views.ProviderWithRelations{
			Provider:         p.Provider,
			LLMServiceConfig: p.LLMServiceConfig,
			LLMService:       p.LLMService,
			SystemPrompt:     p.SystemPrompt,
		})
	}
	return templateProviders
}

// getAllLLMServiceConfigs fetches all LLM service configs from all services
func (h *ProvidersHandler) getAllLLMServiceConfigs(ctx echo.Context) []store.LlmServiceConfig {
	llmServices, err := h.llmService.GetAllLLMServices(ctx.Request().Context())
	if err != nil {
		return []store.LlmServiceConfig{}
	}

	var allConfigs []store.LlmServiceConfig
	for _, service := range llmServices {
		configs, err := h.llmService.GetLLMServiceConfigsByServiceID(ctx.Request().Context(), service.ID)
		if err != nil {
			continue // Skip this service if we can't get its configs
		}
		allConfigs = append(allConfigs, configs...)
	}
	return allConfigs
}

// getSystemPrompts fetches all system prompts
func (h *ProvidersHandler) getSystemPrompts(ctx echo.Context) []store.Prompt {
	systemPrompts, err := h.promptService.GetSystemPrompts(ctx.Request().Context())
	if err != nil {
		return []store.Prompt{}
	}
	return systemPrompts
}

// Providers handles GET /providers - displays the providers page with list
func (h *ProvidersHandler) Providers(c echo.Context) error {
	providersWithRelations := h.getProvidersWithRelations(c)
	return views.ProvidersList(providersWithRelations).Render(c.Request().Context(), c.Response().Writer)
}

// NewProvider handles GET /providers/new - shows create form
func (h *ProvidersHandler) NewProvider(c echo.Context) error {
	// Get form data
	allConfigs := h.getAllLLMServiceConfigs(c)
	systemPrompts := h.getSystemPrompts(c)

	return views.ProviderDetailsForm(nil, allConfigs, systemPrompts).Render(c.Request().Context(), c.Response().Writer)
}

// CreateProvider handles POST /providers/create - processes form submission
func (h *ProvidersHandler) CreateProvider(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	description := c.FormValue("description")
	llmServiceConfigIDStr := c.FormValue("llm_service_config_id")
	systemPromptIDStr := c.FormValue("system_prompt_id")

	// Validate required fields
	if name == "" || description == "" || llmServiceConfigIDStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers/new?error=Name, description and LLM service configuration are required")
	}

	// Parse LLM service config ID
	llmServiceConfigID, err := strconv.ParseInt(llmServiceConfigIDStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/new?error=Invalid LLM service configuration ID")
	}

	// Parse optional system prompt ID
	var systemPromptID *int64
	if systemPromptIDStr != "" {
		id, err := strconv.ParseInt(systemPromptIDStr, 10, 64)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/providers/new?error=Invalid system prompt ID")
		}
		systemPromptID = &id
	}

	// Create the provider
	_, err = h.providerService.CreateProvider(c.Request().Context(), name, description, llmServiceConfigID, systemPromptID)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/new?error=Failed to create provider")
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

	// Get form data
	allConfigs := h.getAllLLMServiceConfigs(c)
	systemPrompts := h.getSystemPrompts(c)

	return views.ProviderDetailsForm(provider, allConfigs, systemPrompts).Render(c.Request().Context(), c.Response().Writer)
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
	llmServiceConfigIDStr := c.FormValue("llm_service_config_id")
	systemPromptIDStr := c.FormValue("system_prompt_id")

	// Validate required fields
	if name == "" || description == "" || llmServiceConfigIDStr == "" {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Name, description and LLM service configuration are required")
	}

	// Parse LLM service config ID
	llmServiceConfigID, err := strconv.ParseInt(llmServiceConfigIDStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Invalid LLM service configuration ID")
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
	if err := h.providerService.UpdateProvider(c.Request().Context(), id, name, description, llmServiceConfigID, systemPromptID); err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Failed to update provider")
	}

	// Redirect back to providers page with success message
	return c.Redirect(http.StatusSeeOther, "/providers?success=Provider updated successfully")
}
