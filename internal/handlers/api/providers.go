package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type ProvidersHandler struct {
	providerService *services.ProviderService
}

// NewProvidersHandler creates a new ProvidersHandler with dependency injection
func NewProvidersHandler(providerService *services.ProviderService) *ProvidersHandler {
	return &ProvidersHandler{
		providerService: providerService,
	}
}

// CreateProvider handles POST /api/providers
func (h *ProvidersHandler) CreateProvider(c echo.Context) error {
	var req struct {
		Name               string `json:"name"`
		Description        string `json:"description"`
		LlmServiceConfigID int64  `json:"llm_service_config_id"`
		SystemPromptID     *int64 `json:"system_prompt_id,omitempty"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.LlmServiceConfigID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Valid LLM service config ID is required"})
	}

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	provider, err := h.providerService.CreateProvider(
		c.Request().Context(),
		req.Name,
		req.Description,
		req.LlmServiceConfigID,
		req.SystemPromptID,
		actorID,
		actorID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create provider"})
	}

	return c.JSON(http.StatusCreated, provider)
}

// GetProvider handles GET /api/providers/{id}
func (h *ProvidersHandler) GetProvider(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Provider ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid provider ID"})
	}

	provider, err := h.providerService.GetProviderByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Provider not found"})
	}

	return c.JSON(http.StatusOK, provider)
}

// GetAllProviders handles GET /api/providers
func (h *ProvidersHandler) GetAllProviders(c echo.Context) error {
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch providers"})
	}

	return c.JSON(http.StatusOK, providers)
}

// UpdateProvider handles PUT /api/providers/{id}
func (h *ProvidersHandler) UpdateProvider(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Provider ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid provider ID"})
	}

	var req struct {
		Name               string `json:"name"`
		Description        string `json:"description"`
		LlmServiceConfigID int64  `json:"llm_service_config_id"`
		SystemPromptID     *int64 `json:"system_prompt_id,omitempty"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.LlmServiceConfigID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Valid LLM service config ID is required"})
	}

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	if err := h.providerService.UpdateProvider(
		c.Request().Context(),
		id,
		req.Name,
		req.Description,
		req.LlmServiceConfigID,
		req.SystemPromptID,
		actorID,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update provider"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Provider updated successfully"})
}

// DeleteProvider handles DELETE /api/providers/{id}
func (h *ProvidersHandler) DeleteProvider(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Provider ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid provider ID"})
	}

	if err := h.providerService.DeleteProvider(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete provider"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Provider deleted successfully"})
}

// GetProvidersByLLMServiceConfig handles GET /api/providers/by-llm-service-config/{config_id}
func (h *ProvidersHandler) GetProvidersByLLMServiceConfig(c echo.Context) error {
	configIDStr := c.Param("config_id")
	if configIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service config ID is required"})
	}

	configID, err := strconv.ParseInt(configIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service config ID"})
	}

	providers, err := h.providerService.GetProvidersByLLMServiceConfig(c.Request().Context(), configID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch providers"})
	}

	return c.JSON(http.StatusOK, providers)
}

// GetProvidersByLLMService handles GET /api/providers/by-llm-service/{service_id}
func (h *ProvidersHandler) GetProvidersByLLMService(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	providers, err := h.providerService.GetProvidersByLLMService(c.Request().Context(), serviceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch providers"})
	}

	return c.JSON(http.StatusOK, providers)
}

// GetProvidersBySystemPrompt handles GET /api/providers/by-system-prompt/{prompt_id}
func (h *ProvidersHandler) GetProvidersBySystemPrompt(c echo.Context) error {
	promptIDStr := c.Param("prompt_id")
	if promptIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "System prompt ID is required"})
	}

	promptID, err := strconv.ParseInt(promptIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid system prompt ID"})
	}

	providers, err := h.providerService.GetProvidersBySystemPrompt(c.Request().Context(), promptID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch providers"})
	}

	return c.JSON(http.StatusOK, providers)
}
