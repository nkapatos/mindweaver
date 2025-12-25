package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nkapatos/mindweaver/packages/mindweaver/shared/types"
)

// HTTPTransportConfig configures the HTTP transport
type HTTPTransportConfig struct {
	BaseURL      string        // Mind service URL (e.g., http://localhost:8081)
	AuthToken    string        // Bearer token for authentication
	Timeout      time.Duration // HTTP request timeout (default: 30s)
	MaxRetries   int           // Maximum number of retries for transient failures (default: 3)
	RetryBackoff time.Duration // Initial backoff duration for retries (default: 1s)
	UserAgent    string        // User-Agent header (default: "imex-cli/1.0")
}

// HTTPTransport implements Transport interface using HTTP
type HTTPTransport struct {
	config HTTPTransportConfig
	client *http.Client
	stats  TransportStats
	mu     sync.Mutex // Protects stats
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(config HTTPTransportConfig) *HTTPTransport {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryBackoff == 0 {
		config.RetryBackoff = 1 * time.Second
	}
	if config.UserAgent == "" {
		config.UserAgent = "imex-cli/1.0"
	}

	return &HTTPTransport{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		stats: TransportStats{},
	}
}

// Send implements Transport.Send
func (t *HTTPTransport) Send(ctx context.Context, batch *BatchRequest) (*Operation, error) {
	start := time.Now()

	// Generate batch ID if not set
	if batch.BatchID == "" {
		batch.BatchID = uuid.New().String()
	}

	// Initialize metadata if nil
	if batch.Metadata == nil {
		batch.Metadata = make(map[string]any)
	}
	batch.Metadata["source"] = "imex-cli"
	batch.Metadata["version"] = "1.0"
	batch.Metadata["sent_at"] = time.Now().UTC().Format(time.RFC3339)

	var lastErr error
	var op *Operation

	// Retry loop
	for attempt := 0; attempt <= t.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := t.config.RetryBackoff * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			t.mu.Lock()
			t.stats.RetryCount++
			t.mu.Unlock()
		}

		op, lastErr = t.sendOnce(ctx, batch)
		if lastErr == nil {
			// Success
			t.updateStats(true, len(batch.Files), batch.Metadata, time.Since(start))
			return op, nil
		}

		// Check if error is retryable
		if !isRetryableError(lastErr) {
			break
		}
	}

	// All retries failed
	t.updateStats(false, len(batch.Files), batch.Metadata, time.Since(start))
	return nil, fmt.Errorf("failed after %d attempts: %w", t.config.MaxRetries+1, lastErr)
}

// sendOnce sends a single HTTP request without retries
// Parses the wrapped types.Response[Operation] from Mind service
func (t *HTTPTransport) sendOnce(ctx context.Context, batch *BatchRequest) (*Operation, error) {
	// Marshal request body
	body, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/mind/import/batch", t.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", t.config.UserAgent)
	if t.config.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.config.AuthToken))
	}

	// Send request
	httpResp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse wrapped response (flattened - data IS the operation)
	var wrappedResp types.Response[Operation]
	if err := json.Unmarshal(respBody, &wrappedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if wrappedResp.Error != nil {
		// Handle 429 Too Many Requests (backpressure)
		if wrappedResp.Error.Code == http.StatusTooManyRequests {
			// Try to extract retry-after from error details
			retryAfter := 5 * time.Second // Default
			queueSize := 0
			for _, detail := range wrappedResp.Error.Details {
				if detail.Domain == "backpressure" {
					// Extract from metadata if available
					if retryStr, ok := detail.Metadata["retry_after"]; ok {
						if duration, err := time.ParseDuration(retryStr); err == nil {
							retryAfter = duration
						}
					}
				}
			}
			return nil, &BackpressureError{
				RetryAfter: retryAfter,
				QueueSize:  queueSize,
				Message:    wrappedResp.Error.Message,
			}
		}

		return nil, fmt.Errorf("HTTP %d: %s", wrappedResp.Error.Code, wrappedResp.Error.Message)
	}

	// Extract operation directly from data (no items[] array)
	if wrappedResp.Data == nil {
		return nil, fmt.Errorf("response missing operation data")
	}

	return wrappedResp.Data, nil
}

// GetOperation implements Transport.GetOperation
// Retrieves the current state of a long-running operation
func (t *HTTPTransport) GetOperation(ctx context.Context, operationName string) (*Operation, error) {
	// Extract operation ID from name (e.g., "operations/uuid" -> "uuid")
	// If operationName doesn't have prefix, use it as-is
	opID := operationName
	if len(operationName) > 11 && operationName[:11] == "operations/" {
		opID = operationName[11:]
	}

	url := fmt.Sprintf("%s/api/mind/import/operations/%s", t.config.BaseURL, opID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", t.config.UserAgent)
	if t.config.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.config.AuthToken))
	}

	// Send request
	httpResp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Handle 404 (operation not found)
	if httpResp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse wrapped response (flattened - data IS the operation)
	var wrappedResp types.Response[Operation]
	if err := json.Unmarshal(respBody, &wrappedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error in response
	if wrappedResp.Error != nil {
		return nil, fmt.Errorf("HTTP %d: %s", wrappedResp.Error.Code, wrappedResp.Error.Message)
	}

	// Extract operation directly from data (no items[] array)
	if wrappedResp.Data == nil {
		return nil, fmt.Errorf("response missing operation data")
	}

	return wrappedResp.Data, nil
}

// StreamOperationProgress streams operation progress updates via Server-Sent Events (SSE)
// Calls the progress callback for each update until the operation completes or context is cancelled
func (t *HTTPTransport) StreamOperationProgress(ctx context.Context, operationName string, progressCallback func(*Operation)) (*Operation, error) {
	// Extract operation ID from name
	opID := operationName
	if len(operationName) > 11 && operationName[:11] == "operations/" {
		opID = operationName[11:]
	}

	// Create HTTP request for SSE stream
	url := fmt.Sprintf("%s/api/mind/import/operations/%s/stream", t.config.BaseURL, opID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", t.config.UserAgent)
	if t.config.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.config.AuthToken))
	}

	// Send request
	httpResp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Handle non-200 status
	if httpResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, string(body))
	}

	// Read SSE events
	scanner := bufio.NewScanner(httpResp.Body)
	var lastOp *Operation

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return lastOp, ctx.Err()
		default:
		}

		line := scanner.Text()

		// SSE events start with "data: "
		if strings.HasPrefix(line, "data: ") {
			jsonData := strings.TrimPrefix(line, "data: ")

			// Parse SSE data (simplified format - not full Operation)
			var eventData struct {
				Status         string       `json:"status"`
				ProgressPct    int          `json:"progress_pct"`
				ProcessedFiles int          `json:"processed_files"`
				TotalFiles     int          `json:"total_files"`
				Done           bool         `json:"done"`
				Result         *BatchResult `json:"result,omitempty"`
			}

			if err := json.Unmarshal([]byte(jsonData), &eventData); err != nil {
				// Skip malformed events
				continue
			}

			// Reconstruct operation from event data
			op := &Operation{
				Name: operationName,
				Done: eventData.Done,
				Metadata: &OperationMeta{
					Status:         eventData.Status,
					ProgressPct:    eventData.ProgressPct,
					ProcessedFiles: eventData.ProcessedFiles,
					TotalFiles:     eventData.TotalFiles,
					UpdateTime:     time.Now().UTC(),
				},
				Result: eventData.Result,
			}

			lastOp = op

			// Call progress callback
			if progressCallback != nil {
				progressCallback(op)
			}

			// If operation is done, return
			if eventData.Done {
				return op, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return lastOp, fmt.Errorf("SSE stream error: %w", err)
	}

	// Stream ended without completion
	if lastOp == nil {
		return nil, fmt.Errorf("SSE stream ended without any data")
	}

	return lastOp, nil
}

// WaitForCompletion implements Transport.WaitForCompletion
// Uses SSE streaming if available, falls back to polling if SSE fails
func (t *HTTPTransport) WaitForCompletion(ctx context.Context, operationName string, pollInterval time.Duration) (*Operation, error) {
	// Try SSE first
	op, err := t.StreamOperationProgress(ctx, operationName, nil)
	if err == nil {
		return op, nil
	}

	// SSE failed, fall back to polling
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			op, err := t.GetOperation(ctx, operationName)
			if err != nil {
				return nil, err
			}
			if op == nil {
				return nil, fmt.Errorf("operation not found: %s", operationName)
			}

			// Check if done
			if op.Done {
				return op, nil
			}

			// Continue polling
		}
	}
}

// Close implements Transport.Close
func (t *HTTPTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

// Stats implements Transport.Stats
func (t *HTTPTransport) Stats() TransportStats {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.stats
}

// updateStats updates transport statistics
func (t *HTTPTransport) updateStats(success bool, fileCount int, metadata map[string]any, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stats.TotalBatches++
	t.stats.TotalFiles += fileCount
	t.stats.Duration += duration

	if success {
		t.stats.SuccessCount++
	} else {
		t.stats.FailureCount++
	}

	// Update average latency
	if t.stats.TotalBatches > 0 {
		t.stats.AverageLatency = t.stats.Duration / time.Duration(t.stats.TotalBatches)
	}

	// Calculate total bytes (if metadata contains size info)
	if size, ok := metadata["total_size_bytes"].(int64); ok {
		t.stats.TotalBytes += size
	}
}

// BackpressureError represents a 429 Too Many Requests error
type BackpressureError struct {
	RetryAfter time.Duration
	QueueSize  int
	Message    string
}

func (e *BackpressureError) Error() string {
	return fmt.Sprintf("backpressure: %s (retry after %s, queue size: %d)",
		e.Message, e.RetryAfter, e.QueueSize)
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Backpressure errors are retryable
	if _, ok := err.(*BackpressureError); ok {
		return true
	}

	// Context errors are not retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Network errors are retryable
	// (This is a simplified check; production code should be more thorough)
	return true
}
