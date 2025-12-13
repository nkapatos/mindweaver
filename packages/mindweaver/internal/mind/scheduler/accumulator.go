// Package scheduler tracks note changes and periodically syncs them to Brain.
// This implements the "push" pattern where Mind notifies Brain of changes,
// rather than Brain polling Mind for updates.
package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// ChangeEvent represents a single note modification that Brain should process.
type ChangeEvent struct {
	EventType  string    `json:"event_type"`  // "note_created", "note_updated", "note_deleted"
	NoteID     int64     `json:"note_id"`     // ID of the affected note
	Timestamp  time.Time `json:"timestamp"`   // When the change occurred
	UserAction bool      `json:"user_action"` // true if user-initiated (vs. system)
}

// ChangeAccumulator collects note changes and periodically flushes them to Brain.
// This batching reduces the number of HTTP requests and allows Brain to process
// changes efficiently.
type ChangeAccumulator struct {
	mu       sync.Mutex
	changes  []ChangeEvent
	ticker   *time.Ticker
	stopChan chan struct{}

	brainURL string // Brain ingestion API endpoint
	logger   *slog.Logger

	// Config
	flushInterval time.Duration
	batchSize     int // Max changes per batch
}

// Config holds scheduler configuration.
type Config struct {
	BrainURL      string        // e.g., "http://localhost:8080"
	FlushInterval time.Duration // e.g., 5 * time.Minute
	BatchSize     int           // e.g., 100
}

// NewChangeAccumulator creates a new change accumulator.
func NewChangeAccumulator(cfg Config, logger *slog.Logger) *ChangeAccumulator {
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 5 * time.Minute // Default: 5 minutes
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 100 // Default: 100 changes per batch
	}

	return &ChangeAccumulator{
		changes:       make([]ChangeEvent, 0),
		stopChan:      make(chan struct{}),
		brainURL:      cfg.BrainURL,
		logger:        logger.With("component", "scheduler"),
		flushInterval: cfg.FlushInterval,
		batchSize:     cfg.BatchSize,
	}
}

// Start begins accumulating changes and flushing them periodically.
func (c *ChangeAccumulator) Start() {
	c.logger.Info("starting change accumulator",
		"flush_interval", c.flushInterval,
		"batch_size", c.batchSize,
		"brain_url", c.brainURL)

	c.ticker = time.NewTicker(c.flushInterval)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				if err := c.flush(context.Background()); err != nil {
					c.logger.Error("failed to flush changes", "error", err)
				}
			case <-c.stopChan:
				c.logger.Info("stopping change accumulator")
				return
			}
		}
	}()
}

// Stop stops the accumulator and flushes any pending changes.
func (c *ChangeAccumulator) Stop() error {
	c.logger.Info("stopping change accumulator")

	if c.ticker != nil {
		c.ticker.Stop()
	}

	close(c.stopChan)

	// Final flush
	return c.flush(context.Background())
}

// TrackChange records a note modification event.
// This is called by Mind's note services after create/update/delete operations.
func (c *ChangeAccumulator) TrackChange(eventType string, noteID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.changes = append(c.changes, ChangeEvent{
		EventType:  eventType,
		NoteID:     noteID,
		Timestamp:  time.Now(),
		UserAction: true, // All tracked changes are user-initiated
	})

	c.logger.Debug("tracked change",
		"event_type", eventType,
		"note_id", noteID,
		"pending_changes", len(c.changes))

	// If we hit the batch size limit, flush immediately
	if len(c.changes) >= c.batchSize {
		c.logger.Info("batch size limit reached, flushing immediately",
			"pending_changes", len(c.changes))
		go func() {
			if err := c.flush(context.Background()); err != nil {
				c.logger.Error("failed to flush changes", "error", err)
			}
		}()
	}
}

// flush sends accumulated changes to Brain's ingestion API.
func (c *ChangeAccumulator) flush(ctx context.Context) error {
	c.mu.Lock()

	if len(c.changes) == 0 {
		c.mu.Unlock()
		c.logger.Debug("no changes to flush")
		return nil
	}

	// Take a snapshot and clear the accumulator
	changesToFlush := make([]ChangeEvent, len(c.changes))
	copy(changesToFlush, c.changes)
	c.changes = c.changes[:0] // Clear

	c.mu.Unlock()

	c.logger.Info("flushing changes to Brain",
		"count", len(changesToFlush),
		"brain_url", c.brainURL)

	// Send to Brain
	if err := c.sendToBrain(ctx, changesToFlush); err != nil {
		c.logger.Error("failed to send changes to Brain",
			"error", err,
			"count", len(changesToFlush))

		// TODO: Implement retry queue for failed batches
		// For now, we log and drop (Brain can re-ingest via manual API if needed)
		return err
	}

	c.logger.Info("successfully flushed changes to Brain", "count", len(changesToFlush))
	return nil
}

// sendToBrain sends a batch of changes to Brain's ingestion API.
func (c *ChangeAccumulator) sendToBrain(ctx context.Context, changes []ChangeEvent) error {
	endpoint := fmt.Sprintf("%s/api/brain/ingest/batch", c.brainURL)

	payload := map[string]interface{}{
		"changes": changes,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal changes: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Brain returned non-OK status: %d", resp.StatusCode)
	}

	c.logger.Debug("Brain accepted batch", "status", resp.StatusCode)
	return nil
}

// GetPendingCount returns the number of changes waiting to be flushed.
// Useful for monitoring/debugging.
func (c *ChangeAccumulator) GetPendingCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.changes)
}
