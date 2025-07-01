package web

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type NotFoundHandler struct{}

func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

func (h *NotFoundHandler) NotFound(c echo.Context) error {
	// fmt.Printf("ctx: %+v\n", c.Request().Context())
	return views.NotFoundPage().Render(c.Request().Context(), c.Response().Writer)
}
