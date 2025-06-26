package api

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/adapters"
)

type LLMHandler struct {
	llmService LLMService
}

// LLMService defines the interface for LLM operations
type LLMService interface {
	Generate(ctx echo.Context, prompt string, options adapters.GenerateOptions) (*adapters.GenerateResponse, error)
	Chat(ctx echo.Context, messages []adapters.ChatMessage, options adapters.ChatOptions) (*adapters.ChatResponse, error)
}

// NewLLMHandler creates a new LLMHandler with dependency injection
func NewLLMHandler(llmService LLMService) *LLMHandler {
	return &LLMHandler{
		llmService: llmService,
	}
}

// Generate handles POST /api/generate
func (h *LLMHandler) Generate(c echo.Context) error {
	var request struct {
		Prompt  string                   `json:"prompt"`
		Options adapters.GenerateOptions `json:"options"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if request.Prompt == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Prompt is required"})
	}

	response, err := h.llmService.Generate(c, request.Prompt, request.Options)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}

// Chat handles POST /api/chat
func (h *LLMHandler) Chat(c echo.Context) error {
	var request struct {
		Messages []adapters.ChatMessage `json:"messages"`
		Options  adapters.ChatOptions   `json:"options"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if len(request.Messages) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "At least one message is required"})
	}

	response, err := h.llmService.Chat(c, request.Messages, request.Options)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, response)
}
