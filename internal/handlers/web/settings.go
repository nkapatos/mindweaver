package web

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

func (h *SettingsHandler) Settings(c echo.Context) error {
	currentPath := c.Path()
	return views.SettingsPage(currentPath).Render(c.Request().Context(), c.Response().Writer)
}
