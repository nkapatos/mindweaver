package transport

import (
	"github.com/google/uuid"
)

// SimpleBatcher implements Batcher interface with basic size/count limits
type SimpleBatcher struct {
	config       BatcherConfig
	currentBatch *BatchRequest
	currentSize  int64
}

// NewSimpleBatcher creates a new simple batcher
func NewSimpleBatcher(config BatcherConfig) *SimpleBatcher {
	// Set defaults
	if config.MaxFiles == 0 {
		config.MaxFiles = 50
	}
	if config.MaxSizeBytes == 0 {
		config.MaxSizeBytes = 10 * 1024 * 1024 // 10MB
	}

	return &SimpleBatcher{
		config: config,
		currentBatch: &BatchRequest{
			BatchID: uuid.New().String(),
			Files:   make([]FileData, 0, config.MaxFiles),
		},
		currentSize: 0,
	}
}

// Add implements Batcher.Add
func (b *SimpleBatcher) Add(file FileData) bool {
	b.currentBatch.Files = append(b.currentBatch.Files, file)
	b.currentSize += file.Size

	return b.IsFull()
}

// Flush implements Batcher.Flush
func (b *SimpleBatcher) Flush() *BatchRequest {
	if len(b.currentBatch.Files) == 0 {
		return nil
	}

	batch := b.currentBatch

	// Reset for next batch
	b.currentBatch = &BatchRequest{
		BatchID: uuid.New().String(),
		Files:   make([]FileData, 0, b.config.MaxFiles),
	}
	b.currentSize = 0

	return batch
}

// IsFull implements Batcher.IsFull
func (b *SimpleBatcher) IsFull() bool {
	return len(b.currentBatch.Files) >= b.config.MaxFiles ||
		b.currentSize >= b.config.MaxSizeBytes
}

// Count implements Batcher.Count
func (b *SimpleBatcher) Count() int {
	return len(b.currentBatch.Files)
}
