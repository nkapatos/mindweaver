package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// LLMConfiguration represents the universal configuration for LLM services
// This struct covers the common parameters that work across all major LLM providers
type LLMConfiguration struct {
	// Core Model Settings
	Model string `json:"model" validate:"required"` // e.g., "gpt-4", "claude-3-opus", "llama-3.1-8b"

	// Generation Parameters (universal across providers)
	Temperature      *float64 `json:"temperature,omitempty"`       // 0.0 to 2.0, typically 0.7
	MaxTokens        *int     `json:"max_tokens,omitempty"`        // Maximum tokens to generate
	TopP             *float64 `json:"top_p,omitempty"`             // Nucleus sampling, 0.0 to 1.0
	TopK             *int     `json:"top_k,omitempty"`             // Top-k sampling
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"` // -2.0 to 2.0
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`  // -2.0 to 2.0

	// Response Format (for structured outputs)
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// Safety & Moderation
	SafeMode *bool `json:"safe_mode,omitempty"` // Enable safety filters

	// Provider-specific extensions (stored as raw JSON)
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// ResponseFormat defines how the LLM should format its response
type ResponseFormat struct {
	Type string `json:"type" validate:"oneof=text json"` // "text" or "json"
}

// DefaultConfiguration returns a sensible default configuration
func DefaultConfiguration(model string) *LLMConfiguration {
	temp := 0.7
	maxTokens := 1000
	topP := 1.0

	return &LLMConfiguration{
		Model:          model,
		Temperature:    &temp,
		MaxTokens:      &maxTokens,
		TopP:           &topP,
		ResponseFormat: &ResponseFormat{Type: "text"},
		SafeMode:       nil, // Let provider decide
		Extensions:     make(map[string]interface{}),
	}
}

// Validate ensures the configuration is valid
func (c *LLMConfiguration) Validate() error {
	var errors []string

	// Required fields
	if c.Model == "" {
		errors = append(errors, "model is required")
	}

	// Validate temperature range
	if c.Temperature != nil {
		if *c.Temperature < 0.0 || *c.Temperature > 2.0 {
			errors = append(errors, "temperature must be between 0.0 and 2.0")
		}
	}

	// Validate max tokens
	if c.MaxTokens != nil {
		if *c.MaxTokens <= 0 {
			errors = append(errors, "max_tokens must be positive")
		}
		if *c.MaxTokens > 100000 {
			errors = append(errors, "max_tokens cannot exceed 100,000")
		}
	}

	// Validate top_p range
	if c.TopP != nil {
		if *c.TopP < 0.0 || *c.TopP > 1.0 {
			errors = append(errors, "top_p must be between 0.0 and 1.0")
		}
	}

	// Validate top_k
	if c.TopK != nil {
		if *c.TopK <= 0 {
			errors = append(errors, "top_k must be positive")
		}
	}

	// Validate frequency penalty
	if c.FrequencyPenalty != nil {
		if *c.FrequencyPenalty < -2.0 || *c.FrequencyPenalty > 2.0 {
			errors = append(errors, "frequency_penalty must be between -2.0 and 2.0")
		}
	}

	// Validate presence penalty
	if c.PresencePenalty != nil {
		if *c.PresencePenalty < -2.0 || *c.PresencePenalty > 2.0 {
			errors = append(errors, "presence_penalty must be between -2.0 and 2.0")
		}
	}

	// Validate response format
	if c.ResponseFormat != nil {
		if c.ResponseFormat.Type != "text" && c.ResponseFormat.Type != "json" {
			errors = append(errors, "response_format.type must be 'text' or 'json'")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ToJSON converts the configuration to JSON string for database storage
func (c *LLMConfiguration) ToJSON() (string, error) {
	if err := c.Validate(); err != nil {
		return "", fmt.Errorf("cannot serialize invalid configuration: %w", err)
	}

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal configuration to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// FromJSON creates a configuration from JSON string
func FromJSON(jsonStr string) (*LLMConfiguration, error) {
	var config LLMConfiguration

	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to configuration: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration from JSON: %w", err)
	}

	return &config, nil
}

// GetProviderSpecificConfig extracts provider-specific configuration
func (c *LLMConfiguration) GetProviderSpecificConfig(provider string) map[string]interface{} {
	if c.Extensions == nil {
		return nil
	}

	if providerConfig, exists := c.Extensions[provider]; exists {
		if configMap, ok := providerConfig.(map[string]interface{}); ok {
			return configMap
		}
	}

	return nil
}

// SetProviderSpecificConfig sets provider-specific configuration
func (c *LLMConfiguration) SetProviderSpecificConfig(provider string, config map[string]interface{}) {
	if c.Extensions == nil {
		c.Extensions = make(map[string]interface{})
	}
	c.Extensions[provider] = config
}

// LogConfiguration logs the configuration for debugging (without sensitive data)
func (c *LLMConfiguration) LogConfiguration(logger *slog.Logger) {
	logger.Info("LLM Configuration",
		"model", c.Model,
		"temperature", c.Temperature,
		"max_tokens", c.MaxTokens,
		"top_p", c.TopP,
		"top_k", c.TopK,
		"frequency_penalty", c.FrequencyPenalty,
		"presence_penalty", c.PresencePenalty,
		"response_format", c.ResponseFormat,
		"safe_mode", c.SafeMode,
		"has_extensions", len(c.Extensions) > 0,
	)
}
