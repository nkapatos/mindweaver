package web

import (
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
