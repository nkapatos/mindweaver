// Package config provides configuration management for Mindweaver services.
//
// Configuration is loaded using Viper with the following precedence (highest to lowest):
//  1. Environment variables (MW_ prefix)
//  2. Config file (config.yaml)
//  3. Built-in defaults
//
// Supports two deployment modes:
//   - Combined: All services run together (default)
//   - Standalone: Services run independently
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// DeploymentMode defines how services are deployed
type DeploymentMode string

const (
	// ModeStandalone runs a single service independently
	ModeStandalone DeploymentMode = "standalone"
	// ModeCombined runs all services together in one process
	ModeCombined DeploymentMode = "combined"
)

// Config holds all service configurations
type Config struct {
	Mode     DeploymentMode
	DataDir  string // Root directory for all data (databases, config)
	Mind     MindConfig
	Brain    BrainConfig
	Logging  LoggingConfig
	Security SecurityConfig
}

// MindConfig configures the Mind service (PKM/Notes)
type MindConfig struct {
	Port   int
	DBPath string
}

// BrainConfig configures the Brain service (AI Assistant)
type BrainConfig struct {
	Port           int
	DBPath         string
	BadgerDBPath   string // Path for TitleIndex BadgerDB (future use)
	MindServiceURL string // URL to Mind service (standalone mode only)
	LLMEndpoint    string
	SmallModel     string // Fast model for routing/classification
	BigModel       string // Powerful model for complex reasoning
}

// LoggingConfig configures structured logging
type LoggingConfig struct {
	Level  string // DEBUG, INFO, WARN, ERROR
	Format string // text or json
}

// SecurityConfig configures security settings
type SecurityConfig struct {
	ETagSalt string // Salt for ETag hashing (set for production to persist across restarts)
}

// setDefaults configures all default values in Viper.
// This is the single source of truth for configuration defaults.
func setDefaults(v *viper.Viper) {
	// Data directory - root for all persistent data
	v.SetDefault("data_dir", "./data")

	// Mind service defaults
	v.SetDefault("mind.port", 9421)
	v.SetDefault("mind.db_path", "") // Derived from data_dir if empty

	// Brain service defaults
	v.SetDefault("brain.port", 9422)
	v.SetDefault("brain.db_path", "")        // Derived from data_dir if empty
	v.SetDefault("brain.badger_db_path", "") // Derived from data_dir if empty
	v.SetDefault("brain.mind_service_url", "http://localhost:9421")
	v.SetDefault("brain.llm_endpoint", "http://localhost:11434")
	v.SetDefault("brain.small_model", "phi3-mini")
	v.SetDefault("brain.big_model", "phi4")

	// Logging defaults
	v.SetDefault("log.level", "INFO")
	v.SetDefault("log.format", "text")

	// Security defaults - empty means generate random salt
	v.SetDefault("security.etag_salt", "")
}

// configureEnvVars sets up environment variable binding with MW_ prefix.
func configureEnvVars(v *viper.Viper) {
	v.SetEnvPrefix("MW")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

// configureConfigFile sets up config file search paths.
func configureConfigFile(v *viper.Viper) {
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Search paths in order of priority
	v.AddConfigPath("/data")                                        // Docker container
	v.AddConfigPath("$HOME/.config/mindweaver")                     // XDG Linux
	v.AddConfigPath("$HOME/Library/Application Support/Mindweaver") // macOS
	v.AddConfigPath(".")                                            // Current directory (dev)
}

// LoadConfig loads configuration from environment variables and config files.
// The mode parameter sets the initial deployment mode, which can be overridden
// by the MW_MODE environment variable.
func LoadConfig(mode DeploymentMode) (*Config, error) {
	v := viper.New()

	// 1. Set defaults (single source of truth)
	setDefaults(v)

	// 2. Configure config file locations
	configureConfigFile(v)

	// 3. Try to read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is OK - we have defaults
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Config file was found but has errors
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// 4. Configure environment variables (highest priority)
	configureEnvVars(v)

	// 5. Build config struct
	return buildConfig(v, mode)
}

// buildConfig constructs the Config struct from Viper values.
func buildConfig(v *viper.Viper, mode DeploymentMode) (*Config, error) {
	// Allow MODE to be overridden via env var
	if envMode := v.GetString("mode"); envMode != "" {
		if envMode == "standalone" || envMode == "combined" {
			mode = DeploymentMode(envMode)
		}
	}

	dataDir := v.GetString("data_dir")

	// Derive paths from data_dir if not explicitly set
	mindDBPath := v.GetString("mind.db_path")
	if mindDBPath == "" {
		mindDBPath = filepath.Join(dataDir, "mind.db")
	}

	brainDBPath := v.GetString("brain.db_path")
	if brainDBPath == "" {
		brainDBPath = filepath.Join(dataDir, "brain.db")
	}

	badgerDBPath := v.GetString("brain.badger_db_path")
	if badgerDBPath == "" {
		badgerDBPath = filepath.Join(dataDir, "badger")
	}

	// Generate ETag salt if not provided
	etagSalt := v.GetString("security.etag_salt")
	if etagSalt == "" {
		etagSalt = generateRandomSalt()
	}

	cfg := &Config{
		Mode:    mode,
		DataDir: dataDir,
		Mind: MindConfig{
			Port:   v.GetInt("mind.port"),
			DBPath: mindDBPath,
		},
		Brain: BrainConfig{
			Port:           v.GetInt("brain.port"),
			DBPath:         brainDBPath,
			BadgerDBPath:   badgerDBPath,
			MindServiceURL: v.GetString("brain.mind_service_url"),
			LLMEndpoint:    v.GetString("brain.llm_endpoint"),
			SmallModel:     v.GetString("brain.small_model"),
			BigModel:       v.GetString("brain.big_model"),
		},
		Logging: LoggingConfig{
			Level:  v.GetString("log.level"),
			Format: v.GetString("log.format"),
		},
		Security: SecurityConfig{
			ETagSalt: etagSalt,
		},
	}

	return cfg, nil
}

// GetCombinedPort returns the port to use in combined mode.
// It checks for a PORT override first, then falls back to Mind port.
func (c *Config) GetCombinedPort() int {
	if port := os.Getenv("MW_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			return p
		}
	}
	return c.Mind.Port
}

// Validate checks that the configuration is valid and usable.
// It ensures data directories exist and are writable.
func (c *Config) Validate() error {
	// Ensure data directory exists or can be created
	if err := os.MkdirAll(c.DataDir, 0o755); err != nil {
		return fmt.Errorf("cannot create data directory %s: %w", c.DataDir, err)
	}

	// Ensure parent directories for DB files exist
	if err := os.MkdirAll(filepath.Dir(c.Mind.DBPath), 0o755); err != nil {
		return fmt.Errorf("cannot create directory for mind database: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(c.Brain.DBPath), 0o755); err != nil {
		return fmt.Errorf("cannot create directory for brain database: %w", err)
	}

	// Ensure badger directory exists
	if err := os.MkdirAll(c.Brain.BadgerDBPath, 0o755); err != nil {
		return fmt.Errorf("cannot create badger database directory: %w", err)
	}

	// Test writability by creating a temp file
	testFile := filepath.Join(c.DataDir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("data directory %s is not writable: %w", c.DataDir, err)
	}
	closeErr := f.Close()
	removeErr := os.Remove(testFile)
	if closeErr != nil {
		return fmt.Errorf("failed to close test file: %w", closeErr)
	}
	if removeErr != nil {
		return fmt.Errorf("failed to remove test file: %w", removeErr)
	}

	return nil
}

// generateRandomSalt creates a random salt for ETag hashing.
// WARNING: ETags will change on restart without persistent MW_SECURITY_ETAG_SALT.
func generateRandomSalt() string {
	return fmt.Sprintf("mindweaver-%d-%d", os.Getpid(), os.Getppid())
}
