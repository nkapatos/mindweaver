package web

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type ConversationHandler struct{}

func NewConversationHandler() *ConversationHandler {
	return &ConversationHandler{}
}

func (h *ConversationHandler) Conversation(c echo.Context) error {
	currentPath := c.Path()
	return views.Conversation(currentPath).Render(c.Request().Context(), c.Response().Writer)
}
