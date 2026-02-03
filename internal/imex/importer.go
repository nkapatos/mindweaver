package imex

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"
)

// FileData represents a successfully read file (internal to imex)
// This is converted to transport.FileData when sending to Mind
type FileData struct {
	Path    string // Absolute path to the file
	Content []byte // Raw file content
	Size    int64  // File size in bytes
	Hash    string // xxHash64 hex string of content
	ModTime int64  // File modification time (unix timestamp)
	Cached  bool   // Whether this was a cache hit (skipped read)
}

// ImportResult contains the outcome of an import operation
type ImportResult struct {
	Processed int              // Number of successfully processed files
	Failed    int              // Number of failed files
	Errors    map[string]error // Map of file path -> error for failed files
	Duration  float64          // Total duration in seconds
}

// ImportOptions configures the import behavior
type ImportOptions struct {
	Dir            string   // Root directory to scan
	Extensions     []string // File extensions to include (e.g., [".md", ".txt"])
	Workers        int      // Number of concurrent workers (0 = NumCPU)
	MaxFileSize    int64    // Maximum file size in bytes (0 = no limit)
	FollowSymlinks bool     // Whether to follow symbolic links
	IncludeHidden  bool     // Whether to include hidden files (starting with .)
	Incremental    bool     // Enable fingerprint caching for incremental imports
	ClearCache     bool     // Clear fingerprint cache before import

	// Optional: If set, imported files will be sent to the transport
	Transport Transport // Transport to send files to (e.g., HTTP to Mind)
	Batcher   Batcher   // Batcher for grouping files (if nil, creates default)
}

// Importer handles concurrent file import operations
type Importer struct {
	opts             ImportOptions
	ctx              context.Context
	cancel           context.CancelFunc
	progress         *ProgressTracker
	errors           *ErrorTracker
	fingerprintCache *FingerprintCache
	transport        Transport
	batcher          Batcher
	rootDir          string // Root directory for relative path calculation
}

func NewImporter(opts ImportOptions) *Importer {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize fingerprint cache if incremental import is enabled
	var cache *FingerprintCache
	if opts.Incremental {
		var err error
		cache, err = NewFingerprintCache(opts.Dir)
		if err != nil {
			// Handle cache initialization error (log and continue without cache)
			fmt.Println("Warning: failed to initialize fingerprint cache:", err)
			cache = nil
		}
		// Clear cache if requested
		if opts.ClearCache && cache != nil {
			if err := cache.Clear(); err != nil {
				fmt.Println("Warning: failed to clear fingerprint cache:", err)
			}
		}
	}

	// Initialize transport
	transport := opts.Transport
	if transport == nil {
		transport = NewDirectTransport(TransportConfig{})
	}

	// Initialize batcher
	batcher := opts.Batcher
	if batcher == nil {
		batcher = NewSimpleBatcher(BatcherConfig{})
	}

	return &Importer{
		opts:             opts,
		ctx:              ctx,
		cancel:           cancel,
		progress:         NewProgressTracker(),
		errors:           NewErrorTracker(),
		fingerprintCache: cache,
		transport:        transport,
		batcher:          batcher,
		rootDir:          opts.Dir,
	}
}

// RunImport handles import subcommand logic and flag parsing
func RunImport(args []string, dryRun bool, config string) {
	importFlags := flag.NewFlagSet("import", flag.ExitOnError)
	var (
		src              = importFlags.String("src", "", "Source path or URI for import (required)")
		collection       = importFlags.String("collection", "", "Collection name to import into")
		cacheFingerprint = importFlags.Bool("cache-fingerprint", false, "Enable fingerprinting for caching")
		followSymlinks   = importFlags.Bool("follow-symlinks", false, "Follow symlinks during import")
	)
	if err := importFlags.Parse(args); err != nil {
		fmt.Println("Failed to parse import flags:", err)
		return
	}
	// Validate required flag
	if *src == "" {
		fmt.Println("--src is required for import")
		importFlags.Usage()
		return
	}
	fmt.Printf("[import] src=%s, collection=%s, cacheFingerprint=%v, followSymlinks=%v, dryRun=%v, config=%s\n",
		*src, *collection, *cacheFingerprint, *followSymlinks, dryRun, config,
	)

	// Pass the flags to the Importer and execute the import process
	importer := NewImporter(ImportOptions{
		Dir:            *src,
		FollowSymlinks: *followSymlinks,
		Incremental:    *cacheFingerprint,
		Transport:      nil, // Configure transport as needed
	})
	defer importer.Stop()

	fileChan := importer.Import()

	// Consume fileChan to ensure import completes
	for range fileChan {
		// No-op; just consuming the channel
	}

	// Print final summary
	importer.PrintSummary()
}

// Import executes the import operation
// Returns a channel of successfully read files and waits for completion
// If transport is configured, files are batched and sent automatically
func (imp *Importer) Import() <-chan FileData {
	// Start progress reporting (update every 200ms)
	imp.progress.StartReporting(200000000) // 200ms

	// Phase 1: Walk directory and discover files
	paths := make(chan string, 100)
	walker := NewWalker(imp.opts)

	go func() {
		totalFiles, err := walker.Walk(imp.ctx, paths)
		if err != nil {
			// Fatal error during walk - cancel everything
			imp.cancel()
		}
		imp.progress.SetTotal(totalFiles)
	}()

	// Phase 2: Read files concurrently
	reader := NewReader(imp.opts, imp.progress, imp.errors, imp.fingerprintCache)
	results := reader.Read(imp.ctx, paths)

	// Phase 3: If transport configured, batch and send files
	if imp.transport != nil {
		output := make(chan FileData, 100)
		go imp.sendToTransport(results, output)
		return output
	}

	return results
}

// sendToTransport batches and sends files to the configured transport
func (imp *Importer) sendToTransport(input <-chan FileData, output chan<- FileData) {
	defer close(output)

	for file := range input {
		// Convert to transport.FileData
		transportFile := FileData{
			Path:    imp.makeRelativePath(file.Path),
			Content: file.Content,
			Hash:    file.Hash,
			Size:    file.Size,
			ModTime: file.ModTime,
			Cached:  file.Cached,
		}

		// Add to batch
		full := imp.batcher.Add(transportFile)

		// Send batch if full
		if full {
			if err := imp.flushBatch(); err != nil {
				imp.errors.Add(file.Path, ErrorTypeTransport, fmt.Errorf("transport error: %w", err))
			}
		}

		// Forward to output channel
		output <- file
	}

	// Flush remaining files
	if imp.batcher.Count() > 0 {
		if err := imp.flushBatch(); err != nil {
			// Log but don't fail the import
			fmt.Printf("Warning: Failed to send final batch: %v\n", err)
		}
	}
}

// flushBatch sends the current batch via transport and waits for completion
func (imp *Importer) flushBatch() error {
	batch := imp.batcher.Flush()
	if batch == nil {
		return nil
	}

	// Send batch - returns Operation handle
	op, err := imp.transport.Send(imp.ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	// Wait for operation to complete (poll every 2 seconds)
	finalOp, err := imp.transport.WaitForCompletion(imp.ctx, op.Name, 2*time.Second)
	if err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	// Check if operation succeeded (done=true with result)
	if !finalOp.Done {
		return fmt.Errorf("operation did not complete")
	}

	if finalOp.Result == nil {
		return fmt.Errorf("operation completed without result")
	}

	// Check for failed files in the result
	if finalOp.Result.Failed > 0 {
		// Log failures but don't fail the entire batch
		fmt.Printf("Warning: %d/%d files failed in batch %s\n",
			finalOp.Result.Failed, finalOp.Result.TotalFiles, finalOp.Result.BatchID)

		// Optionally log individual errors
		for _, err := range finalOp.Result.Errors {
			fmt.Printf("  %s: %s\n", err.File, err.Message)
		}
	}

	// Success - batch was processed (even if some files failed)
	return nil
}

// makeRelativePath converts absolute path to relative path from root dir
func (imp *Importer) makeRelativePath(absPath string) string {
	relPath, err := filepath.Rel(imp.rootDir, absPath)
	if err != nil {
		// If we can't make it relative, return absolute path
		return absPath
	}
	return relPath
}

// GetResult returns the final import result after completion
func (imp *Importer) GetResult() *ImportResult {
	stats := imp.progress.GetStats()
	return &ImportResult{
		Processed: int(stats.Processed),
		Failed:    int(stats.Failed),
		Errors:    imp.errors.GetAll(),
		Duration:  float64(stats.StartTime.Unix()),
	}
}

// Stop cancels the import operation and cleans up resources
func (imp *Importer) Stop() {
	imp.cancel()
	imp.progress.Stop()

	// Close fingerprint cache
	if imp.fingerprintCache != nil {
		if err := imp.fingerprintCache.Close(); err != nil {
			fmt.Printf("Warning: Failed to close cache: %v\n", err)
		}
	}

	// Close transport
	if imp.transport != nil {
		if err := imp.transport.Close(); err != nil {
			fmt.Printf("Warning: Failed to close transport: %v\n", err)
		}
	}
}

// GetProgress returns current progress statistics
func (imp *Importer) GetProgress() Stats {
	return imp.progress.GetStats()
}

// GetErrors returns all recorded errors
func (imp *Importer) GetErrors() map[string]error {
	return imp.errors.GetAll()
}

// PrintSummary outputs the final summary report
func (imp *Importer) PrintSummary() {
	imp.progress.PrintFinal()

	// Print error details if any
	if imp.errors.HasErrors() {
		errorMap := imp.errors.GetAll()
		errorCount := len(errorMap)

		fmt.Printf("\nErrors (%d):\n", errorCount)

		// Show first 10 errors
		shown := 0
		for path, err := range errorMap {
			if shown >= 10 {
				remaining := errorCount - shown
				fmt.Printf("  ... and %d more errors\n", remaining)
				break
			}
			fmt.Printf("  %s: %v\n", path, err)
			shown++
		}
	}
}
