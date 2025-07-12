package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
)

type ConversationHandler struct {
	conversationService *services.ConversationService
	messageService      *services.MessageService
	providerService     *services.ProviderService
	llmService          *services.LLMService
}

// NewConversationHandler creates a new ConversationHandler with dependency injection
func NewConversationHandler(conversationService *services.ConversationService, messageService *services.MessageService, providerService *services.ProviderService, llmService *services.LLMService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		messageService:      messageService,
		providerService:     providerService,
		llmService:          llmService,
	}
}

// CreateConversation handles POST /api/conversations
func (h *ConversationHandler) CreateConversation(c echo.Context) error {
	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	conversation, err := h.conversationService.CreateConversation(c.Request().Context(), actorID, request.Title, request.Description, true, "", actorID, actorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, conversation)
}

// GetConversation handles GET /api/conversations/{id}
func (h *ConversationHandler) GetConversation(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Conversation ID is required"})
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid conversation ID"})
	}

	conversation, err := h.conversationService.GetConversationByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Conversation not found"})
	}

	return c.JSON(http.StatusOK, conversation)
}

// GetConversations handles GET /api/conversations
func (h *ConversationHandler) GetConversations(c echo.Context) error {
	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	conversations, err := h.conversationService.GetConversationsByActorID(c.Request().Context(), actorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, conversations)
}

// CreateMessage handles POST /api/conversations/{id}/messages
func (h *ConversationHandler) CreateMessage(c echo.Context) error {
	conversationIDStr := c.Param("id")
	if conversationIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Conversation ID is required"})
	}

	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid conversation ID"})
	}

	var request struct {
		Content    string `json:"content"`
		ProviderID *int64 `json:"provider_id,omitempty"`
		PromptID   *int64 `json:"prompt_id,omitempty"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if request.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message content is required"})
	}

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)
	systemActorID := int64(1) // System actor ID for AI responses

	// Create the user message
	userMessage, err := h.messageService.CreateMessage(c.Request().Context(), conversationID, actorID, request.Content, "user", "", actorID, actorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// If provider_id is provided, generate AI response
	if request.ProviderID != nil {
		// Generate AI response
		aiResponse, err := h.llmService.GenerateResponse(c.Request().Context(), *request.ProviderID, request.Content, "")
		if err != nil {
			// Log the error but don't fail the user message creation
			c.Logger().Error("Failed to generate AI response", "error", err)
			// Return the user message even if AI response fails
			return c.JSON(http.StatusCreated, map[string]interface{}{
				"user_message":      userMessage,
				"ai_response_error": err.Error(),
			})
		}

		// Create the AI response message
		aiMessage, err := h.messageService.CreateMessage(c.Request().Context(), conversationID, systemActorID, aiResponse, "assistant", "", systemActorID, systemActorID)
		if err != nil {
			// Log the error but don't fail the user message creation
			c.Logger().Error("Failed to create AI response message", "error", err)
			return c.JSON(http.StatusCreated, map[string]interface{}{
				"user_message":      userMessage,
				"ai_response_error": err.Error(),
			})
		}

		// Return both messages
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"user_message": userMessage,
			"ai_message":   aiMessage,
		})
	}

	// Return just the user message if no provider was specified
	return c.JSON(http.StatusCreated, userMessage)
}

// GetMessages handles GET /api/conversations/{id}/messages
func (h *ConversationHandler) GetMessages(c echo.Context) error {
	conversationIDStr := c.Param("id")
	if conversationIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Conversation ID is required"})
	}

	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid conversation ID"})
	}

	messages, err := h.messageService.GetMessagesByConversationID(c.Request().Context(), conversationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, messages)
}

// UpdateMessage handles PUT /api/messages/{id}
func (h *ConversationHandler) UpdateMessage(c echo.Context) error {
	messageIDStr := c.Param("id")
	if messageIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message ID is required"})
	}

	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid message ID"})
	}

	var request struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(c.Request().Body).Decode(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if request.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message content is required"})
	}

	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	err = h.messageService.UpdateMessage(c.Request().Context(), messageID, request.Content, "user", "", actorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Message updated successfully"})
}

// DeleteMessage handles DELETE /api/messages/{id}
func (h *ConversationHandler) DeleteMessage(c echo.Context) error {
	messageIDStr := c.Param("id")
	if messageIDStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Message ID is required"})
	}

	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid message ID"})
	}

	err = h.messageService.DeleteMessage(c.Request().Context(), messageID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Message deleted successfully"})
}
