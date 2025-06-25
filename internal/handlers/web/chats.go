package web

import (
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

type ChatsHandler struct{}

func NewChatsHandler() *ChatsHandler {
	return &ChatsHandler{}
}

func (h *ChatsHandler) Chats(c echo.Context) error {
	currentPath := c.Path()
	return views.Chat(currentPath).Render(c.Request().Context(), c.Response().Writer)
}
