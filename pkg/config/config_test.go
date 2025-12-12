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
	if cfg.Mind.Port != 8080 {
		t.Errorf("Expected mind port 8080, got %d", cfg.Mind.Port)
	}
	if cfg.Mind.DBPath != "db/mind.db" {
		t.Errorf("Expected mind DB path db/mind.db, got %s", cfg.Mind.DBPath)
	}
	if !cfg.Mind.Enabled {
		t.Error("Expected mind to be enabled by default")
	}

	// Verify brain config
	if cfg.Brain.Port != 8081 {
		t.Errorf("Expected brain port 8081, got %d", cfg.Brain.Port)
	}
	if cfg.Brain.DBPath != "db/brain.db" {
		t.Errorf("Expected brain DB path db/brain.db, got %s", cfg.Brain.DBPath)
	}
	if cfg.Brain.MindServiceURL != "http://localhost:8080" {
		t.Errorf("Expected brain mind service URL http://localhost:8080, got %s", cfg.Brain.MindServiceURL)
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
	if cfg.GetCombinedPort() != 8080 {
		t.Errorf("Expected combined port 8080, got %d", cfg.GetCombinedPort())
	}

	// Both services should be enabled
	if !cfg.Mind.Enabled {
		t.Error("Expected mind to be enabled in combined mode")
	}
	if !cfg.Brain.Enabled {
		t.Error("Expected brain to be enabled in combined mode")
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
	os.Setenv("MIND_ENABLED", "false")
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
	if cfg.Mind.Enabled {
		t.Error("Expected mind to be disabled")
	}
}

// TestBackwardCompatibility verifies old env vars still work
func TestBackwardCompatibility(t *testing.T) {
	// Set old environment variable names
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("ASSISTANT_PORT", "9091")
	os.Setenv("SERVER_DB_PATH", "/old/server.db")
	os.Setenv("ASSISTANT_DB_PATH", "/old/assistant.db")
	defer clearEnv()

	cfg := LoadConfig(ModeStandalone)

	// Verify old names still work
	if cfg.Mind.Port != 9090 {
		t.Errorf("Expected mind port 9090 (from SERVER_PORT), got %d", cfg.Mind.Port)
	}
	if cfg.Brain.Port != 9091 {
		t.Errorf("Expected brain port 9091 (from ASSISTANT_PORT), got %d", cfg.Brain.Port)
	}
	if cfg.Mind.DBPath != "/old/server.db" {
		t.Errorf("Expected mind DB path /old/server.db (from SERVER_DB_PATH), got %s", cfg.Mind.DBPath)
	}
	if cfg.Brain.DBPath != "/old/assistant.db" {
		t.Errorf("Expected brain DB path /old/assistant.db (from ASSISTANT_DB_PATH), got %s", cfg.Brain.DBPath)
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

// TestBooleanParsing verifies various boolean value formats
func TestBooleanParsing(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"", false}, // Empty defaults to false when default is false
	}

	for _, tc := range testCases {
		os.Setenv("LSP_ENABLED", tc.value)
		cfg := LoadConfig(ModeStandalone)
		if cfg.LSP.Enabled != tc.expected {
			t.Errorf("For LSP_ENABLED=%q, expected %v, got %v", tc.value, tc.expected, cfg.LSP.Enabled)
		}
		os.Unsetenv("LSP_ENABLED")
	}
}

// Example: Using config in standalone mind mode
func ExampleLoadConfig_standaloneMind() {
	cfg := LoadConfig(ModeStandalone)

	// Use mind configuration
	_ = cfg.Mind.Port   // 8080
	_ = cfg.Mind.DBPath // "db/mind.db"

	// Brain config exists but won't be used
	_ = cfg.Brain.Enabled // true, but service won't start

	// Output:
}

// Example: Using config in standalone brain mode
func ExampleLoadConfig_standaloneBrain() {
	cfg := LoadConfig(ModeStandalone)

	// Use brain configuration
	_ = cfg.Brain.Port           // 8081
	_ = cfg.Brain.DBPath         // "db/brain.db"
	_ = cfg.Brain.MindServiceURL // "http://localhost:8080"

	// Mind config exists but won't be used
	_ = cfg.Mind.Enabled // true, but service won't start

	// Output:
}

// Example: Using config in combined mode
func ExampleLoadConfig_combined() {
	cfg := LoadConfig(ModeCombined)

	// Both services enabled on same port
	port := cfg.GetCombinedPort() // 8080 (single port for both)

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
		"MIND_ENABLED",
		"BRAIN_ENABLED",
		"LSP_ENABLED",
		"LOG_LEVEL",
		"LOG_FORMAT",
		"PORT",
		"MODE",
		// Old names for backward compatibility
		"SERVER_PORT",
		"ASSISTANT_PORT",
		"SERVER_DB_PATH",
		"ASSISTANT_DB_PATH",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
