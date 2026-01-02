// Package events provides a publish-subscribe hub for domain events.
// The hub fans out events to all subscribers (SSE connections, future gRPC streams, etc.)
// using a fire-and-forget model - slow subscribers are dropped, not buffered.
package events

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	mindv3 "github.com/nkapatos/mindweaver/gen/proto/mind/v3"
	"github.com/nkapatos/mindweaver/shared/middleware"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Hub is the interface for publishing and subscribing to domain events.
type Hub interface {
	// Publish sends an event to all subscribers. Non-blocking, fire-and-forget.
	// The context is used to extract the origin session ID from the request.
	Publish(ctx context.Context, domain mindv3.EventDomain, eventType mindv3.EventType, entityID int64)

	// Subscribe returns a channel that receives events. The channel is closed
	// when Unsubscribe is called or the hub is closed.
	Subscribe() <-chan *mindv3.Event

	// Unsubscribe removes a subscription. The channel will be closed.
	Unsubscribe(ch <-chan *mindv3.Event)

	// Close shuts down the hub and all subscriptions.
	Close()
}

// hub is the concrete implementation of Hub.
type hub struct {
	mu          sync.RWMutex
	subscribers map[chan *mindv3.Event]struct{}
	eventID     atomic.Int64
	logger      *slog.Logger
	closed      bool
}

// NewHub creates a new event hub.
func NewHub(logger *slog.Logger) Hub {
	return &hub{
		subscribers: make(map[chan *mindv3.Event]struct{}),
		logger:      logger.With("component", "event-hub"),
	}
}

// Publish sends an event to all subscribers.
// Non-blocking: if a subscriber's channel is full, the event is dropped for that subscriber.
// Extracts session ID from context to include as origin_session_id in the event.
func (h *hub) Publish(ctx context.Context, domain mindv3.EventDomain, eventType mindv3.EventType, entityID int64) {
	event := &mindv3.Event{
		Id:              h.eventID.Add(1),
		Domain:          domain,
		Type:            eventType,
		EntityId:        entityID,
		Timestamp:       timestamppb.New(time.Now()),
		OriginSessionId: middleware.GetSessionID(ctx),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.closed {
		return
	}

	for ch := range h.subscribers {
		select {
		case ch <- event:
			// sent
		default:
			// subscriber too slow, drop event
			h.logger.Warn("dropped event for slow subscriber",
				"event_id", event.Id,
				"domain", domain.String(),
				"type", eventType.String(),
			)
		}
	}
}

// Subscribe creates a new subscription and returns the event channel.
// The channel has a buffer to absorb small bursts.
func (h *hub) Subscribe() <-chan *mindv3.Event {
	ch := make(chan *mindv3.Event, 64) // buffer for burst tolerance

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		close(ch)
		return ch
	}

	h.subscribers[ch] = struct{}{}
	h.logger.Info("new subscriber", "total_subscribers", len(h.subscribers))

	return ch
}

// Unsubscribe removes a subscription and closes its channel.
func (h *hub) Unsubscribe(ch <-chan *mindv3.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Type assert to get the writable channel for map lookup
	if writeCh, ok := h.findWriteChannel(ch); ok {
		delete(h.subscribers, writeCh)
		close(writeCh)
		h.logger.Info("subscriber removed", "total_subscribers", len(h.subscribers))
	}
}

// findWriteChannel finds the writable channel that matches the read-only channel.
// Must be called with lock held.
func (h *hub) findWriteChannel(readCh <-chan *mindv3.Event) (chan *mindv3.Event, bool) {
	for ch := range h.subscribers {
		if ch == readCh {
			return ch, true
		}
	}
	return nil, false
}

// Close shuts down the hub, broadcasting a shutdown event and closing all subscriptions.
func (h *hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}
	h.closed = true

	// Send shutdown event to all subscribers
	shutdownEvent := &mindv3.Event{
		Id:        h.eventID.Add(1),
		Domain:    mindv3.EventDomain_EVENT_DOMAIN_SYSTEM,
		Type:      mindv3.EventType_EVENT_TYPE_SHUTDOWN,
		EntityId:  0,
		Timestamp: timestamppb.New(time.Now()),
	}

	for ch := range h.subscribers {
		select {
		case ch <- shutdownEvent:
		default:
		}
		close(ch)
	}

	h.subscribers = nil
	h.logger.Info("event hub closed")
}
