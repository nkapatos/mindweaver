// Package setup provides a first-run setup wizard for Mindweaver.
//
// The wizard is shown when no config.yaml exists in the data directory.
// It collects minimal configuration and generates a fully-commented config file.
package setup

import (
	"crypto/rand"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

//go:embed templates/*.html
var templateFS embed.FS

// Handler handles setup wizard requests.
type Handler struct {
	dataDir string
	logger  *slog.Logger
	tmpl    *template.Template
}

// WizardData contains data for rendering the wizard template.
type WizardData struct {
	DataDir       string
	Host          string
	Port          int
	LogLevel      string
	AllowExternal bool
	IsDocker      bool
	Error         string
	Success       bool
	ConfigPath    string
	GeneratedSalt string
}

// NewHandler creates a new setup handler.
func NewHandler(dataDir string, logger *slog.Logger) (*Handler, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Handler{
		dataDir: dataDir,
		logger:  logger.With("component", "setup-wizard"),
		tmpl:    tmpl,
	}, nil
}

// ConfigExists checks if config.yaml exists in the data directory.
func ConfigExists(dataDir string) bool {
	configPath := filepath.Join(dataDir, "config.yaml")
	_, err := os.Stat(configPath)
	return err == nil
}

// IsRunningInDocker detects if running inside a Docker container.
func IsRunningInDocker() bool {
	// Check for /.dockerenv file (most reliable)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	// Fallback: check cgroup (older Docker versions)
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker")
	}
	return false
}

// generateSecureSalt creates a cryptographically secure random salt.
func generateSecureSalt() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RegisterRoutes registers setup wizard routes on the Echo instance.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	admin := e.Group("/admin")
	admin.GET("/setup", h.HandleGetSetup)
	admin.POST("/setup", h.HandlePostSetup)
}

// HandleGetSetup renders the setup wizard form.
func (h *Handler) HandleGetSetup(c echo.Context) error {
	isDocker := IsRunningInDocker()

	// Generate salt for the form
	salt, err := generateSecureSalt()
	if err != nil {
		h.logger.Error("failed to generate salt", "error", err)
		salt = fmt.Sprintf("mindweaver-%d", time.Now().UnixNano())
	}

	// Default host based on environment
	defaultHost := "localhost"
	allowExternal := false
	if isDocker {
		defaultHost = "0.0.0.0"
		allowExternal = true
	}

	data := WizardData{
		DataDir:       h.dataDir,
		Host:          defaultHost,
		Port:          9421,
		LogLevel:      "WARN",
		AllowExternal: allowExternal,
		IsDocker:      isDocker,
		GeneratedSalt: salt,
	}

	return h.renderTemplate(c, http.StatusOK, data)
}

// HandlePostSetup processes the setup form and generates config.yaml.
func (h *Handler) HandlePostSetup(c echo.Context) error {
	isDocker := IsRunningInDocker()

	// Parse form values
	allowExternal := c.FormValue("allow_external") == "on"
	portStr := c.FormValue("port")
	logLevel := c.FormValue("log_level")
	etagSalt := c.FormValue("etag_salt")

	// Validate port
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return h.renderError(c, "Invalid port number. Must be between 1 and 65535.", isDocker)
	}

	// Validate log level
	validLevels := map[string]bool{"DEBUG": true, "INFO": true, "WARN": true, "ERROR": true}
	if !validLevels[logLevel] {
		return h.renderError(c, "Invalid log level.", isDocker)
	}

	// Determine host based on toggle
	host := "localhost"
	if allowExternal {
		host = "0.0.0.0"
	}

	// Generate salt if not provided
	if etagSalt == "" {
		etagSalt, err = generateSecureSalt()
		if err != nil {
			h.logger.Error("failed to generate salt", "error", err)
			etagSalt = fmt.Sprintf("mindweaver-%d", time.Now().UnixNano())
		}
	}

	// Generate config content
	configContent := generateConfigYAML(h.dataDir, host, port, logLevel, etagSalt)

	// Ensure data directory exists
	if err := os.MkdirAll(h.dataDir, 0o755); err != nil {
		h.logger.Error("failed to create data directory", "error", err, "path", h.dataDir)
		return h.renderError(c, fmt.Sprintf("Cannot create data directory: %s", err), isDocker)
	}

	// Write config file
	configPath := filepath.Join(h.dataDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		h.logger.Error("failed to write config file", "error", err, "path", configPath)
		return h.renderError(c, fmt.Sprintf("Cannot write config file: %s", err), isDocker)
	}

	h.logger.Info("setup wizard completed", "config_path", configPath)

	// Render success page
	data := WizardData{
		DataDir:       h.dataDir,
		Host:          host,
		Port:          port,
		LogLevel:      logLevel,
		AllowExternal: allowExternal,
		IsDocker:      isDocker,
		Success:       true,
		ConfigPath:    configPath,
	}

	return h.renderTemplate(c, http.StatusOK, data)
}

// renderTemplate renders the wizard template with the given data.
func (h *Handler) renderTemplate(c echo.Context, status int, data WizardData) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(status)
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "wizard.html", data)
}

// renderError re-renders the form with an error message.
func (h *Handler) renderError(c echo.Context, errMsg string, isDocker bool) error {
	salt, err := generateSecureSalt()
	if err != nil {
		h.logger.Error("failed to generate salt for error form", "error", err)
		salt = fmt.Sprintf("mindweaver-%d", time.Now().UnixNano())
	}
	data := WizardData{
		DataDir:       h.dataDir,
		Host:          "localhost",
		Port:          9421,
		LogLevel:      "WARN",
		AllowExternal: isDocker,
		IsDocker:      isDocker,
		Error:         errMsg,
		GeneratedSalt: salt,
	}
	return h.renderTemplate(c, http.StatusBadRequest, data)
}

// generateConfigYAML creates the fully-commented config.yaml content.
func generateConfigYAML(dataDir, host string, port int, logLevel, etagSalt string) string {
	timestamp := time.Now().Format(time.RFC3339)

	return fmt.Sprintf(`# Mindweaver Configuration
# Generated by setup wizard on %s
#
# This file configures Mindweaver server settings.
# Environment variables (MW_*) take precedence over these values.
# See: https://github.com/nkapatos/mindweaver for documentation.

# Data directory - root location for all databases and persistent data
# All database paths below are relative to this directory unless absolute paths are specified.
data_dir: %s

# Mind service configuration (PKM/Notes)
mind:
  # Host to bind to:
  # - "localhost" restricts access to this machine only (more secure)
  # - "0.0.0.0" allows connections from other devices (required for Docker)
  host: %s
  
  # Port for the Mind service
  port: %d
  
  # Database path (optional - defaults to $data_dir/mind.db)
  # db_path: /custom/path/mind.db

# Brain service configuration (AI Assistant)
# brain:
#   port: 9422
#   # db_path: /custom/path/brain.db
#   # badger_db_path: /custom/path/badger
#   
#   # LLM configuration (requires Ollama or compatible API)
#   # llm_endpoint: http://localhost:11434
#   # small_model: phi3-mini    # Fast model for routing/classification
#   # big_model: phi4           # Powerful model for complex reasoning

# Logging configuration
log:
  # Log level: DEBUG (verbose), INFO, WARN (recommended for production), ERROR (minimal)
  level: %s
  
  # Log format: "text" (human-readable) or "json" (structured, for log aggregators)
  format: text

# Security settings
security:
  # ETag salt for cache validation
  # This value should remain constant across restarts for consistent caching.
  # If changed, all client caches will be invalidated.
  etag_salt: %s
`, timestamp, dataDir, host, port, logLevel, etagSalt)
}

// SetupRequiredMiddleware redirects to /admin/setup if config.yaml doesn't exist.
// It skips redirection for health checks and setup routes themselves.
func SetupRequiredMiddleware(dataDir string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path

			// Always allow these paths without config
			if path == "/health" ||
				strings.HasPrefix(path, "/admin/setup") {
				return next(c)
			}

			// Check if config exists
			if !ConfigExists(dataDir) {
				return c.Redirect(http.StatusTemporaryRedirect, "/admin/setup")
			}

			return next(c)
		}
	}
}
