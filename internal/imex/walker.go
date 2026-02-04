package imex

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

	// visited tracks inode keys we've recursed into to avoid cycles
	visited := make(map[uint64]struct{})

	inodeKey := func(p string) (uint64, bool) {
		st, err := os.Stat(p)
		if err != nil {
			return 0, false
		}
		sys := st.Sys()
		if sys == nil {
			return 0, false
		}
		s, ok := sys.(*syscall.Stat_t)
		if !ok {
			return 0, false
		}
		// combine device and inode into single 64-bit key
		// combine device and inode into a stable 64-bit key using FNV-1a
		h := fnv.New64a()
		if _, err := fmt.Fprintf(h, "%v-%v", s.Dev, s.Ino); err != nil {
			// fmt.Fprintf on a hash writer should not fail; log just in case
			fmt.Printf("Warning: failed to write inode key to hasher: %v\n", err)
		}
		return h.Sum64(), true
	}

	var walkDir func(dir string) error
	walkDir = func(dir string) error {
		// Respect cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			// Non-fatal: skip directories we can't read
			return nil
		}

		for _, e := range entries {
			entryPath := filepath.Join(dir, e.Name())

			// Respect cancellation between entries
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Use Lstat to detect symlinks without following them
			info, lerr := os.Lstat(entryPath)
			if lerr != nil {
				// skip entries we can't stat
				continue
			}

			// Handle symlinks
			if info.Mode()&os.ModeSymlink != 0 {
				if !w.opts.FollowSymlinks {
					continue
				}

				// Resolve the symlink target
				resolved, err := filepath.EvalSymlinks(entryPath)
				if err != nil {
					continue
				}

				tinfo, err := os.Stat(resolved)
				if err != nil {
					continue
				}

				if tinfo.IsDir() {
					// Recurse into resolved directory if not visited
					if k, ok := inodeKey(resolved); ok {
						if _, seen := visited[k]; seen {
							continue
						}
						visited[k] = struct{}{}
					}
					if err := walkDir(resolved); err != nil {
						return err
					}
					continue
				}

				// Treat resolved target as a regular file for filtering below
				// Use entryPath as the reported path but tinfo for size checks
				if !w.opts.IncludeHidden {
					if strings.HasPrefix(filepath.Base(entryPath), ".") {
						continue
					}
				}
				if !w.matchesExtension(entryPath) {
					continue
				}
				if w.opts.MaxFileSize > 0 && tinfo.Size() > w.opts.MaxFileSize {
					continue
				}

				count++
				select {
				case paths <- entryPath:
				case <-ctx.Done():
					return ctx.Err()
				}
				continue
			}

			// Non-symlink directory: recurse
			if info.IsDir() {
				if k, ok := inodeKey(entryPath); ok {
					if _, seen := visited[k]; seen {
						continue
					}
					visited[k] = struct{}{}
				}
				if err := walkDir(entryPath); err != nil {
					return err
				}
				continue
			}

			// Regular file: apply filters
			if !w.opts.IncludeHidden {
				if strings.HasPrefix(filepath.Base(entryPath), ".") {
					continue
				}
			}
			if !w.matchesExtension(entryPath) {
				continue
			}
			if w.opts.MaxFileSize > 0 && info.Size() > w.opts.MaxFileSize {
				continue
			}

			count++
			select {
			case paths <- entryPath:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}

	// Seed visited with the starting directory inode key
	if k, ok := inodeKey(w.opts.Dir); ok {
		visited[k] = struct{}{}
	}

	if err := walkDir(w.opts.Dir); err != nil {
		return count, err
	}
	return count, nil
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
