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

	// Get all prompts with relations for display
	promptsWithRelations, err := h.promptService.GetAllPromptsWithRelations(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		// In a real app, you might want to show an error message
		promptsWithRelations = []struct {
			Prompt store.Prompt
			Actor  *store.Actor
		}{}
	}

	// Convert to template format
	var templatePrompts []views.PromptWithRelations
	for _, p := range promptsWithRelations {
		templatePrompts = append(templatePrompts, views.PromptWithRelations{
			Prompt: p.Prompt,
			Actor:  p.Actor,
		})
	}

	return views.PromptsList(templatePrompts).Render(c.Request().Context(), c.Response().Writer)
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

	// Create the prompt (for now without actor_id, we'll add that later with sessions)
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

// EditPrompt handles GET /prompts/edit/{id} - shows edit form
func (h *PromptsHandler) EditPrompt(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Prompt ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Invalid prompt ID")
	}

	// Get the prompt to edit
	prompt, err := h.promptService.GetPromptByID(c.Request().Context(), id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Prompt not found")
	}

	// Get all prompts with relations for display
	promptsWithRelations, err := h.promptService.GetAllPromptsWithRelations(c.Request().Context())
	if err != nil {
		promptsWithRelations = []struct {
			Prompt store.Prompt
			Actor  *store.Actor
		}{}
	}

	// Convert to template format
	var templatePrompts []views.PromptWithRelations
	for _, p := range promptsWithRelations {
		templatePrompts = append(templatePrompts, views.PromptWithRelations{
			Prompt: p.Prompt,
			Actor:  p.Actor,
		})
	}

	return views.PromptDetailsForm(&prompt).Render(c.Request().Context(), c.Response().Writer)
}

// UpdatePrompt handles POST /prompts/edit/{id} - processes edit form submission
func (h *PromptsHandler) UpdatePrompt(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Prompt ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/prompts?error=Invalid prompt ID")
	}

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
		return c.Redirect(http.StatusSeeOther, "/prompts/edit/"+idStr+"?error=Title and content are required")
	}

	// Convert is_system checkbox to boolean
	isSystem := isSystemStr == "1"

	// Update the prompt (for now without actor_id, we'll add that later with sessions)
	if err := h.promptService.UpdatePrompt(c.Request().Context(), id, nil, title, content, isSystem); err != nil {
		return c.Redirect(http.StatusSeeOther, "/prompts/edit/"+idStr+"?error=Failed to update prompt")
	}

	// Redirect back to prompts page with success message
	return c.Redirect(http.StatusSeeOther, "/prompts?success=Prompt updated successfully")
}
