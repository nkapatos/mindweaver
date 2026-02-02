package imex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type FingerprintEntry struct {
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	Mtime       int64  `json:"mtime"`
	LastChecked int64  `json:"last_checked"`
}

type FingerprintCache struct {
	db      *badger.DB
	baseDir string
}

func NewFingerprintCache(baseDir string) (*FingerprintCache, error) {
	// Get absolute path to the base directory
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of base directory: %v", err)
	}

	// Create the cache directory if it doesn't exist
	cacheDir := filepath.Join(absBaseDir, ".imex")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	// Open BadgerDB in the cache directory
	opts := badger.DefaultOptions(cacheDir).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %v", err)
	}

	return &FingerprintCache{
		db:      db,
		baseDir: absBaseDir,
	}, nil
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

	var entry *FingerprintEntry
	err = fc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(relPath))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			entry = &FingerprintEntry{}
			return json.Unmarshal(val, entry)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			// New file, must read
			return true, "", nil
		}
		return true, "", fmt.Errorf("cache lookup failed: %w", err)
	}

	// Fast path: size and mtime unchanged → file unchanged
	if entry.Size == info.Size() && entry.Mtime == info.ModTime().Unix() {
		return false, entry.Hash, nil
	}

	// Size or mtime changed → must re-read and verify
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

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	return fc.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(relPath), data)
	})
}

// Get retrieves the cached fingerprint for a file (for testing/debugging)
func (fc *FingerprintCache) Get(path string) (*FingerprintEntry, error) {
	relPath, err := fc.relativePath(path)
	if err != nil {
		return nil, err
	}

	var entry FingerprintEntry
	err = fc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(relPath))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &entry)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &entry, nil
}

// Clear removes all entries from the cache
func (fc *FingerprintCache) Clear() error {
	return fc.db.DropAll()
}

// Stats returns cache statistics
func (fc *FingerprintCache) Stats() (int, error) {
	count := 0
	err := fc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})

	return count, err
}

// Close closes the cache database
func (fc *FingerprintCache) Close() error {
	return fc.db.Close()
}

// RunGC triggers garbage collection on the BadgerDB database
// Should be called periodically to reclaim disk space
func (fc *FingerprintCache) RunGC(discardRatio float64) error {
	return fc.db.RunValueLogGC(discardRatio)
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
