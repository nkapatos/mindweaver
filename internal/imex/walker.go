package imex

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

// Walker handles directory traversal and file discovery
type Walker struct {
	opts ImportOptions
}

// NewWalker creates a new file walker
func NewWalker(opts ImportOptions) *Walker {
	return &Walker{opts: opts}
}

// Walk traverses the directory and sends discovered file paths to the channel
// Returns total file count and any fatal errors
func (w *Walker) Walk(ctx context.Context, paths chan<- string) (int64, error) {
	defer close(paths)

	var count int64

	err := filepath.WalkDir(w.opts.Dir, func(path string, d fs.DirEntry, err error) error {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Non-fatal: skip files we can't access
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip symlinks unless explicitly allowed
		if !w.opts.FollowSymlinks {
			if info, err := d.Info(); err == nil {
				if info.Mode()&fs.ModeSymlink != 0 {
					return nil
				}
			}
		}

		// Skip hidden files unless explicitly allowed
		if !w.opts.IncludeHidden {
			if strings.HasPrefix(filepath.Base(path), ".") {
				return nil
			}
		}

		// Filter by extension
		if !w.matchesExtension(path) {
			return nil
		}

		// Check file size if limit is set
		if w.opts.MaxFileSize > 0 {
			if info, err := d.Info(); err == nil {
				if info.Size() > w.opts.MaxFileSize {
					return nil
				}
			}
		}

		count++

		select {
		case paths <- path:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})

	return count, err
}

// matchesExtension checks if file matches any of the configured extensions
func (w *Walker) matchesExtension(path string) bool {
	if len(w.opts.Extensions) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, allowedExt := range w.opts.Extensions {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}

	return false
}

// CountFiles does a dry-run to count matching files without reading them
// Useful for setting progress totals
func CountFiles(dir string, extensions []string, includeHidden bool) (int64, error) {
	walker := NewWalker(ImportOptions{
		Dir:            dir,
		Extensions:     extensions,
		FollowSymlinks: false,
		IncludeHidden:  includeHidden,
	})

	// Create a dummy channel that just counts
	paths := make(chan string, 100)

	ctx := context.Background()
	done := make(chan error, 1)

	go func() {
		for range paths {
		}
		done <- nil
	}()

	total, err := walker.Walk(ctx, paths)
	if err != nil {
		return 0, err
	}

	<-done
	return total, nil
}
