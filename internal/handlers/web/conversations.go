package web

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func (h *ConversationHandler) Conversation(c echo.Context) error {
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load providers")
	}

	// Compose the provider dropdown component with its specific data
	providerDropdownData := components.ProviderDropdownData{
		Providers:        providers,
		SelectedProvider: nil, // No provider selected for main conversation page
		DropdownID:       "conversation-provider-dropdown",
	}

	// Compose the message input component with its specific data
	messageInputData := components.MessageInputData{
		Placeholder: "Type your message here...",
		IsDisabled:  true, // Disabled because no provider is selected
	}

	// For the main conversation page, no conversation is selected yet
	return views.Conversation(providerDropdownData, messageInputData).Render(c.Request().Context(), c.Response().Writer)
}

func (h *ConversationHandler) NewConversation(c echo.Context) error {
	// Get all providers for selection
	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load providers")
	}

	currentPath := c.Path()
	return views.NewConversationPage(providers, currentPath).Render(c.Request().Context(), c.Response().Writer)
}

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

	// For now, use a default actor ID (we'll need to implement user authentication later)
	actorID := int64(1) // TODO: Get from session/authentication

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

	// Create the conversation
	conversation, err := h.conversationService.CreateConversation(
		c.Request().Context(),
		actorID,
		title,
		description,
		true, // isActive
		string(metadataJSON),
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create conversation")
	}

	// Redirect to the conversation
	return c.Redirect(http.StatusSeeOther, "/conversations/"+strconv.FormatInt(conversation.ID, 10))
}

func (h *ConversationHandler) ViewConversation(c echo.Context) error {
	conversationIDStr := c.Param("id")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid conversation ID")
	}

	// Get the conversation with messages
	conversation, _, _, err := h.conversationService.GetConversationWithActorAndMessages(
		c.Request().Context(),
		conversationID,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Conversation not found")
	}

	providers, err := h.providerService.GetAllProviders(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load providers")
	}

	// Extract default provider from conversation metadata
	defaultProvider := getDefaultProviderFromMetadata(conversation, providers)

	// Compose the provider dropdown component with its specific data
	providerDropdownData := components.ProviderDropdownData{
		Providers:        providers,
		SelectedProvider: defaultProvider,
		DropdownID:       "conversation-provider-dropdown",
	}

	// Compose the message input component with its specific data
	messageInputData := components.MessageInputData{
		Placeholder: "Type your message here...",
		IsDisabled:  defaultProvider == nil,
	}

	return views.Conversation(providerDropdownData, messageInputData).Render(c.Request().Context(), c.Response().Writer)
}
