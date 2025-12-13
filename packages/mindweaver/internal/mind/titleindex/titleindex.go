// FIXME: this is not the right place for this
// Package titleindex provides a persistent key-value store for note title→uuid mapping
// using BadgerDB. This enables fast WikiLink resolution without database queries.
package titleindex

import (
	"fmt"
	"log/slog"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
)

// TitleIndex provides a persistent key-value store for note title→uuid mapping.
// It uses BadgerDB for fast lookups and persistence across restarts.
type TitleIndex struct {
	db     *badger.DB
	logger *slog.Logger
}

// NewTitleIndex creates a new title index backed by BadgerDB.
// The database will be created at the specified path if it doesn't exist.
func NewTitleIndex(dbPath string, logger *slog.Logger) (*TitleIndex, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable BadgerDB's internal logging

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB at %s: %w", dbPath, err)
	}

	logger.Info("📖 Title index initialized", "path", dbPath)

	return &TitleIndex{
		db:     db,
		logger: logger,
	}, nil
}

// NormalizeTitle converts a note title to a canonical form for consistent lookups.
// This ensures that "My Note", "my note", and "My  Note" all resolve to the same entry.
func NormalizeTitle(title string) string {
	// Convert to lowercase and collapse multiple spaces
	normalized := strings.ToLower(strings.TrimSpace(title))
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

// Set stores a title→uuid mapping in the index.
// The title is automatically normalized for consistent lookups.
func (ti *TitleIndex) Set(title, uuid string) error {
	normalizedTitle := NormalizeTitle(title)

	err := ti.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(normalizedTitle), []byte(uuid))
	})
	if err != nil {
		return fmt.Errorf("failed to set title→uuid mapping: %w", err)
	}

	return nil
}

// Get retrieves the UUID for a given note title.
// The title is automatically normalized before lookup.
// Returns empty string and nil error if the title is not found.
func (ti *TitleIndex) Get(title string) (string, error) {
	normalizedTitle := NormalizeTitle(title)
	var uuid string

	err := ti.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(normalizedTitle))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			uuid = string(val)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		return "", nil // Not found is not an error
	}

	if err != nil {
		return "", fmt.Errorf("failed to get uuid for title: %w", err)
	}

	return uuid, nil
}

// BatchSet stores multiple title→uuid mappings in a single transaction.
// This is more efficient than calling Set() multiple times.
func (ti *TitleIndex) BatchSet(mappings map[string]string) error {
	err := ti.db.Update(func(txn *badger.Txn) error {
		for title, uuid := range mappings {
			normalizedTitle := NormalizeTitle(title)
			if err := txn.Set([]byte(normalizedTitle), []byte(uuid)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to batch set title→uuid mappings: %w", err)
	}

	return nil
}

// Delete removes a title→uuid mapping from the index.
// The title is automatically normalized before deletion.
func (ti *TitleIndex) Delete(title string) error {
	normalizedTitle := NormalizeTitle(title)

	err := ti.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(normalizedTitle))
	})

	if err == badger.ErrKeyNotFound {
		return nil // Already deleted, not an error
	}

	if err != nil {
		return fmt.Errorf("failed to delete title→uuid mapping: %w", err)
	}

	return nil
}

// BatchDelete removes multiple title→uuid mappings in a single transaction.
func (ti *TitleIndex) BatchDelete(titles []string) error {
	err := ti.db.Update(func(txn *badger.Txn) error {
		for _, title := range titles {
			normalizedTitle := NormalizeTitle(title)
			if err := txn.Delete([]byte(normalizedTitle)); err != nil && err != badger.ErrKeyNotFound {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to batch delete title→uuid mappings: %w", err)
	}

	return nil
}

// GetAll retrieves all title→uuid mappings from the index.
// This is useful for verification and synchronization with the database.
func (ti *TitleIndex) GetAll() (map[string]string, error) {
	mappings := make(map[string]string)

	err := ti.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			err := item.Value(func(val []byte) error {
				mappings[key] = string(val)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all mappings: %w", err)
	}

	return mappings, nil
}

// Clear removes all title→uuid mappings from the index.
// Use this carefully - it will require a full rebuild from the database.
func (ti *TitleIndex) Clear() error {
	err := ti.db.DropAll()
	if err != nil {
		return fmt.Errorf("failed to clear title index: %w", err)
	}

	ti.logger.Info("📖 Title index cleared")
	return nil
}

// Close closes the BadgerDB database.
// This should be called when the application shuts down.
func (ti *TitleIndex) Close() error {
	ti.logger.Info("📖 Closing title index")
	return ti.db.Close()
}

// Stats returns statistics about the title index.
type Stats struct {
	KeyCount int64
	DiskSize int64
}

// GetStats returns statistics about the title index.
func (ti *TitleIndex) GetStats() (Stats, error) {
	stats := Stats{}

	// Count keys
	err := ti.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need to count
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			stats.KeyCount++
		}
		return nil
	})
	if err != nil {
		return stats, fmt.Errorf("failed to count keys: %w", err)
	}

	// Get disk size (approximate)
	lsm, vlog := ti.db.Size()
	stats.DiskSize = lsm + vlog

	return stats, nil
}
