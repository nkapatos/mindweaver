package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStandaloneMode verifies configuration in standalone mode
func TestStandaloneMode(t *testing.T) {
	// Clear environment to use defaults
	clearEnv()

	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify mode
	if cfg.Mode != ModeStandalone {
		t.Errorf("Expected mode %s, got %s", ModeStandalone, cfg.Mode)
	}

	// Verify mind config
	if cfg.Mind.Port != 9421 {
		t.Errorf("Expected mind port 9421, got %d", cfg.Mind.Port)
	}
	// DB path should be derived from data_dir
	expectedMindDB := filepath.Join("./data", "mind.db")
	if cfg.Mind.DBPath != expectedMindDB {
		t.Errorf("Expected mind DB path %s, got %s", expectedMindDB, cfg.Mind.DBPath)
	}

	// Verify brain config
	if cfg.Brain.Port != 9422 {
		t.Errorf("Expected brain port 9422, got %d", cfg.Brain.Port)
	}
	expectedBrainDB := filepath.Join("./data", "brain.db")
	if cfg.Brain.DBPath != expectedBrainDB {
		t.Errorf("Expected brain DB path %s, got %s", expectedBrainDB, cfg.Brain.DBPath)
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

	// Verify data dir
	if cfg.DataDir != "./data" {
		t.Errorf("Expected data dir ./data, got %s", cfg.DataDir)
	}
}

// TestCombinedMode verifies configuration in combined mode
func TestCombinedMode(t *testing.T) {
	// Clear environment to use defaults
	clearEnv()

	cfg, err := LoadConfig(ModeCombined)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

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
	// Clear environment first
	clearEnv()

	// Set custom environment variables with MW_ prefix
	os.Setenv("MW_MIND_PORT", "9000")
	os.Setenv("MW_BRAIN_PORT", "9001")
	os.Setenv("MW_MIND_DB_PATH", "/custom/mind.db")
	os.Setenv("MW_BRAIN_DB_PATH", "/custom/brain.db")
	os.Setenv("MW_BRAIN_MIND_SERVICE_URL", "http://custom:9000")
	os.Setenv("MW_LOG_LEVEL", "DEBUG")
	os.Setenv("MW_LOG_FORMAT", "json")
	defer clearEnv()

	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

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

// TestCombinedPortOverride verifies MW_PORT env var overrides in combined mode
func TestCombinedPortOverride(t *testing.T) {
	clearEnv()
	os.Setenv("MW_PORT", "3000")
	defer clearEnv()

	cfg, err := LoadConfig(ModeCombined)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.GetCombinedPort() != 3000 {
		t.Errorf("Expected combined port 3000, got %d", cfg.GetCombinedPort())
	}
}

// TestDataDirDerivation verifies that DB paths are derived from data_dir
func TestDataDirDerivation(t *testing.T) {
	clearEnv()
	os.Setenv("MW_DATA_DIR", "/mydata")
	defer clearEnv()

	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Paths should be derived from data_dir
	if cfg.DataDir != "/mydata" {
		t.Errorf("Expected data dir /mydata, got %s", cfg.DataDir)
	}
	if cfg.Mind.DBPath != "/mydata/mind.db" {
		t.Errorf("Expected mind DB path /mydata/mind.db, got %s", cfg.Mind.DBPath)
	}
	if cfg.Brain.DBPath != "/mydata/brain.db" {
		t.Errorf("Expected brain DB path /mydata/brain.db, got %s", cfg.Brain.DBPath)
	}
	if cfg.Brain.BadgerDBPath != "/mydata/badger" {
		t.Errorf("Expected badger DB path /mydata/badger, got %s", cfg.Brain.BadgerDBPath)
	}
}

// TestExplicitPathsOverrideDataDir verifies that explicit DB paths override data_dir derivation
func TestExplicitPathsOverrideDataDir(t *testing.T) {
	clearEnv()
	os.Setenv("MW_DATA_DIR", "/mydata")
	os.Setenv("MW_MIND_DB_PATH", "/explicit/mind.db")
	defer clearEnv()

	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Explicit path should override derived path
	if cfg.Mind.DBPath != "/explicit/mind.db" {
		t.Errorf("Expected mind DB path /explicit/mind.db, got %s", cfg.Mind.DBPath)
	}
	// Brain path should still be derived
	if cfg.Brain.DBPath != "/mydata/brain.db" {
		t.Errorf("Expected brain DB path /mydata/brain.db, got %s", cfg.Brain.DBPath)
	}
}

// TestModeOverride verifies MW_MODE env var can override the mode parameter
func TestModeOverride(t *testing.T) {
	clearEnv()
	os.Setenv("MW_MODE", "combined")
	defer clearEnv()

	// Call with standalone, but env var says combined
	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Mode != ModeCombined {
		t.Errorf("Expected mode to be overridden to combined, got %s", cfg.Mode)
	}
}

// TestValidate verifies the Validate function creates directories
func TestValidate(t *testing.T) {
	clearEnv()

	// Use temp directory
	tmpDir := t.TempDir()
	os.Setenv("MW_DATA_DIR", tmpDir)
	defer clearEnv()

	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Validate should create directories
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Check that badger directory was created
	badgerDir := filepath.Join(tmpDir, "badger")
	if _, err := os.Stat(badgerDir); os.IsNotExist(err) {
		t.Errorf("Expected badger directory %s to be created", badgerDir)
	}
}

// TestLLMConfig verifies LLM configuration defaults
func TestLLMConfig(t *testing.T) {
	clearEnv()

	cfg, err := LoadConfig(ModeStandalone)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Brain.LLMEndpoint != "http://localhost:11434" {
		t.Errorf("Expected LLM endpoint http://localhost:11434, got %s", cfg.Brain.LLMEndpoint)
	}
	if cfg.Brain.SmallModel != "phi3-mini" {
		t.Errorf("Expected small model phi3-mini, got %s", cfg.Brain.SmallModel)
	}
	if cfg.Brain.BigModel != "phi4" {
		t.Errorf("Expected big model phi4, got %s", cfg.Brain.BigModel)
	}
}

// Helper function to clear environment variables
func clearEnv() {
	envVars := []string{
		// New MW_ prefix vars
		"MW_DATA_DIR",
		"MW_MIND_PORT",
		"MW_BRAIN_PORT",
		"MW_MIND_DB_PATH",
		"MW_BRAIN_DB_PATH",
		"MW_BRAIN_BADGER_DB_PATH",
		"MW_BRAIN_MIND_SERVICE_URL",
		"MW_BRAIN_LLM_ENDPOINT",
		"MW_BRAIN_SMALL_MODEL",
		"MW_BRAIN_BIG_MODEL",
		"MW_LOG_LEVEL",
		"MW_LOG_FORMAT",
		"MW_PORT",
		"MW_MODE",
		"MW_SECURITY_ETAG_SALT",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
