package adapters

import (
	"context"
	"testing"
)

func TestNewOpenAIAdapter(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid config with API key",
			apiKey:  "test-api-key",
			wantErr: false,
		},
		{
			name:    "valid config with API key and base URL",
			apiKey:  "test-api-key",
			baseURL: "https://api.openai.com/v1",
			wantErr: false,
		},
		{
			name:    "missing API key",
			apiKey:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewOpenAIAdapter(tt.apiKey, tt.baseURL)

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
			if adapter.GetName() != "openai" {
				t.Errorf("NewOpenAIAdapter() adapter name = %v, want %v", adapter.GetName(), "openai")
			}

			// Type assertion to check it's an OpenAIAdapter
			_, ok := adapter.(*OpenAIAdapter)
			if !ok {
				t.Errorf("NewOpenAIAdapter() returned wrong type, expected *OpenAIAdapter")
				return
			}
		})
	}
}

func TestOpenAIAdapter_GetModels(t *testing.T) {
	// Create adapter
	adapter, err := NewOpenAIAdapter("test-key", "https://api.openai.com/v1")
	if err != nil {
		t.Fatalf("Failed to create OpenAI adapter: %v", err)
	}

	// Test GetModels method exists
	models, err := adapter.GetModels(context.Background())

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
