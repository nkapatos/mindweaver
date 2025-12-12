// Package logging provides custom structured logging with module context and colors.
package logging

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

// Module identifiers
const (
	ModuleMind  = "mind"
	ModuleBrain = "brain"
	ModuleLSP   = "lsp"
	ModuleNVMW  = "nvmw" // Combined mode
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorBlue   = "\033[34m"
	ColorYellow = "\033[33m"
	ColorPurple = "\033[35m"
	ColorRed    = "\033[31m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// NewModuleLogger creates a logger with module context and optional colors.
// module: one of ModuleMind, ModuleBrain, ModuleLSP, ModuleNVMW
// level: log level (INFO, DEBUG, WARN, ERROR)
// format: "text" or "json"
func NewModuleLogger(module, level, format string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	case "INFO", "":
		logLevel = slog.LevelInfo
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: logLevel}

	if format == "json" {
		// JSON format: Add module as a field, no colors
		handler := &ModuleJSONHandler{
			module:      module,
			baseHandler: slog.NewJSONHandler(os.Stdout, opts),
		}
		return slog.New(handler)
	}

	// Text format: Custom handler with colored module prefix
	handler := NewModuleTextHandler(os.Stdout, module, opts)
	return slog.New(handler)
}

// ModuleTextHandler is a custom text handler that adds colored module prefix
type ModuleTextHandler struct {
	w      io.Writer
	module string
	color  string
	opts   *slog.HandlerOptions
	attrs  []slog.Attr
	mu     sync.Mutex
}

// NewModuleTextHandler creates a text handler with module prefix
func NewModuleTextHandler(w io.Writer, module string, opts *slog.HandlerOptions) *ModuleTextHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ModuleTextHandler{
		w:      w,
		module: module,
		color:  getModuleColor(module),
		opts:   opts,
	}
}

// Enabled implements slog.Handler
func (h *ModuleTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle implements slog.Handler with colored module prefix
func (h *ModuleTextHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	buf := &bytes.Buffer{}

	// Timestamp
	buf.WriteString("time=")
	buf.WriteString(r.Time.Format("2006-01-02T15:04:05.000Z07:00"))
	buf.WriteByte(' ')

	// Level with color
	levelColor := getLevelColor(r.Level)
	buf.WriteString("level=")
	buf.WriteString(levelColor)
	buf.WriteString(r.Level.String())
	buf.WriteString(ColorReset)
	buf.WriteByte(' ')

	// Module prefix with color
	buf.WriteString(h.color)
	buf.WriteString("[")
	buf.WriteString(h.module)
	buf.WriteString("]")
	buf.WriteString(ColorReset)
	buf.WriteByte(' ')

	// Message
	buf.WriteString("msg=")
	buf.WriteString(quote(r.Message))

	// Attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteByte(' ')
		buf.WriteString(a.Key)
		buf.WriteByte('=')
		buf.WriteString(a.Value.String())
		return true
	})

	buf.WriteByte('\n')
	_, err := h.w.Write(buf.Bytes())
	return err
}

// WithAttrs implements slog.Handler
func (h *ModuleTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ModuleTextHandler{
		w:      h.w,
		module: h.module,
		color:  h.color,
		opts:   h.opts,
		attrs:  append(h.attrs, attrs...),
	}
}

// WithGroup implements slog.Handler
func (h *ModuleTextHandler) WithGroup(name string) slog.Handler {
	// For simplicity, groups aren't fully implemented
	return h
}

// ModuleJSONHandler adds module field to JSON output
type ModuleJSONHandler struct {
	module      string
	baseHandler slog.Handler
}

// Enabled implements slog.Handler
func (h *ModuleJSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.baseHandler.Enabled(ctx, level)
}

// Handle implements slog.Handler - adds module field
func (h *ModuleJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	// Add module as first attribute
	r.AddAttrs(slog.String("module", h.module))
	return h.baseHandler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler
func (h *ModuleJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ModuleJSONHandler{
		module:      h.module,
		baseHandler: h.baseHandler.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler
func (h *ModuleJSONHandler) WithGroup(name string) slog.Handler {
	return &ModuleJSONHandler{
		module:      h.module,
		baseHandler: h.baseHandler.WithGroup(name),
	}
}

// Helper functions

// getModuleColor returns ANSI color code for a module
func getModuleColor(module string) string {
	switch module {
	case ModuleMind:
		return ColorGreen
	case ModuleBrain:
		return ColorBlue
	case ModuleLSP:
		return ColorYellow
	case ModuleNVMW:
		return ColorPurple
	default:
		return ColorReset
	}
}

// getLevelColor returns ANSI color for log level
func getLevelColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return ColorRed
	case level >= slog.LevelWarn:
		return ColorYellow
	case level >= slog.LevelInfo:
		return ColorCyan
	default:
		return ColorWhite // DEBUG
	}
}

// quote adds quotes if string contains spaces
func quote(s string) string {
	for _, c := range s {
		if c == ' ' || c == '\t' || c == '\n' {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}
