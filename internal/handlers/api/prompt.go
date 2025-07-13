package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type PromptHandler struct {
	promptService *services.PromptService
}

// NewPromptHandler creates a new PromptHandler with dependency injection
func NewPromptHandler(promptService *services.PromptService) *PromptHandler {
	return &PromptHandler{
		promptService: promptService,
	}
}

// CreatePrompt handles POST /api/prompts
func (h *PromptHandler) CreatePrompt(c echo.Context) error {
	var req struct {
		CreatedBy *int64 `json:"created_by,omitempty"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		IsSystem  bool   `json:"is_system"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Title == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title and content are required"})
	}

	// Get actor ID from session
	sess, _ := session.Get("session", c)
	createdBy, _ := sess.Values["actor_id"].(int64)

	if err := h.promptService.CreatePrompt(c.Request().Context(), req.Title, req.Content, req.IsSystem, createdBy, createdBy); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create prompt"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Prompt created successfully"})
}

// GetPrompt handles GET /api/prompts/{id}
func (h *PromptHandler) GetPrompt(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Prompt ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid prompt ID"})
	}

	prompt, err := h.promptService.GetPromptByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Prompt not found"})
	}

	return c.JSON(http.StatusOK, prompt)
}

// GetAllPrompts handles GET /api/prompts
func (h *PromptHandler) GetAllPrompts(c echo.Context) error {
	prompts, err := h.promptService.GetAllPrompts(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch prompts"})
	}

	return c.JSON(http.StatusOK, prompts)
}

// GetPromptsByCreatedBy handles GET /api/prompts/by-actor?actor_id={id}
func (h *PromptHandler) GetPromptsByCreatedBy(c echo.Context) error {
	// Get the actor ID from query parameter
	createdByStr := c.QueryParam("actor_id")
	if createdByStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "actor_id parameter is required"})
	}

	createdBy, err := strconv.ParseInt(createdByStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid actor_id parameter"})
	}

	prompts, err := h.promptService.GetPromptsByCreatedBy(c.Request().Context(), createdBy)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch prompts"})
	}

	return c.JSON(http.StatusOK, prompts)
}

// GetSystemPrompts handles GET /api/prompts/system
func (h *PromptHandler) GetSystemPrompts(c echo.Context) error {
	prompts, err := h.promptService.GetSystemPrompts(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch system prompts"})
	}

	return c.JSON(http.StatusOK, prompts)
}

// UpdatePrompt handles PUT /api/prompts/{id}
func (h *PromptHandler) UpdatePrompt(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Prompt ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid prompt ID"})
	}

	var req struct {
		CreatedBy *int64 `json:"created_by,omitempty"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		IsSystem  bool   `json:"is_system"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Title == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Title and content are required"})
	}

	// Get actor ID from session
	sess, _ := session.Get("session", c)
	updatedBy, _ := sess.Values["actor_id"].(int64)

	if err := h.promptService.UpdatePrompt(c.Request().Context(), id, req.CreatedBy, req.Title, req.Content, req.IsSystem, updatedBy); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update prompt"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Prompt updated successfully"})
}

// DeletePrompt handles DELETE /api/prompts/{id}
func (h *PromptHandler) DeletePrompt(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Prompt ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid prompt ID"})
	}

	if err := h.promptService.DeletePrompt(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete prompt"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Prompt deleted successfully"})
}
