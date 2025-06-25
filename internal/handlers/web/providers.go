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
}

func NewProvidersHandler(providerService *services.ProviderService) *ProvidersHandler {
	return &ProvidersHandler{
		providerService: providerService,
	}
}

// Providers handles GET /providers - displays the providers page with form and list
func (h *ProvidersHandler) Providers(c echo.Context) error {
	// Get all providers to display in the list
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		// For now, just log the error and continue with empty list
		// In a real app, you might want to show an error message
		providers = []store.Provider{}
	}

	// For now, we'll pass an empty slice for models
	// This can be enhanced later when we implement model management
	providerModels := []store.Model{}

	return views.ProvidersPage(providers, nil, providerModels).Render(c.Request().Context(), c.Response().Writer)
}

// CreateProvider handles POST /providers - processes form submission
func (h *ProvidersHandler) CreateProvider(c echo.Context) error {
	// Parse form data
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Invalid form data")
	}

	// Extract form values
	name := c.FormValue("name")
	providerType := c.FormValue("type")
	isActiveStr := c.FormValue("is_active")

	// Validate required fields
	if name == "" || providerType == "" {
		// Redirect back to providers page with error
		return c.Redirect(http.StatusSeeOther, "/providers?error=Name and type are required")
	}

	// Convert is_active checkbox to boolean
	isActive := isActiveStr == "1"

	// Create the provider
	if err := h.providerService.CreateProvider(c.Request().Context(), name, providerType, isActive); err != nil {
		// Redirect back with error
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

	// Get all providers to display in the list
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		providers = []store.Provider{}
	}

	// For now, we'll pass an empty slice for models
	providerModels := []store.Model{}

	return views.ProvidersPage(providers, &provider, providerModels).Render(c.Request().Context(), c.Response().Writer)
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
	providerType := c.FormValue("type")
	isActiveStr := c.FormValue("is_active")

	// Validate required fields
	if name == "" || providerType == "" {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Name and type are required")
	}

	// Convert is_active checkbox to boolean
	isActive := isActiveStr == "1"

	// Update the provider
	if err := h.providerService.UpdateProvider(c.Request().Context(), id, name, providerType, isActive); err != nil {
		return c.Redirect(http.StatusSeeOther, "/providers/edit/"+idStr+"?error=Failed to update provider")
	}

	// Redirect back to providers page with success message
	return c.Redirect(http.StatusSeeOther, "/providers?success=Provider updated successfully")
}
