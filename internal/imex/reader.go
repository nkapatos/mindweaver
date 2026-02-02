package imex

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"unicode/utf8"

	"github.com/cespare/xxhash/v2"
)

// Reader handles concurrent file reading with a worker pool
type Reader struct {
	opts             ImportOptions
	progress         *ProgressTracker
	errors           *ErrorTracker
	fingerprintCache *FingerprintCache
}

// NewReader creates a new file reader
func NewReader(opts ImportOptions, progress *ProgressTracker, errors *ErrorTracker, cache *FingerprintCache) *Reader {
	return &Reader{
		opts:             opts,
		progress:         progress,
		errors:           errors,
		fingerprintCache: cache,
	}
}

// Read processes files from the paths channel using a worker pool
// Returns a channel of successfully read files
func (r *Reader) Read(ctx context.Context, paths <-chan string) <-chan FileData {
	results := make(chan FileData, 100)

	workers := r.opts.Workers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go r.worker(ctx, paths, results, &wg)
	}

	// Close results channel when all workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// worker processes files from the paths channel
func (r *Reader) worker(ctx context.Context, paths <-chan string, results chan<- FileData, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case path, ok := <-paths:
			if !ok {
				return
			}

			data, err := r.readFile(path)
			if err != nil {
				r.progress.IncrementFailed()
				continue
			}

			// Track whether this was a cache hit
			if data.Cached {
				r.progress.IncrementSkipped()
			} else {
				r.progress.IncrementProcessed()
				r.progress.AddBytesRead(data.Size)
			}

			select {
			case results <- data:
			case <-ctx.Done():
				return
			}
		}
	}
}

// readFile reads a single file and handles all error cases
// Computes xxHash64 of content during read
func (r *Reader) readFile(path string) (FileData, error) {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			r.errors.Add(path, ErrorTypeNotFound, err)
		} else if os.IsPermission(err) {
			r.errors.Add(path, ErrorTypePermission, err)
		} else {
			r.errors.Add(path, ErrorTypeIO, err)
		}
		return FileData{}, err
	}

	// Check fingerprint cache if enabled
	if r.fingerprintCache != nil {
		shouldRead, cachedHash, err := r.fingerprintCache.ShouldRead(path, info)
		if err == nil && !shouldRead {
			// Cache hit - file unchanged
			return FileData{
				Path:    path,
				Content: nil, // No content needed
				Size:    info.Size(),
				Hash:    cachedHash,
				ModTime: info.ModTime().Unix(),
				Cached:  true,
			}, nil
		}
	}

	// Check file size limit
	if r.opts.MaxFileSize > 0 && info.Size() > r.opts.MaxFileSize {
		err := fmt.Errorf("file size %d exceeds limit %d", info.Size(), r.opts.MaxFileSize)
		r.errors.Add(path, ErrorTypeSize, err)
		return FileData{}, err
	}

	// Open file
	file, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) {
			r.errors.Add(path, ErrorTypePermission, err)
		} else {
			r.errors.Add(path, ErrorTypeIO, err)
		}
		return FileData{}, err
	}
	defer file.Close()

	// Stream read content + compute hash simultaneously
	hasher := xxhash.New()
	content, err := io.ReadAll(io.TeeReader(file, hasher))
	if err != nil {
		r.errors.Add(path, ErrorTypeIO, err)
		return FileData{}, err
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(content) {
		err := errors.New("invalid UTF-8 encoding")
		r.errors.Add(path, ErrorTypeEncoding, err)
		return FileData{}, err
	}

	// Compute hash
	hash := fmt.Sprintf("%016x", hasher.Sum64())

	// Update cache if enabled
	if r.fingerprintCache != nil {
		if err := r.fingerprintCache.Update(path, hash, info.Size(), info.ModTime().Unix()); err != nil {
			// Log warning but don't fail import
			fmt.Printf("Warning: Failed to update cache for %s: %v\n", path, err)
		}
	}

	return FileData{
		Path:    path,
		Content: content,
		Size:    info.Size(),
		Hash:    hash,
		ModTime: info.ModTime().Unix(),
		Cached:  false,
	}, nil
}
