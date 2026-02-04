package imex

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"sync"
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
	MaxFileSize    int64    // Maximum file size in bytes (0 = use default 2MB)
	FollowSymlinks bool     // Whether to follow symbolic links
	IncludeHidden  bool     // Whether to include hidden files (starting with .)
	Incremental    *bool    // Enable fingerprint caching for incremental imports (nil = use default)
	ClearCache     bool     // Clear fingerprint cache before import

	// Optional: If set, imported files will be sent to the transport
	Transport Transport // Transport to send files to (e.g., HTTP to Mind)
	Batcher   Batcher   // Batcher for grouping files (if nil, creates default)
	// Internal tuning knobs (optional)
	PathChanBuffer   int           // Buffer size for path/file channels (defaulted)
	ProgressInterval time.Duration // Progress reporting interval (defaulted)
	// Collection prefix on the server to place imported files under.
	// If empty, a default "imports" prefix will be used and collections will be
	// derived from the first path segment under the import root.
	Collection string
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
	// summary aggregation
	summaryMu     sync.Mutex
	perCollection map[string]CollectionSummary
}

// recordFileForSummary updates per-collection aggregates for a processed file
func (imp *Importer) recordFileForSummary(f FileData) {
	col := deriveCollection(imp.rootDir, f.Path, imp.opts.Collection)
	imp.summaryMu.Lock()
	defer imp.summaryMu.Unlock()
	if imp.perCollection == nil {
		imp.perCollection = make(map[string]CollectionSummary)
	}
	s := imp.perCollection[col]
	s.Files++
	s.Bytes += f.Size
	imp.perCollection[col] = s
}

func NewImporter(opts ImportOptions) *Importer {
	// Centralize and apply defaults
	applyDefaults(&opts)

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize fingerprint cache if incremental import is enabled
	var cache *FingerprintCache
	if opts.Incremental != nil && *opts.Incremental {
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
		cacheFingerprint = importFlags.Bool("cache-fingerprint", true, "Enable fingerprinting for caching")
		followSymlinks   = importFlags.Bool("follow-symlinks", false, "Follow symlinks during import")
		yes              = importFlags.Bool("yes", false, "Assume yes for prompts (non-interactive)")
		dryRunFlag       = importFlags.Bool("dry-run", false, "Print summary and exit without importing")
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
	fmt.Printf("[import] src=%s, collection=%s, cacheFingerprint=%v, followSymlinks=%v, dryRun=%v, yes=%v, config=%s\n",
		*src, *collection, *cacheFingerprint, *followSymlinks, *dryRunFlag || dryRun, *yes, config,
	)

	// Pass the flags to the Importer and execute the import process
	// Build options for summary and import
	opts := ImportOptions{
		Dir:            *src,
		FollowSymlinks: *followSymlinks,
		Incremental:    cacheFingerprint,
		Transport:      nil, // Configure transport as needed
		Collection:     *collection,
	}

	// Compute and print summary
	summary, err := ComputeImportSummaryFromOptions(opts)
	if err != nil {
		fmt.Printf("Failed to compute import summary: %v\n", err)
		return
	}

	fmt.Println("\nIMPORT SUMMARY")
	fmt.Printf("  - Total files: %d\n", summary.TotalFiles)
	fmt.Printf("  - Total size: %.2f MB\n", float64(summary.TotalBytes)/(1024*1024))
	if len(summary.Collections) > 0 {
		fmt.Println("Collections:")
		for col, s := range summary.Collections {
			fmt.Printf("  - %s: %d files, %.2f MB\n", col, s.Files, float64(s.Bytes)/(1024*1024))
		}
	}

	// If dry-run flag, exit after printing summary
	if *dryRunFlag || dryRun {
		return
	}

	// Prompt user unless --yes provided or not a TTY
	if !*yes {
		fmt.Printf("Proceed with import? [Y/n]: ")
		var resp string
		if _, err := fmt.Scanln(&resp); err != nil {
			fmt.Println("No input; aborting. Use --yes to skip prompt in scripts.")
			return
		}
		if resp != "" && (resp[0] == 'n' || resp[0] == 'N') {
			fmt.Println("Aborted by user.")
			return
		}
	}

	importer := NewImporter(opts)
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
	// Start progress reporting with configured interval
	imp.progress.StartReporting(imp.opts.ProgressInterval)

	// Phase 1: Walk directory and discover files
	paths := make(chan string, imp.opts.PathChanBuffer)
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

	// If transport is not configured, wrap results so we can record summaries
	if imp.transport == nil {
		out := make(chan FileData, imp.opts.PathChanBuffer)
		go func() {
			defer close(out)
			for f := range results {
				imp.recordFileForSummary(f)
				out <- f
			}
		}()
		results = out
	}

	// Phase 3: If transport configured, batch and send files
	if imp.transport != nil {
		output := make(chan FileData, imp.opts.PathChanBuffer)
		go imp.sendToTransport(results, output)
		return output
	}

	return results
}

// sendToTransport batches and sends files to the configured transport
func (imp *Importer) sendToTransport(input <-chan FileData, output chan<- FileData) {
	defer close(output)

	for file := range input {
		// Record per-collection summary
		imp.recordFileForSummary(file)

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

// ComputeImportSummary returns an ImportSummary combining progress snapshot and
// per-collection aggregates recorded during processing.
func (imp *Importer) ComputeImportSummary() ImportSummary {
	stats := imp.progress.GetStats()
	imp.summaryMu.Lock()
	defer imp.summaryMu.Unlock()
	// copy per-collection map
	cols := make(map[string]CollectionSummary, len(imp.perCollection))
	var totalBytes int64
	var totalFiles int64
	for k, v := range imp.perCollection {
		cols[k] = v
		totalBytes += v.Bytes
		totalFiles += int64(v.Files)
	}

	return ImportSummary{
		TotalFiles:  totalFiles,
		TotalBytes:  totalBytes,
		Processed:   stats.Processed,
		Skipped:     stats.Skipped,
		Failed:      stats.Failed,
		BytesRead:   stats.BytesRead,
		Collections: cols,
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

	// Print collection breakdown
	summary := imp.ComputeImportSummary()
	if len(summary.Collections) > 0 {
		fmt.Println("\nCollections:")
		for col, s := range summary.Collections {
			fmt.Printf("  - %s: %d files, %.2f MB\n", col, s.Files, float64(s.Bytes)/(1024*1024))
		}
	}

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
