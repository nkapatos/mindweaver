package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
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
	// Get all providers with relations for display
	providersWithRelations, err := h.providerService.GetAllProvidersWithRelations(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		providersWithRelations = []struct {
			Provider         store.Provider
			LLMServiceConfig store.LlmServiceConfig
			LLMService       store.LlmService
			SystemPrompt     *store.Prompt
		}{}
	}

	// Convert to template format
	var templateProviders []views.ProviderWithRelations
	for _, p := range providersWithRelations {
		templateProviders = append(templateProviders, views.ProviderWithRelations{
			Provider:         p.Provider,
			LLMServiceConfig: p.LLMServiceConfig,
			LLMService:       p.LLMService,
			SystemPrompt:     p.SystemPrompt,
		})
	}

	return views.ProvidersList(templateProviders).Render(c.Request().Context(), c.Response().Writer)
}

// NewProvider handles GET /providers/new - shows create form
func (h *ProvidersHandler) NewProvider(c echo.Context) error {
	// Get all LLM services to get their configs
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get all configs from all services
	var configs []store.LlmServiceConfig
	for _, service := range llmServices {
		serviceConfigs, err := h.llmService.GetLLMServiceConfigsByServiceID(c.Request().Context(), service.ID)
		if err != nil {
			continue
		}
		configs = append(configs, serviceConfigs...)
	}

	// Get all system prompts for the dropdown
	prompts, err := h.promptService.GetSystemPrompts(c.Request().Context())
	if err != nil {
		prompts = []store.Prompt{}
	}

	return views.ProviderDetailsForm(nil, configs, prompts).Render(c.Request().Context(), c.Response().Writer)
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

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	// Create the provider
	_, err = h.providerService.CreateProvider(c.Request().Context(), name, description, llmServiceConfigID, systemPromptID, actorID, actorID)
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

	// Get all LLM services to get their configs
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		llmServices = []store.LlmService{}
	}

	// Get all configs from all services
	var configs []store.LlmServiceConfig
	for _, service := range llmServices {
		serviceConfigs, err := h.llmService.GetLLMServiceConfigsByServiceID(c.Request().Context(), service.ID)
		if err != nil {
			continue
		}
		configs = append(configs, serviceConfigs...)
	}

	// Get all system prompts for the dropdown
	prompts, err := h.promptService.GetSystemPrompts(c.Request().Context())
	if err != nil {
		prompts = []store.Prompt{}
	}

	// Get all providers with relations for display
	providersWithRelations, err := h.providerService.GetAllProvidersWithRelations(c.Request().Context())
	if err != nil {
		providersWithRelations = []struct {
			Provider         store.Provider
			LLMServiceConfig store.LlmServiceConfig
			LLMService       store.LlmService
			SystemPrompt     *store.Prompt
		}{}
	}

	// Convert to template format
	var templateProviders []views.ProviderWithRelations
	for _, p := range providersWithRelations {
		templateProviders = append(templateProviders, views.ProviderWithRelations{
			Provider:         p.Provider,
			LLMServiceConfig: p.LLMServiceConfig,
			LLMService:       p.LLMService,
			SystemPrompt:     p.SystemPrompt,
		})
	}

	return views.ProviderDetailsForm(provider, configs, prompts).Render(c.Request().Context(), c.Response().Writer)
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

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	// Update the provider
	if err := h.providerService.UpdateProvider(c.Request().Context(), id, name, description, llmServiceConfigID, systemPromptID, actorID); err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Failed to update provider")
	}

	// Redirect back to providers page with success message
	return c.Redirect(http.StatusSeeOther, "/providers?success=Provider updated successfully")
}
