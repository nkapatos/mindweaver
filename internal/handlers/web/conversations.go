package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nkapatos/mindweaver/internal/services"
	"github.com/nkapatos/mindweaver/internal/store"
	"github.com/nkapatos/mindweaver/internal/templates/components"
	"github.com/nkapatos/mindweaver/internal/templates/views"
)

// ConversationMetadata represents the metadata structure for conversations
type ConversationMetadata struct {
	DefaultProviderID   int64  `json:"default_provider_id"`
	DefaultProviderName string `json:"default_provider_name"`
	ConversationType    string `json:"conversation_type,omitempty"`
}

// getDefaultProviderFromMetadata extracts provider info from conversation metadata
func getDefaultProviderFromMetadata(conversation *store.Conversation, providers []store.Provider) *store.Provider {
	if conversation == nil || !conversation.Metadata.Valid {
		return nil
	}

	var metadata ConversationMetadata
	if err := json.Unmarshal([]byte(conversation.Metadata.String), &metadata); err != nil {
		return nil
	}

	// Find the provider in the providers list
	for _, provider := range providers {
		if provider.ID == metadata.DefaultProviderID {
			return &provider
		}
	}

	return nil
}

type ConversationHandler struct {
	conversationService *services.ConversationService
	providerService     *services.ProviderService
}

func NewConversationHandler(conversationService *services.ConversationService, providerService *services.ProviderService) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		providerService:     providerService,
	}
}

// Conversation handles GET /conversations - displays the conversations page
func (h *ConversationHandler) Conversation(c echo.Context) error {
	// Get actor ID from authentication context
	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	// Get all conversations for the actor
	_, err := h.conversationService.GetConversationsByActorID(c.Request().Context(), actorID)
	if err != nil {
		// For now, just log the error and continue with empty list
	}

	// Get all providers for the dropdown
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		providers = []store.Provider{}
	}

	// Use the NewConversationPage template
	return views.NewConversationPage(providers, "/conversations").Render(c.Request().Context(), c.Response().Writer)
}

// NewConversation handles GET /conversations/new - shows create form
func (h *ConversationHandler) NewConversation(c echo.Context) error {
	// Get all providers for the dropdown
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		providers = []store.Provider{}
	}

	return views.NewConversationPage(providers, "/conversations/new").Render(c.Request().Context(), c.Response().Writer)
}

// CreateConversation handles POST /conversations/create - processes form submission
func (h *ConversationHandler) CreateConversation(c echo.Context) error {
	// Parse form data
	providerIDStr := c.FormValue("provider_id")
	title := c.FormValue("title")
	description := c.FormValue("description")

	if providerIDStr == "" || title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Provider and title are required")
	}

	providerID, err := strconv.ParseInt(providerIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid provider ID")
	}

	// Get the provider to include its name in metadata
	provider, err := h.providerService.GetProviderByID(c.Request().Context(), providerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Provider not found")
	}

	// Get actor ID from authentication context
	sess, _ := session.Get("session", c)
	actorID, _ := sess.Values["actor_id"].(int64)

	// Create metadata with the default provider
	metadata := ConversationMetadata{
		DefaultProviderID:   providerID,
		DefaultProviderName: provider.Name,
		ConversationType:    "chat",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create metadata")
	}

	// Get actor ID from authentication context for audit trail
	sess, _ = session.Get("session", c)
	systemActorID, _ := sess.Values["actor_id"].(int64)

	// Create the conversation
	conversation, err := h.conversationService.CreateConversation(
		c.Request().Context(),
		actorID,
		title,
		description,
		true, // isActive
		string(metadataJSON),
		systemActorID, // createdBy
		systemActorID, // updatedBy
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create conversation")
	}

	// Redirect to the conversation
	return c.Redirect(http.StatusSeeOther, "/conversations/"+strconv.FormatInt(conversation.ID, 10))
}

// ViewConversation handles GET /conversations/{id} - shows conversation view
func (h *ConversationHandler) ViewConversation(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Conversation ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid conversation ID")
	}

	// Get the conversation with messages
	_, _, err = h.conversationService.GetConversationWithMessages(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Conversation not found")
	}

	// Get all providers for the dropdown
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		providers = []store.Provider{}
	}

	// Use the Conversation template with provider dropdown data
	providerDropdownData := components.ProviderDropdownData{
		Providers:        providers,
		SelectedProvider: nil, // Will be set based on conversation metadata
	}

	messageInputData := components.MessageInputData{
		Placeholder: "Type your message...",
		IsDisabled:  false,
	}

	return views.Conversation(providerDropdownData, messageInputData).Render(c.Request().Context(), c.Response().Writer)
}
