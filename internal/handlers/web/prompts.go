package web

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type PromptsHandler struct {
	promptService *services.PromptService
}

func NewPromptsHandler(promptService *services.PromptService) *PromptsHandler {
	return &PromptsHandler{
		promptService: promptService,
	}
}

// Prompts handles GET /prompts - displays the prompts page with form and list
func (h *PromptsHandler) Prompts(c echo.Context) error {
	// Get all prompts to display in the list
	prompts, err := h.promptService.GetAllPrompts(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		// In a real app, you might want to show an error message
		prompts = []store.Prompt{}
	}

	return views.PromptsPage(prompts).Render(c.Request().Context(), c.Response().Writer)
}

// CreatePrompt handles POST /prompts - processes form submission
func (h *PromptsHandler) CreatePrompt(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	title := c.FormValue("title")
	content := c.FormValue("content")
	isSystemStr := c.FormValue("is_system")

	// Validate required fields
	if title == "" || content == "" {
		// Redirect back to prompts page with error
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Title and content are required")
	}

	// Convert is_system checkbox to boolean
	isSystem := isSystemStr == "1"

	// Create the prompt (for now without user_id, we'll add that later with sessions)
	if err := h.promptService.CreatePrompt(c.Request().Context(), nil, title, content, isSystem); err != nil {
		// Redirect back with error
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Failed to create prompt")
	}

	// Redirect back to prompts page with success message
	return c.Redirect(http.StatusSeeOther, "/prompts?success=Prompt created successfully")
}

// DeletePrompt handles POST /prompts/delete - deletes a prompt
func (h *PromptsHandler) DeletePrompt(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract prompt ID
	idStr := c.FormValue("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Prompt ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Invalid prompt ID")
	}

	// Delete the prompt
	if err := h.promptService.DeletePrompt(c.Request().Context(), id); err != nil {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Failed to delete prompt")
	}

	// Redirect back to prompts page with success message
	return c.Redirect(http.StatusSeeOther, "/prompts?success=Prompt deleted successfully")
}
