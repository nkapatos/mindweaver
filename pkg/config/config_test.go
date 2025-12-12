package config

import (
	"os"
	"testing"
)

// TestStandaloneMode verifies configuration in standalone mode
func TestStandaloneMode(t *testing.T) {
	// Clear environment to use defaults
	clearEnv()

	cfg := LoadConfig(ModeStandalone)

	// Verify mode
	if cfg.Mode != ModeStandalone {
		t.Errorf("Expected mode %s, got %s", ModeStandalone, cfg.Mode)
	}

	// Verify mind config
	if cfg.Mind.Port != 9421 {
		t.Errorf("Expected mind port 9421, got %d", cfg.Mind.Port)
	}
	if cfg.Mind.DBPath != "db/mind.db" {
		t.Errorf("Expected mind DB path db/mind.db, got %s", cfg.Mind.DBPath)
	}

	// Verify brain config
	if cfg.Brain.Port != 9422 {
		t.Errorf("Expected brain port 9422, got %d", cfg.Brain.Port)
	}
	if cfg.Brain.DBPath != "db/brain.db" {
		t.Errorf("Expected brain DB path db/brain.db, got %s", cfg.Brain.DBPath)
	}
	if cfg.Brain.MindServiceURL != "http://localhost:9421" {
		t.Errorf("Expected brain mind service URL http://localhost:9421, got %s", cfg.Brain.MindServiceURL)
	}

	// Verify logging config
	if cfg.Logging.Level != "INFO" {
		t.Errorf("Expected log level INFO, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Expected log format text, got %s", cfg.Logging.Format)
	}
}

// TestCombinedMode verifies configuration in combined mode
func TestCombinedMode(t *testing.T) {
	// Clear environment to use defaults
	clearEnv()

	cfg := LoadConfig(ModeCombined)

	// Verify mode
	if cfg.Mode != ModeCombined {
		t.Errorf("Expected mode %s, got %s", ModeCombined, cfg.Mode)
	}

	// Combined port should use mind port by default
	if cfg.GetCombinedPort() != 9421 {
		t.Errorf("Expected combined port 9421, got %d", cfg.GetCombinedPort())
	}
}

// TestEnvironmentVariables verifies that env vars override defaults
func TestEnvironmentVariables(t *testing.T) {
	// Set custom environment variables
	os.Setenv("MIND_PORT", "9000")
	os.Setenv("BRAIN_PORT", "9001")
	os.Setenv("MIND_DB_PATH", "/custom/mind.db")
	os.Setenv("BRAIN_DB_PATH", "/custom/brain.db")
	os.Setenv("MIND_SERVICE_URL", "http://custom:9000")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_FORMAT", "json")
	defer clearEnv()

	cfg := LoadConfig(ModeStandalone)

	// Verify overrides
	if cfg.Mind.Port != 9000 {
		t.Errorf("Expected mind port 9000, got %d", cfg.Mind.Port)
	}
	if cfg.Brain.Port != 9001 {
		t.Errorf("Expected brain port 9001, got %d", cfg.Brain.Port)
	}
	if cfg.Mind.DBPath != "/custom/mind.db" {
		t.Errorf("Expected mind DB path /custom/mind.db, got %s", cfg.Mind.DBPath)
	}
	if cfg.Brain.DBPath != "/custom/brain.db" {
		t.Errorf("Expected brain DB path /custom/brain.db, got %s", cfg.Brain.DBPath)
	}
	if cfg.Brain.MindServiceURL != "http://custom:9000" {
		t.Errorf("Expected brain mind service URL http://custom:9000, got %s", cfg.Brain.MindServiceURL)
	}
	if cfg.Logging.Level != "DEBUG" {
		t.Errorf("Expected log level DEBUG, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format json, got %s", cfg.Logging.Format)
	}
}

// TestCombinedPortOverride verifies PORT env var overrides in combined mode
func TestCombinedPortOverride(t *testing.T) {
	os.Setenv("PORT", "3000")
	defer clearEnv()

	cfg := LoadConfig(ModeCombined)

	if cfg.GetCombinedPort() != 3000 {
		t.Errorf("Expected combined port 3000, got %d", cfg.GetCombinedPort())
	}
}

// Example: Using config in standalone mind mode
func ExampleLoadConfig_standaloneMind() {
	cfg := LoadConfig(ModeStandalone)

	// Use mind configuration
	_ = cfg.Mind.Port   // 9421
	_ = cfg.Mind.DBPath // "db/mind.db"

	// Output:
}

// Example: Using config in standalone brain mode
func ExampleLoadConfig_standaloneBrain() {
	cfg := LoadConfig(ModeStandalone)

	// Use brain configuration
	_ = cfg.Brain.Port           // 9422
	_ = cfg.Brain.DBPath         // "db/brain.db"
	_ = cfg.Brain.MindServiceURL // "http://localhost:9421"

	// Output:
}

// Example: Using config in combined mode
func ExampleLoadConfig_combined() {
	cfg := LoadConfig(ModeCombined)

	// Both services enabled on same port
	port := cfg.GetCombinedPort() // 9421 (single port for both)

	_ = port
	// Mind routes: /api/notes, /api/tags, /api/templates
	// Brain routes: /api/assistants, /api/chat, /api/conversations

	// Both use same Echo instance on same port
	// with naturally separated routes (no prefix conflicts)

	// Output:
}

// Example: Using config with environment overrides
func ExampleLoadConfig_withEnvOverrides() {
	// Set custom configuration via environment
	os.Setenv("MIND_PORT", "9090")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_FORMAT", "json")
	defer func() {
		os.Unsetenv("MIND_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
	}()

	cfg := LoadConfig(ModeStandalone)

	_ = cfg.Mind.Port      // 9090 (overridden)
	_ = cfg.Logging.Level  // "DEBUG" (overridden)
	_ = cfg.Logging.Format // "json" (overridden)

	// Output:
}

// Helper function to clear environment variables
func clearEnv() {
	envVars := []string{
		"MIND_PORT",
		"BRAIN_PORT",
		"MIND_DB_PATH",
		"BRAIN_DB_PATH",
		"MIND_SERVICE_URL",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"PORT",
		"MODE",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
