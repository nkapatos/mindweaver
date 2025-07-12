package adapters

import (
	"context"
	"fmt"
)

// Model represents a unified model structure across all providers
type Model struct {
	ID          string `json:"id"`          // Model identifier (e.g., "gpt-4", "claude-3-opus")
	Name        string `json:"name"`        // Human-readable name
	Provider    string `json:"provider"`    // Provider name (e.g., "openai", "anthropic")
	Description string `json:"description"` // Model description
	Created     int64  `json:"created"`     // Unix timestamp when model was created
	OwnedBy     string `json:"owned_by"`    // Organization that owns the model
}

// LLMProvider defines the core interface that all LLM adapters must implement
type LLMProvider interface {
	// For one shot prompts/responses independently of the model/provider and the media type, i.e. text, image, audio, video, etc.
	Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error)

	// For multi-turn conversations, i.e. chat
	// Chat(ctx context.Context, messages []ChatMessage, options ChatOptions) (*ChatResponse, error)

	// GetName returns the provider's name
	GetName() string

	// GetModels returns available models for this provider
	GetModels(ctx context.Context) ([]Model, error)

	// GetDefaultConfig returns the default configuration for a model
	GetDefaultConfig(model string) map[string]interface{}

	// ValidateConfig validates a configuration for this provider
	ValidateConfig(config map[string]interface{}) error
}

// GenerateOptions represents common options for text generation
type GenerateOptions struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int64   `json:"max_tokens"`
}

// GenerateResponse represents the response from text generation
type GenerateResponse struct {
	Content string `json:"content"`
}

// ChatOptions represents common options for chat
type ChatOptions struct {
	Model     string `json:"model"`
	MaxTokens int64  `json:"max_tokens"`
}

// ChatResponse represents the response from chat
type ChatResponse struct {
	Messages []ChatMessage `json:"messages"`
	Usage    Usage         `json:"usage"`
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage represents the usage of the chat
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// BaseAdapter provides common functionality for all adapters
type BaseAdapter struct {
	Name    string
	APIKey  string
	BaseURL string
}

// GetName implements LLMProvider.GetName
func (b *BaseAdapter) GetName() string {
	return b.Name
}

// GetModels provides a default implementation that returns an empty list
// Individual adapters should override this method
func (b *BaseAdapter) GetModels(ctx context.Context) ([]Model, error) {
	return []Model{}, nil
}

// GetDefaultConfig provides common configuration structure
// Individual adapters should override this method
func (b *BaseAdapter) GetDefaultConfig(model string) map[string]interface{} {
	return map[string]interface{}{
		"model":       model,
		"temperature": 0.7,
		"max_tokens":  1000,
		"top_p":       1.0,
	}
}

// ValidateConfig provides basic validation for common configuration
// Individual adapters should override this method
func (b *BaseAdapter) ValidateConfig(config map[string]interface{}) error {
	// Basic validation for common fields
	if model, exists := config["model"]; !exists || model == "" {
		return fmt.Errorf("model is required")
	}

	if temp, exists := config["temperature"]; exists {
		if tempFloat, ok := temp.(float64); ok {
			if tempFloat < 0.0 || tempFloat > 2.0 {
				return fmt.Errorf("temperature must be between 0.0 and 2.0")
			}
		}
	}

	if maxTokens, exists := config["max_tokens"]; exists {
		if maxInt, ok := maxTokens.(int); ok {
			if maxInt <= 0 {
				return fmt.Errorf("max_tokens must be positive")
			}
		}
	}

	return nil
}

// NewAdapter creates a new adapter instance with the given parameters
func NewAdapter(adapterName, apiKey, baseURL string) (LLMProvider, error) {
	switch adapterName {
	case "openai":
		return NewOpenAIAdapter(apiKey, baseURL)
	case "openrouter":
		// OpenRouter uses OpenAI-compatible API
		return NewOpenAIAdapter(apiKey, baseURL)
	case "ollama":
		// Ollama can serve OpenAI-compatible endpoints
		return NewOpenAIAdapter(apiKey, baseURL)
	default:
		return nil, fmt.Errorf("unsupported adapter: %s", adapterName)
	}
}

// GenerateWithConfig generates a response using the unified configuration system
func GenerateWithConfig(ctx context.Context, adapter LLMProvider, prompt string, config map[string]interface{}) (*GenerateResponse, error) {
	// Validate configuration using adapter's validation
	if err := adapter.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Extract common options
	model := config["model"].(string)
	temperature := 0.7       // default
	maxTokens := int64(1000) // default

	if temp, exists := config["temperature"]; exists {
		if tempFloat, ok := temp.(float64); ok {
			temperature = tempFloat
		}
	}

	if max, exists := config["max_tokens"]; exists {
		if maxInt, ok := max.(int); ok {
			maxTokens = int64(maxInt)
		}
	}

	// Call the adapter's Generate method
	return adapter.Generate(ctx, prompt, GenerateOptions{
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	})
}

// GenerateResponseWithConfig generates a response using the unified configuration system
// This is a convenience method that combines prompt and system prompt
func GenerateResponseWithConfig(ctx context.Context, adapter LLMProvider, prompt, systemPrompt string, config map[string]interface{}) (string, error) {
	// Combine system prompt and user prompt
	fullPrompt := prompt
	if systemPrompt != "" {
		fullPrompt = systemPrompt + "\n\n" + prompt
	}

	// Generate response using the unified config
	response, err := GenerateWithConfig(ctx, adapter, fullPrompt, config)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}
