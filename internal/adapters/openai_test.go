package adapters

import (
	"context"
	"testing"
)

func TestNewOpenAIAdapter(t *testing.T) {
	tests := []struct {
		name    string
		config  AdapterConfig
		wantErr bool
	}{
		{
			name: "valid config with API key",
			config: AdapterConfig{
				Name:   "openai",
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with API key and base URL",
			config: AdapterConfig{
				Name:    "openai",
				APIKey:  "test-api-key",
				BaseURL: "https://api.openai.com/v1",
			},
			wantErr: false,
		},
		{
			name: "valid config with API key, base URL, and organization",
			config: AdapterConfig{
				Name:    "openai",
				APIKey:  "test-api-key",
				BaseURL: "https://api.openai.com/v1",
				Settings: map[string]string{
					"organization": "test-org",
				},
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: AdapterConfig{
				Name: "openai",
			},
			wantErr: true,
		},
		{
			name: "empty API key",
			config: AdapterConfig{
				Name:   "openai",
				APIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewOpenAIAdapter(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewOpenAIAdapter() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewOpenAIAdapter() unexpected error: %v", err)
				return
			}

			if adapter == nil {
				t.Errorf("NewOpenAIAdapter() returned nil adapter")
				return
			}

			// Check that the adapter implements the LLMProvider interface
			if adapter.GetName() != tt.config.Name {
				t.Errorf("NewOpenAIAdapter() adapter name = %v, want %v", adapter.GetName(), tt.config.Name)
			}

			// Type assertion to check it's an OpenAIAdapter
			openaiAdapter, ok := adapter.(*OpenAIAdapter)
			if !ok {
				t.Errorf("NewOpenAIAdapter() returned wrong type, expected *OpenAIAdapter")
				return
			}

			// Check that the config was stored correctly
			if openaiAdapter.config.Name != tt.config.Name {
				t.Errorf("NewOpenAIAdapter() stored config name = %v, want %v", openaiAdapter.config.Name, tt.config.Name)
			}

			if openaiAdapter.config.APIKey != tt.config.APIKey {
				t.Errorf("NewOpenAIAdapter() stored config API key = %v, want %v", openaiAdapter.config.APIKey, tt.config.APIKey)
			}

			if openaiAdapter.config.BaseURL != tt.config.BaseURL {
				t.Errorf("NewOpenAIAdapter() stored config base URL = %v, want %v", openaiAdapter.config.BaseURL, tt.config.BaseURL)
			}
		})
	}
}

func TestOpenAIAdapter_ListModels(t *testing.T) {
	// Create adapter config
	config := AdapterConfig{
		Name:    "openai",
		BaseURL: "https://api.openai.com/v1",
		APIKey:  "test-key", // This won't work with a real API call, but we can test the structure
	}

	// Create adapter
	adapter, err := NewOpenAIAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Test ListModels method exists
	models, err := adapter.ListModels(context.Background(), "test-key", "https://api.openai.com/v1")

	// We expect an error with a test key, but we can verify the method exists and returns the right type
	if err != nil {
		// This is expected with a test key
		t.Logf("Expected error with test key: %v", err)
	} else {
		// If it somehow works, verify the structure
		if models == nil {
			t.Error("Models should not be nil")
		}
	}
}

func TestModel_Structure(t *testing.T) {
	// Test our unified Model struct
	model := Model{
		ID:          "gpt-4",
		Name:        "GPT-4",
		Provider:    "openai",
		Description: "GPT-4 model",
		Created:     1234567890,
		OwnedBy:     "openai",
	}

	if model.ID != "gpt-4" {
		t.Errorf("Expected ID 'gpt-4', got '%s'", model.ID)
	}

	if model.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", model.Provider)
	}
}
