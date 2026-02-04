package imex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FingerprintEntry struct {
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	Mtime       int64  `json:"mtime"`
	LastChecked int64  `json:"last_checked"`
}

type FingerprintCache struct {
	store         map[string]FingerprintEntry
	cacheFile     string
	baseDir       string
	mu            sync.RWMutex
	dirty         bool
	stopFlush     chan struct{}
	flushDone     chan struct{}
	flushSignal   chan struct{}
	flushInterval time.Duration
	lockFile      string
	lockStale     time.Duration
}

func NewFingerprintCache(baseDir string) (*FingerprintCache, error) {
	// Get absolute path to the base directory
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of base directory: %w", err)
	}

	// Create the cache directory if it doesn't exist
	cacheDir := filepath.Join(absBaseDir, ".imex")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheFile := filepath.Join(cacheDir, "fingerprints.json")
	lockFile := filepath.Join(cacheDir, "fingerprints.lock")

	fc := &FingerprintCache{
		store:         make(map[string]FingerprintEntry),
		cacheFile:     cacheFile,
		baseDir:       absBaseDir,
		stopFlush:     make(chan struct{}),
		flushDone:     make(chan struct{}),
		flushSignal:   make(chan struct{}, 1),
		flushInterval: 2 * time.Second,
		lockFile:      lockFile,
		lockStale:     30 * time.Second,
	}

	// Load existing cache file if present (non-fatal)
	if err := fc.loadFromFile(); err != nil {
		// treat load errors as non-fatal: start with empty cache
		fmt.Printf("Warning: failed to load fingerprint cache: %v\n", err)
	}

	// Start background flusher
	go fc.flushLoop()

	return fc, nil
}

// ShouldRead returns whether the file should be read, the cached hash (if any), and error
// if lookup failed.
// Decision logic:
// - If no cached entry exists → must read (new file)
// - If size or mtime changed → must read (file modified)
// - If size and mtime unchanged → skip read (file unchanged)
func (fc *FingerprintCache) ShouldRead(path string, info os.FileInfo) (bool, string, error) {
	relPath, err := fc.relativePath(path)
	if err != nil {
		return true, "", err
	}
	fc.mu.RLock()
	entry, ok := fc.store[relPath]
	fc.mu.RUnlock()

	if !ok {
		return true, "", nil
	}

	if entry.Size == info.Size() && entry.Mtime == info.ModTime().Unix() {
		return false, entry.Hash, nil
	}

	return true, "", nil
}

// Update stores or updates the fingerprint for a file
func (fc *FingerprintCache) Update(path, hash string, size, mtime int64) error {
	relPath, err := fc.relativePath(path)
	if err != nil {
		return err
	}

	entry := FingerprintEntry{
		Hash:        hash,
		Size:        size,
		Mtime:       mtime,
		LastChecked: time.Now().Unix(),
	}
	fc.mu.Lock()
	fc.store[relPath] = entry
	fc.dirty = true
	fc.mu.Unlock()

	fmt.Printf("Fingerprint Update: %s (size=%d mtime=%d)\n", relPath, size, mtime)

	// signal flusher (non-blocking)
	select {
	case fc.flushSignal <- struct{}{}:
	default:
	}

	return nil
}

// Get retrieves the cached fingerprint for a file (for testing/debugging)
func (fc *FingerprintCache) Get(path string) (*FingerprintEntry, error) {
	relPath, err := fc.relativePath(path)
	if err != nil {
		return nil, err
	}
	fc.mu.RLock()
	e, ok := fc.store[relPath]
	fc.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	// return a copy
	entry := e
	return &entry, nil
}

// Clear removes all entries from the cache
func (fc *FingerprintCache) Clear() error {
	fc.mu.Lock()
	fc.store = make(map[string]FingerprintEntry)
	fc.dirty = true
	fc.mu.Unlock()

	// remove cache file on disk (best-effort)
	if err := os.Remove(fc.cacheFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove cache file %s: %v\n", fc.cacheFile, err)
	}

	// signal flusher to persist cleared state
	select {
	case fc.flushSignal <- struct{}{}:
	default:
	}

	return nil
}

// Stats returns cache statistics
func (fc *FingerprintCache) Stats() (int, error) {
	fc.mu.RLock()
	count := len(fc.store)
	fc.mu.RUnlock()
	return count, nil
}

// Close closes the cache database
func (fc *FingerprintCache) Close() error {
	// stop flusher and wait for final flush
	close(fc.stopFlush)
	<-fc.flushDone
	return nil
}

// RunGC triggers garbage collection on the BadgerDB database
// Should be called periodically to reclaim disk space
func (fc *FingerprintCache) RunGC(discardRatio float64) error {
	// no-op for file-backed JSON cache
	return nil
}

// relativePath converts absolute path to relative path from base directory
func (fc *FingerprintCache) relativePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(fc.baseDir, absPath)
	if err != nil {
		return "", fmt.Errorf("path outside base directory: %w", err)
	}

	return relPath, nil
}

// loadFromFile loads the JSON cache file into memory. If the file does not exist,
// it's a no-op. Returns an error if reading or unmarshaling fails.
func (fc *FingerprintCache) loadFromFile() error {
	data, err := os.ReadFile(fc.cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var m map[string]FingerprintEntry
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	fc.mu.Lock()
	fc.store = m
	fc.mu.Unlock()
	return nil
}

// flushLoop runs in background and persists the in-memory cache to disk.
func (fc *FingerprintCache) flushLoop() {
	defer close(fc.flushDone)

	// simple debounce: when signaled, wait a short period then flush
	debounce := 200 * time.Millisecond

	for {
		select {
		case <-fc.flushSignal:
			// debounce
			timer := time.NewTimer(debounce)
			select {
			case <-timer.C:
			case <-fc.stopFlush:
				timer.Stop()
				fc.flushOnce()
				return
			}
			// after debounce, perform flush
			fc.flushOnce()
			// then wait for either more signals or stop
		case <-fc.stopFlush:
			fc.flushOnce()
			return
		}
	}
}

// flushOnce writes the cache to disk if dirty. Errors are ignored here but could be
// surfaced via logs or returned on Close if desired.
func (fc *FingerprintCache) flushOnce() {
	fc.mu.RLock()
	if !fc.dirty {
		fc.mu.RUnlock()
		return
	}
	// snapshot
	snapshot := make(map[string]FingerprintEntry, len(fc.store))
	for k, v := range fc.store {
		snapshot[k] = v
	}
	fc.mu.RUnlock()

	data, err := json.Marshal(snapshot)
	if err != nil {
		fmt.Printf("Warning: failed to marshal fingerprint snapshot: %v\n", err)
		return
	}

	// attempt to write atomically with a simple lockfile
	if err := fc.atomicWriteWithLock(data); err != nil {
		fmt.Printf("Warning: failed to write fingerprint cache: %v\n", err)
	}

	// clear dirty flag on success
	fc.mu.Lock()
	fc.dirty = false
	fc.mu.Unlock()

	// debug: indicate file written
	fmt.Printf("Fingerprint cache written to %s (entries=%d)\n", fc.cacheFile, len(snapshot))
}

// atomicWriteWithLock writes data to the cache file atomically using a temp file
// and a simple lockfile based on O_EXCL. If a stale lock exists (older than
// lockStale), it may be removed.
func (fc *FingerprintCache) atomicWriteWithLock(data []byte) error {
	// try to acquire lock by creating lock file with O_EXCL
	deadline := time.Now().Add(5 * time.Second)
	for {
		f, err := os.OpenFile(fc.lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			// we own the lock; write metadata (timestamp)
			ts := time.Now().Unix()
			if _, werr := fmt.Fprintf(f, "%d\n", ts); werr != nil {
				// best-effort: log and continue; we still hold the lock file
				fmt.Printf("Warning: failed to write lockfile %s: %v\n", fc.lockFile, werr)
			}
			if serr := f.Sync(); serr != nil {
				fmt.Printf("Warning: failed to sync lockfile %s: %v\n", fc.lockFile, serr)
			}
			if cerr := f.Close(); cerr != nil {
				fmt.Printf("Warning: failed to close lockfile %s: %v\n", fc.lockFile, cerr)
			}
			break
		}
		// couldn't create lock file; check if stale
		st, statErr := os.Stat(fc.lockFile)
		if statErr == nil {
			if time.Since(st.ModTime()) > fc.lockStale {
				// stale: remove it and try again
				if rerr := os.Remove(fc.lockFile); rerr != nil && !os.IsNotExist(rerr) {
					fmt.Printf("Warning: failed to remove stale lockfile %s: %v\n", fc.lockFile, rerr)
				}
				continue
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out acquiring lock for %s", fc.lockFile)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// ensure lockfile removed at the end
	defer func() {
		if rerr := os.Remove(fc.lockFile); rerr != nil && !os.IsNotExist(rerr) {
			fmt.Printf("Warning: failed to remove lockfile %s: %v\n", fc.lockFile, rerr)
		}
	}()

	// write to tmp
	tmp := fc.cacheFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}
	// best effort sync
	if f, err := os.Open(tmp); err == nil {
		if serr := f.Sync(); serr != nil {
			fmt.Printf("Warning: failed to sync tmp file %s: %v\n", tmp, serr)
		}
		if cerr := f.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close tmp file %s: %v\n", tmp, cerr)
		}
	}

	if err := os.Rename(tmp, fc.cacheFile); err != nil {
		return fmt.Errorf("rename tmp: %w", err)
	}

	// fsync containing directory if possible (best-effort)
	dir := filepath.Dir(fc.cacheFile)
	if dfd, err := os.Open(dir); err == nil {
		if serr := dfd.Sync(); serr != nil {
			fmt.Printf("Warning: failed to sync dir %s: %v\n", dir, serr)
		}
		if cerr := dfd.Close(); cerr != nil {
			fmt.Printf("Warning: failed to close dir fd %s: %v\n", dir, cerr)
		}
	}

	return nil
}
