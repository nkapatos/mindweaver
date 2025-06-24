package web

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type PromptsHandler struct{}

func NewPromptsHandler() *PromptsHandler {
	return &PromptsHandler{}
}

func (h *PromptsHandler) Prompts(c echo.Context) error {
	return views.PromptsPage().Render(c.Request().Context(), c.Response().Writer)
}
