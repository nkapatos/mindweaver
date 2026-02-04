package imex

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CollectionSummary aggregates stats for a single collection prefix
type CollectionSummary struct {
	Files int
	Bytes int64
}

// ImportSummary aggregates global import statistics including per-collection breakdown
type ImportSummary struct {
	TotalFiles  int64
	TotalBytes  int64
	Processed   int64
	Skipped     int64
	Failed      int64
	BytesRead   int64
	Collections map[string]CollectionSummary
}

// deriveCollection derives a collection name for a given absolute path.
// If opts.Collection is set, use that as a prefix and derive sub-collection based on
// the first path segment under opts.Dir. If the file is at the root, use opts.Collection
// directly.
func deriveCollection(root, absPath, prefix string) string {
	rel, err := filepath.Rel(root, absPath)
	if err != nil || rel == "." || rel == "" {
		return prefix
	}
	// If the file is directly under the root, return prefix
	dir := filepath.Dir(rel)
	if dir == "." || dir == "" {
		return prefix
	}
	// Split on OS path separator and use the first non-empty segment
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) > 0 && parts[0] != "" {
		return filepath.Join(prefix, parts[0])
	}
	return prefix
}

// ComputeImportSummaryFromOptions walks the tree according to opts and computes
// an ImportSummary without performing reads or sends. It uses the walker to
// discover matching files and stats via os.Stat to gather sizes.
func ComputeImportSummaryFromOptions(opts ImportOptions) (ImportSummary, error) {
	// Expect caller to provide initialized options (defaults applied). Use
	// the provided PathChanBuffer when creating the channel.
	w := NewWalker(opts)
	paths := make(chan string, opts.PathChanBuffer)
	go func() {
		// ignore walker count and errors here; Walk sends paths and returns when done
		if _, err := w.Walk(context.Background(), paths); err != nil {
			// log and let the consumer continue; Walk itself closes paths on return
			fmt.Printf("Warning: walker error during summary: %v\n", err)
		}
	}()

	cols := make(map[string]CollectionSummary)
	var totalFiles int64
	var totalBytes int64

	for p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			// skip files we can't stat
			continue
		}
		col := deriveCollection(opts.Dir, p, opts.Collection)
		s := cols[col]
		s.Files++
		s.Bytes += fi.Size()
		cols[col] = s
		totalFiles++
		totalBytes += fi.Size()
	}

	return ImportSummary{
		TotalFiles:  totalFiles,
		TotalBytes:  totalBytes,
		Collections: cols,
	}, nil
}
