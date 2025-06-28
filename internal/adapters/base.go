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

	// ListModels returns available models for this provider
	ListModels(ctx context.Context, apiKey, baseURL string) ([]Model, error)
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
	Name string
}

// GetName implements LLMProvider.GetName
func (b *BaseAdapter) GetName() string {
	return b.Name
}

// ListModels provides a default implementation that returns an empty list
// Individual adapters should override this method
func (b *BaseAdapter) ListModels(ctx context.Context, apiKey, baseURL string) ([]Model, error) {
	return []Model{}, nil
}

// AdapterConfig represents common configuration for all adapters
type AdapterConfig struct {
	Name     string            `json:"name"`
	BaseURL  string            `json:"base_url"`
	APIKey   string            `json:"api_key"`
	Settings map[string]string `json:"settings"`
}

// NewAdapter creates a new adapter instance with the given config
func NewAdapter(config AdapterConfig) (LLMProvider, error) {
	switch config.Name {
	case "openai":
		return NewOpenAIAdapter(config)
	case "openrouter":
		// OpenRouter uses OpenAI-compatible API
		config.Name = "openai" // Use OpenAI adapter
		return NewOpenAIAdapter(config)
	case "ollama":
		// Ollama can serve OpenAI-compatible endpoints
		config.Name = "openai" // Use OpenAI adapter
		return NewOpenAIAdapter(config)
	default:
		return nil, fmt.Errorf("unsupported adapter: %s", config.Name)
	}
}
