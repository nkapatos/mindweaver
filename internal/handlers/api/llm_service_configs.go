package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type LLMServiceConfigsHandler struct {
	llmService *services.LLMService
}

// NewLLMServiceConfigsHandler creates a new LLMServiceConfigsHandler with dependency injection
func NewLLMServiceConfigsHandler(llmService *services.LLMService) *LLMServiceConfigsHandler {
	return &LLMServiceConfigsHandler{
		llmService: llmService,
	}
}

// CreateLLMServiceConfig handles POST /api/llm-service-configs
func (h *LLMServiceConfigsHandler) CreateLLMServiceConfig(c echo.Context) error {
	var req struct {
		LlmServiceID  int64  `json:"llm_service_id"`
		Name          string `json:"name"`
		Description   string `json:"description"`
		Configuration struct {
			Model            string  `json:"model"`
			Temperature      float64 `json:"temperature"`
			MaxTokens        int     `json:"max_tokens"`
			TopP             float64 `json:"top_p"`
			FrequencyPenalty float64 `json:"frequency_penalty"`
			PresencePenalty  float64 `json:"presence_penalty"`
		} `json:"configuration"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.LlmServiceID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Valid LLM service ID is required"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Configuration.Model == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model is required"})
	}

	// Create configuration object
	config := &services.LLMConfiguration{
		Model: req.Configuration.Model,
	}

	// Convert direct values to pointers
	if req.Configuration.Temperature != 0 {
		config.Temperature = &req.Configuration.Temperature
	}
	if req.Configuration.MaxTokens != 0 {
		config.MaxTokens = &req.Configuration.MaxTokens
	}
	if req.Configuration.TopP != 0 {
		config.TopP = &req.Configuration.TopP
	}
	if req.Configuration.FrequencyPenalty != 0 {
		config.FrequencyPenalty = &req.Configuration.FrequencyPenalty
	}
	if req.Configuration.PresencePenalty != 0 {
		config.PresencePenalty = &req.Configuration.PresencePenalty
	}

	// TODO: Get actual actor ID from authentication/session
	// For now, use system actor ID (1) for audit trail
	systemActorID := int64(1)

	llmServiceConfig, err := h.llmService.CreateLLMServiceConfig(
		c.Request().Context(),
		req.LlmServiceID,
		req.Name,
		req.Description,
		config,
		systemActorID,
		systemActorID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create LLM service config"})
	}

	return c.JSON(http.StatusCreated, llmServiceConfig)
}

// GetLLMServiceConfig handles GET /api/llm-service-configs/{id}
func (h *LLMServiceConfigsHandler) GetLLMServiceConfig(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service config ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service config ID"})
	}

	llmServiceConfig, err := h.llmService.GetLLMServiceConfigByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "LLM service config not found"})
	}

	return c.JSON(http.StatusOK, llmServiceConfig)
}

// GetAllLLMServiceConfigs handles GET /api/llm-service-configs
func (h *LLMServiceConfigsHandler) GetAllLLMServiceConfigs(c echo.Context) error {
	// This would require a new method in the service to get all configs
	// For now, we'll return an error indicating this endpoint needs implementation
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "Get all LLM service configs not implemented"})
}

// UpdateLLMServiceConfig handles PUT /api/llm-service-configs/{id}
func (h *LLMServiceConfigsHandler) UpdateLLMServiceConfig(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service config ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service config ID"})
	}

	var req struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		Configuration struct {
			Model            string  `json:"model"`
			Temperature      float64 `json:"temperature"`
			MaxTokens        int     `json:"max_tokens"`
			TopP             float64 `json:"top_p"`
			FrequencyPenalty float64 `json:"frequency_penalty"`
			PresencePenalty  float64 `json:"presence_penalty"`
		} `json:"configuration"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Configuration.Model == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model is required"})
	}

	// Create configuration object
	config := &services.LLMConfiguration{
		Model: req.Configuration.Model,
	}

	// Convert direct values to pointers
	if req.Configuration.Temperature != 0 {
		config.Temperature = &req.Configuration.Temperature
	}
	if req.Configuration.MaxTokens != 0 {
		config.MaxTokens = &req.Configuration.MaxTokens
	}
	if req.Configuration.TopP != 0 {
		config.TopP = &req.Configuration.TopP
	}
	if req.Configuration.FrequencyPenalty != 0 {
		config.FrequencyPenalty = &req.Configuration.FrequencyPenalty
	}
	if req.Configuration.PresencePenalty != 0 {
		config.PresencePenalty = &req.Configuration.PresencePenalty
	}

	// TODO: Get actual actor ID from authentication/session
	// For now, use system actor ID (1) for audit trail
	systemActorID := int64(1)

	if err := h.llmService.UpdateLLMServiceConfig(
		c.Request().Context(),
		id,
		req.Name,
		req.Description,
		config,
		systemActorID,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update LLM service config"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "LLM service config updated successfully"})
}

// DeleteLLMServiceConfig handles DELETE /api/llm-service-configs/{id}
func (h *LLMServiceConfigsHandler) DeleteLLMServiceConfig(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service config ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service config ID"})
	}

	if err := h.llmService.DeleteLLMServiceConfig(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete LLM service config"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "LLM service config deleted successfully"})
}
