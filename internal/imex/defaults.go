package imex

import "time"

// Default constants for imex package
const (
	DefaultPathChanBuffer = 100
)

var (
	DefaultExtensions             = []string{".md"}
	DefaultMaxFileSize      int64 = 2 * 1024 * 1024 // 2 MiB
	DefaultProgressInterval       = 200 * time.Millisecond
	DefaultCollectionName         = "imports"
)

// applyDefaults fills unset ImportOptions fields with sensible defaults.
func applyDefaults(opts *ImportOptions) {
	if len(opts.Extensions) == 0 {
		opts.Extensions = append([]string(nil), DefaultExtensions...)
	}
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = DefaultMaxFileSize
	}
	if opts.Incremental == nil {
		t := true
		opts.Incremental = &t
	}
	if opts.PathChanBuffer == 0 {
		opts.PathChanBuffer = DefaultPathChanBuffer
	}
	if opts.ProgressInterval == 0 {
		opts.ProgressInterval = DefaultProgressInterval
	}
	if opts.Collection == "" {
		opts.Collection = DefaultCollectionName
	}
}
