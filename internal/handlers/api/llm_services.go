package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type LLMServicesHandler struct {
	llmService *services.LLMService
}

// NewLLMServicesHandler creates a new LLMServicesHandler with dependency injection
func NewLLMServicesHandler(llmService *services.LLMService) *LLMServicesHandler {
	return &LLMServicesHandler{
		llmService: llmService,
	}
}

// CreateLLMService handles POST /api/llm-services
func (h *LLMServicesHandler) CreateLLMService(c echo.Context) error {
	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Adapter      string `json:"adapter"`
		ApiKey       string `json:"api_key"`
		BaseURL      string `json:"base_url"`
		Organization string `json:"organization"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Adapter == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Adapter is required"})
	}

	if req.ApiKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "API key is required"})
	}

	// Get actor ID from session
	sess, _ := session.Get("session", c)
	createdBy, _ := sess.Values["actor_id"].(int64)

	llmService, err := h.llmService.CreateLLMService(
		c.Request().Context(),
		req.Name,
		req.Description,
		req.Adapter,
		req.ApiKey,
		req.BaseURL,
		req.Organization,
		createdBy,
		createdBy,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create LLM service"})
	}

	return c.JSON(http.StatusCreated, llmService)
}

// GetLLMService handles GET /api/llm-services/{id}
func (h *LLMServicesHandler) GetLLMService(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	llmService, err := h.llmService.GetLLMServiceByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "LLM service not found"})
	}

	return c.JSON(http.StatusOK, llmService)
}

// GetAllLLMServices handles GET /api/llm-services
func (h *LLMServicesHandler) GetAllLLMServices(c echo.Context) error {
	llmServices, err := h.llmService.GetAllLLMServices(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch LLM services"})
	}

	return c.JSON(http.StatusOK, llmServices)
}

// UpdateLLMService handles PUT /api/llm-services/{id}
func (h *LLMServicesHandler) UpdateLLMService(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Adapter      string `json:"adapter"`
		ApiKey       string `json:"api_key"`
		BaseURL      string `json:"base_url"`
		Organization string `json:"organization"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Adapter == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Adapter is required"})
	}

	if req.ApiKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "API key is required"})
	}

	// Get actor ID from session
	sess, _ := session.Get("session", c)
	updatedBy, _ := sess.Values["actor_id"].(int64)

	if err := h.llmService.UpdateLLMService(
		c.Request().Context(),
		id,
		req.Name,
		req.Description,
		req.Adapter,
		req.ApiKey,
		req.BaseURL,
		req.Organization,
		updatedBy,
	); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update LLM service"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "LLM service updated successfully"})
}

// DeleteLLMService handles DELETE /api/llm-services/{id}
func (h *LLMServicesHandler) DeleteLLMService(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	if err := h.llmService.DeleteLLMService(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete LLM service"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "LLM service deleted successfully"})
}

// GetLLMServiceConfigs handles GET /api/llm-services/{id}/configs
func (h *LLMServicesHandler) GetLLMServiceConfigs(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "LLM service ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid LLM service ID"})
	}

	configs, err := h.llmService.GetLLMServiceConfigsByServiceID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch LLM service configs"})
	}

	return c.JSON(http.StatusOK, configs)
}
