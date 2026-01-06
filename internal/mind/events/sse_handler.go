package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
)

const (
	// heartbeatInterval is the interval at which the server sends SSE comments
	// to detect dead client connections. If the write fails, the connection is
	// considered dead and will be cleaned up.
	heartbeatInterval = 60 * time.Second
)

// SSEHandler handles Server-Sent Events connections for real-time updates.
type SSEHandler struct {
	hub    Hub
	logger *slog.Logger
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(hub Hub, logger *slog.Logger) *SSEHandler {
	return &SSEHandler{
		hub:    hub,
		logger: logger.With("component", "sse-handler"),
	}
}

// sseEvent is the JSON payload sent in SSE data field.
// We use a simpler struct than the proto to avoid proto JSON quirks.
type sseEvent struct {
	Type            string `json:"type"`
	EntityID        int64  `json:"entity_id,omitempty"`
	Timestamp       int64  `json:"ts"`
	OriginSessionID string `json:"origin_session_id,omitempty"` // Session that triggered this event
	SessionID       string `json:"session_id,omitempty"`        // Assigned session ID (only in connected event)
	// Relocated event payload fields (only set for relocated events)
	OldTitle        *string `json:"old_title,omitempty"`
	NewTitle        *string `json:"new_title,omitempty"`
	OldCollectionID *int64  `json:"old_collection_id,omitempty"`
	NewCollectionID *int64  `json:"new_collection_id,omitempty"`
}

// domainToEventName maps proto domain enum to SSE event name.
func domainToEventName(domain mindv3.EventDomain) string {
	switch domain {
	case mindv3.EventDomain_EVENT_DOMAIN_NOTE:
		return "note"
	case mindv3.EventDomain_EVENT_DOMAIN_COLLECTION:
		return "collection"
	case mindv3.EventDomain_EVENT_DOMAIN_TAG:
		return "tag"
	case mindv3.EventDomain_EVENT_DOMAIN_LINK:
		return "link"
	case mindv3.EventDomain_EVENT_DOMAIN_NOTE_TYPE:
		return "note_type"
	case mindv3.EventDomain_EVENT_DOMAIN_TEMPLATE:
		return "template"
	case mindv3.EventDomain_EVENT_DOMAIN_NOTE_META:
		return "note_meta"
	case mindv3.EventDomain_EVENT_DOMAIN_SYSTEM:
		return "system"
	default:
		return "unknown"
	}
}

// eventTypeToString maps proto event type enum to string.
func eventTypeToString(t mindv3.EventType) string {
	switch t {
	case mindv3.EventType_EVENT_TYPE_CREATED:
		return "created"
	case mindv3.EventType_EVENT_TYPE_UPDATED:
		return "updated"
	case mindv3.EventType_EVENT_TYPE_DELETED:
		return "deleted"
	case mindv3.EventType_EVENT_TYPE_RELOCATED:
		return "relocated"
	case mindv3.EventType_EVENT_TYPE_CONNECTED:
		return "connected"
	case mindv3.EventType_EVENT_TYPE_SHUTDOWN:
		return "shutdown"
	case mindv3.EventType_EVENT_TYPE_HEARTBEAT:
		return "heartbeat"
	default:
		return "unknown"
	}
}

// HandleStream handles GET /events/stream - the SSE endpoint.
// Generates a unique session ID for this connection that clients should
// include in subsequent requests via X-Session-Id header.
func (h *SSEHandler) HandleStream(c echo.Context) error {
	// Generate session ID for this connection
	sessionID := fmt.Sprintf("sess_%s", uuid.NewString())
	h.logger.Info("new SSE connection", "remote_addr", c.RealIP(), "session_id", sessionID)

	// Set SSE headers
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Subscribe to events
	eventCh := h.hub.Subscribe()
	defer h.hub.Unsubscribe(eventCh)

	// Send connected event with assigned session ID
	if err := h.writeEvent(w, &mindv3.Event{
		Id:        0,
		Domain:    mindv3.EventDomain_EVENT_DOMAIN_SYSTEM,
		Type:      mindv3.EventType_EVENT_TYPE_CONNECTED,
		SessionId: sessionID,
	}); err != nil {
		return err
	}
	w.Flush()

	// Get request context for cancellation
	ctx := c.Request().Context()

	// Start heartbeat ticker to detect dead connections
	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("SSE connection closed by client", "remote_addr", c.RealIP(), "session_id", sessionID)
			return nil

		case <-heartbeat.C:
			// Send SSE comment to detect dead connections
			// If write fails, the connection is dead and we should clean up
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				h.logger.Debug("heartbeat failed, client disconnected", "remote_addr", c.RealIP(), "session_id", sessionID)
				return nil // defer Unsubscribe() will clean up
			}
			w.Flush()

		case event, ok := <-eventCh:
			if !ok {
				// Hub closed the channel (shutdown)
				h.logger.Info("SSE connection closed by hub", "remote_addr", c.RealIP(), "session_id", sessionID)
				return nil
			}

			if err := h.writeEvent(w, event); err != nil {
				h.logger.Error("failed to write SSE event", "error", err, "session_id", sessionID)
				return err
			}
			w.Flush()
		}
	}
}

// writeEvent writes a single SSE event to the response.
func (h *SSEHandler) writeEvent(w http.ResponseWriter, event *mindv3.Event) error {
	eventName := domainToEventName(event.Domain)

	data := sseEvent{
		Type:            eventTypeToString(event.Type),
		EntityID:        event.EntityId,
		OriginSessionID: event.OriginSessionId,
		SessionID:       event.SessionId,
	}
	if event.Timestamp != nil {
		data.Timestamp = event.Timestamp.AsTime().UnixMilli()
	}

	// Populate relocated payload if present
	if relocated := event.GetRelocated(); relocated != nil {
		data.OldTitle = relocated.OldTitle
		data.NewTitle = relocated.NewTitle
		data.OldCollectionID = relocated.OldCollectionId
		data.NewCollectionID = relocated.NewCollectionId
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// SSE format:
	// id: <event_id>
	// event: <domain>
	// data: <json>
	// <blank line>
	if event.Id > 0 {
		if _, err := fmt.Fprintf(w, "id: %d\n", event.Id); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", eventName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
		return err
	}

	return nil
}
