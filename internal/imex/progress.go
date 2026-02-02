package imex

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Stats tracks import statistics
type Stats struct {
	TotalFiles int64 // Total files discovered
	Processed  int64 // Files successfully processed
	Failed     int64 // Files that failed
	Skipped    int64 // Files skipped (cache hits)
	BytesRead  int64 // Total bytes read
	StartTime  time.Time
}

// ProgressTracker manages real-time progress reporting
// Thread-safe for concurrent updates
type ProgressTracker struct {
	stats        *Stats
	lastUpdate   atomic.Value // time.Time
	updateTicker *time.Ticker
	done         chan struct{}
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	pt := &ProgressTracker{
		stats: &Stats{
			StartTime: time.Now(),
		},
		done: make(chan struct{}),
	}
	pt.lastUpdate.Store(time.Time{})
	return pt
}

// SetTotal sets the total number of files to process
func (pt *ProgressTracker) SetTotal(total int64) {
	atomic.StoreInt64(&pt.stats.TotalFiles, total)
}

// IncrementProcessed increments the processed file count
func (pt *ProgressTracker) IncrementProcessed() {
	atomic.AddInt64(&pt.stats.Processed, 1)
}

// IncrementFailed increments the failed file count
func (pt *ProgressTracker) IncrementFailed() {
	atomic.AddInt64(&pt.stats.Failed, 1)
}

// IncrementSkipped increments the skipped file count (cache hits)
func (pt *ProgressTracker) IncrementSkipped() {
	atomic.AddInt64(&pt.stats.Skipped, 1)
}

// AddBytesRead adds to the total bytes read
func (pt *ProgressTracker) AddBytesRead(bytes int64) {
	atomic.AddInt64(&pt.stats.BytesRead, bytes)
}

// GetStats returns a snapshot of current statistics
func (pt *ProgressTracker) GetStats() Stats {
	return Stats{
		TotalFiles: atomic.LoadInt64(&pt.stats.TotalFiles),
		Processed:  atomic.LoadInt64(&pt.stats.Processed),
		Failed:     atomic.LoadInt64(&pt.stats.Failed),
		Skipped:    atomic.LoadInt64(&pt.stats.Skipped),
		BytesRead:  atomic.LoadInt64(&pt.stats.BytesRead),
		StartTime:  pt.stats.StartTime,
	}
}

// StartReporting begins periodic progress reporting to stdout
// Updates every interval until Stop is called
func (pt *ProgressTracker) StartReporting(interval time.Duration) {
	pt.updateTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-pt.updateTicker.C:
				pt.printProgress()
			case <-pt.done:
				return
			}
		}
	}()
}

// Stop halts progress reporting
func (pt *ProgressTracker) Stop() {
	if pt.updateTicker != nil {
		pt.updateTicker.Stop()
	}
	close(pt.done)
}

// printProgress outputs current progress to stdout
// Uses carriage return to update same line
func (pt *ProgressTracker) printProgress() {
	stats := pt.GetStats()
	elapsed := time.Since(stats.StartTime).Seconds()

	var percent float64
	if stats.TotalFiles > 0 {
		completed := stats.Processed + stats.Failed + stats.Skipped
		percent = float64(completed) / float64(stats.TotalFiles) * 100
	}

	rate := float64(stats.Processed+stats.Skipped) / elapsed

	fmt.Printf("\r\033[K") // Clear line
	if stats.Skipped > 0 {
		fmt.Printf("Progress: %d/%d (%.1f%%) | Changed: %d | Unchanged: %d | Errors: %d | Rate: %.1f files/s",
			stats.Processed+stats.Failed+stats.Skipped,
			stats.TotalFiles,
			percent,
			stats.Processed,
			stats.Skipped,
			stats.Failed,
			rate,
		)
	} else {
		fmt.Printf("Progress: %d/%d (%.1f%%) | Errors: %d | Rate: %.1f files/s | Elapsed: %.1fs",
			stats.Processed+stats.Failed,
			stats.TotalFiles,
			percent,
			stats.Failed,
			rate,
			elapsed,
		)
	}
}

// PrintFinal outputs the final summary report
func (pt *ProgressTracker) PrintFinal() {
	fmt.Println() // New line after progress updates
	stats := pt.GetStats()
	duration := time.Since(stats.StartTime).Seconds()

	fmt.Println("\n✓ Import complete")

	if stats.Skipped > 0 {
		// Incremental mode with cache stats
		total := stats.Processed + stats.Skipped + stats.Failed
		unchangedPct := float64(stats.Skipped) / float64(total) * 100
		fmt.Printf("  - Total files: %d\n", total)
		fmt.Printf("  - Changed: %d files (%.1f%%)\n", stats.Processed, 100-unchangedPct)
		fmt.Printf("  - Unchanged: %d files (%.1f%%)\n", stats.Skipped, unchangedPct)
		fmt.Printf("  - Errors: %d files\n", stats.Failed)
	} else {
		// Normal mode
		fmt.Printf("  - Processed: %d files\n", stats.Processed)
		fmt.Printf("  - Errors: %d files\n", stats.Failed)
	}

	fmt.Printf("  - Duration: %.2fs\n", duration)

	if stats.Processed+stats.Skipped > 0 {
		rate := float64(stats.Processed+stats.Skipped) / duration
		fmt.Printf("  - Rate: %.1f files/sec", rate)
		if stats.Skipped > 0 {
			fmt.Printf(" (cache accelerated)")
		}
		fmt.Println()
	}

	if stats.BytesRead > 0 {
		mb := float64(stats.BytesRead) / (1024 * 1024)
		fmt.Printf("  - Data: %.2f MB\n", mb)
	}
}
