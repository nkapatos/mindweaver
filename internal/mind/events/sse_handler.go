package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
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
	Type     string `json:"type"`
	EntityID int64  `json:"entity_id,omitempty"`
	// Timestamp as Unix milliseconds for easy client parsing
	Timestamp int64 `json:"ts"`
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
func (h *SSEHandler) HandleStream(c echo.Context) error {
	h.logger.Info("new SSE connection", "remote_addr", c.RealIP())

	// Set SSE headers
	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Subscribe to events
	eventCh := h.hub.Subscribe()
	defer h.hub.Unsubscribe(eventCh)

	// Send connected event
	if err := h.writeEvent(w, &mindv3.Event{
		Id:     0,
		Domain: mindv3.EventDomain_EVENT_DOMAIN_SYSTEM,
		Type:   mindv3.EventType_EVENT_TYPE_CONNECTED,
	}); err != nil {
		return err
	}
	w.Flush()

	// Get request context for cancellation
	ctx := c.Request().Context()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("SSE connection closed by client", "remote_addr", c.RealIP())
			return nil

		case event, ok := <-eventCh:
			if !ok {
				// Hub closed the channel (shutdown)
				h.logger.Info("SSE connection closed by hub", "remote_addr", c.RealIP())
				return nil
			}

			if err := h.writeEvent(w, event); err != nil {
				h.logger.Error("failed to write SSE event", "error", err)
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
		Type:     eventTypeToString(event.Type),
		EntityID: event.EntityId,
	}
	if event.Timestamp != nil {
		data.Timestamp = event.Timestamp.AsTime().UnixMilli()
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
