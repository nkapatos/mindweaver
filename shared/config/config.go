// Package config provides configuration management for Mindweaver services.
//
// Configuration is loaded from environment variables with sensible defaults.
// Supports two deployment modes:
//   - Combined: All services run together (default)
//   - Standalone: Services run independently
package config

import (
	"os"
	"strconv"
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

// LoadConfig loads configuration from environment variables with defaults.
func LoadConfig(mode DeploymentMode) *Config {
	// Allow MODE to be overridden via env var
	if envMode := os.Getenv("MODE"); envMode != "" {
		if envMode == "standalone" || envMode == "combined" {
			mode = DeploymentMode(envMode)
		}
	}

	return &Config{
		Mode: mode,
		Mind: MindConfig{
			Port:   getEnvInt("MIND_PORT", 9421),
			DBPath: getEnv("MIND_DB_PATH", "db/mind.db"),
		},
		Brain: BrainConfig{
			Port:           getEnvInt("BRAIN_PORT", 9422),
			DBPath:         getEnv("BRAIN_DB_PATH", "db/brain.db"),
			MindServiceURL: getEnv("MIND_SERVICE_URL", "http://localhost:9421"),
			LLMEndpoint:    getEnv("LLM_ENDPOINT", "http://localhost:11434"),
			SmallModel:     getEnv("LLM_SMALL_MODEL", "phi3-mini"),
			BigModel:       getEnv("LLM_BIG_MODEL", "phi4"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "INFO"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Security: SecurityConfig{
			ETagSalt: getEnv("ETAG_SALT", generateRandomSalt()),
		},
	}
}

// GetCombinedPort returns the port to use in combined mode.
func (c *Config) GetCombinedPort() int {
	return getEnvInt("PORT", c.Mind.Port)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// NOTE: Currently not used, but can be enabled if boolean configs are needed
// func getEnvBool(key string, defaultValue bool) bool {
// 	if value := os.Getenv(key); value != "" {
// 		return value == "true" || value == "1" || value == "yes"
// 	}
// 	return defaultValue
// }

func generateRandomSalt() string {
	// Generate random salt for ETag hashing when ETAG_SALT is not set
	// WARNING: ETags will change on restart without persistent ETAG_SALT
	return "mindweaver-" + strconv.FormatInt(int64(os.Getpid()), 36) + "-" + strconv.FormatInt(int64(os.Getppid()), 36)
}
