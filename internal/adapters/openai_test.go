package adapters

import (
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
