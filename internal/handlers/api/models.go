package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type ModelsHandler struct {
	llmService *services.LLMService
}

// NewModelsHandler creates a new ModelsHandler with dependency injection
func NewModelsHandler(llmService *services.LLMService) *ModelsHandler {
	return &ModelsHandler{
		llmService: llmService,
	}
}

// GetModels handles GET /api/llm-services/{service_id}/models
func (h *ModelsHandler) GetModels(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	// Get models for the service
	models, err := h.llmService.GetCachedModels(c.Request().Context(), serviceID, false)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch models"})
	}

	return c.JSON(http.StatusOK, models)
}

// GetModel handles GET /api/llm-services/{service_id}/models/{model_id}
func (h *ModelsHandler) GetModel(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	modelID := c.Param("model_id")

	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	if modelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model ID is required"})
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	// Get model by service and model ID
	model, err := h.llmService.GetModelByServiceAndModelID(c.Request().Context(), serviceID, modelID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Model not found"})
	}

	return c.JSON(http.StatusOK, model)
}

// RefreshModels handles POST /api/llm-services/{service_id}/models/refresh
func (h *ModelsHandler) RefreshModels(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	// Refresh models for the service
	models, err := h.llmService.RefreshModelsForService(c.Request().Context(), serviceID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to refresh models"})
	}

	return c.JSON(http.StatusOK, models)
}

// CreateModel handles POST /api/llm-services/{service_id}/models
func (h *ModelsHandler) CreateModel(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	_, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	var req struct {
		ModelID     string `json:"model_id"`
		Name        string `json:"name"`
		Provider    string `json:"provider"`
		Description string `json:"description"`
		CreatedAt   *int64 `json:"created_at,omitempty"`
		OwnedBy     string `json:"owned_by,omitempty"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.ModelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model ID is required"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Provider is required"})
	}

	// Models should be created through refresh operations
	return c.JSON(http.StatusMethodNotAllowed, map[string]string{"error": "Models should be created through refresh operations"})
}

// UpdateModel handles PUT /api/llm-services/{service_id}/models/{model_id}
func (h *ModelsHandler) UpdateModel(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	modelID := c.Param("model_id")

	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	if modelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model ID is required"})
	}

	_, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	var req struct {
		Name        string `json:"name"`
		Provider    string `json:"provider"`
		Description string `json:"description"`
		CreatedAt   *int64 `json:"created_at,omitempty"`
		OwnedBy     string `json:"owned_by,omitempty"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Provider is required"})
	}

	// Models should be updated through refresh operations
	return c.JSON(http.StatusMethodNotAllowed, map[string]string{"error": "Models should be updated through refresh operations"})
}

// DeleteModel handles DELETE /api/llm-services/{service_id}/models/{model_id}
func (h *ModelsHandler) DeleteModel(c echo.Context) error {
	serviceIDStr := c.Param("service_id")
	modelID := c.Param("model_id")

	if serviceIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	if modelID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Model ID is required"})
	}

	_, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	// Models should be deleted through refresh operations
	return c.JSON(http.StatusMethodNotAllowed, map[string]string{"error": "Models should be deleted through refresh operations"})
}
