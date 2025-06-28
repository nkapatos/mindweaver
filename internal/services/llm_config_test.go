package services

import (
	"encoding/json"
	"testing"
)

func TestLLMConfiguration_DefaultConfiguration(t *testing.T) {
	config := DefaultConfiguration("gpt-4")

	if config.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", config.Model)
	}

	if config.Temperature == nil || *config.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", config.Temperature)
	}

	if config.MaxTokens == nil || *config.MaxTokens != 1000 {
		t.Errorf("Expected max_tokens 1000, got %v", config.MaxTokens)
	}

	if config.TopP == nil || *config.TopP != 1.0 {
		t.Errorf("Expected top_p 1.0, got %v", config.TopP)
	}
}

func TestLLMConfiguration_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *LLMConfiguration
		wantErr bool
	}{
		{
			name:    "valid configuration",
			config:  DefaultConfiguration("gpt-4"),
			wantErr: false,
		},
		{
			name: "missing model",
			config: &LLMConfiguration{
				Model: "",
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			config: &LLMConfiguration{
				Model:       "gpt-4",
				Temperature: float64Ptr(3.0),
			},
			wantErr: true,
		},
		{
			name: "invalid max_tokens",
			config: &LLMConfiguration{
				Model:     "gpt-4",
				MaxTokens: intPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "invalid top_p",
			config: &LLMConfiguration{
				Model: "gpt-4",
				TopP:  float64Ptr(1.5),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LLMConfiguration.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLLMConfiguration_ToJSON(t *testing.T) {
	config := DefaultConfiguration("gpt-4")

	jsonStr, err := config.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() error = %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("Generated JSON is invalid: %v", err)
	}

	// Verify required fields are present
	if parsed["model"] != "gpt-4" {
		t.Errorf("Expected model in JSON, got %v", parsed["model"])
	}
}

func TestFromJSON(t *testing.T) {
	jsonStr := `{
		"model": "claude-3-opus",
		"temperature": 0.8,
		"max_tokens": 2000,
		"top_p": 0.9
	}`

	config, err := FromJSON(jsonStr)
	if err != nil {
		t.Errorf("FromJSON() error = %v", err)
		return
	}

	if config.Model != "claude-3-opus" {
		t.Errorf("Expected model 'claude-3-opus', got '%s'", config.Model)
	}

	if config.Temperature == nil || *config.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %v", config.Temperature)
	}
}

// Helper functions for creating pointers
func float64Ptr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}
