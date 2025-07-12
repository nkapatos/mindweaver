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

	// Metadata for tracking and validation (not stored in JSON)
	Provider    string `json:"-"` // Provider name (e.g., "openai", "anthropic")
	ServiceID   int64  `json:"-"` // LLM service ID
	ConfigID    int64  `json:"-"` // LLM service config ID
	ValidatedAt int64  `json:"-"` // Unix timestamp of last validation
}

// ResponseFormat defines how the LLM should format its response
type ResponseFormat struct {
	Type string `json:"type" validate:"oneof=text json"` // "text" or "json"
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ProviderValidator defines the interface for provider-specific validation
type ProviderValidator interface {
	ValidateConfig(config *LLMConfiguration) ValidationResult
	GetSupportedModels() []string
	GetDefaultConfig(model string) *LLMConfiguration
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

// ValidateWithProvider performs validation with provider-specific rules
func (c *LLMConfiguration) ValidateWithProvider(validator ProviderValidator) ValidationResult {
	// First, perform basic validation
	if err := c.Validate(); err != nil {
		return ValidationResult{
			Valid:  false,
			Errors: []string{err.Error()},
		}
	}

	// Then, perform provider-specific validation
	return validator.ValidateConfig(c)
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

// Merge merges another configuration into this one, with this config taking precedence
func (c *LLMConfiguration) Merge(other *LLMConfiguration) {
	if other == nil {
		return
	}

	// Merge basic fields (this config takes precedence)
	if c.Model == "" && other.Model != "" {
		c.Model = other.Model
	}
	if c.Temperature == nil && other.Temperature != nil {
		c.Temperature = other.Temperature
	}
	if c.MaxTokens == nil && other.MaxTokens != nil {
		c.MaxTokens = other.MaxTokens
	}
	if c.TopP == nil && other.TopP != nil {
		c.TopP = other.TopP
	}
	if c.TopK == nil && other.TopK != nil {
		c.TopK = other.TopK
	}
	if c.FrequencyPenalty == nil && other.FrequencyPenalty != nil {
		c.FrequencyPenalty = other.FrequencyPenalty
	}
	if c.PresencePenalty == nil && other.PresencePenalty != nil {
		c.PresencePenalty = other.PresencePenalty
	}
	if c.ResponseFormat == nil && other.ResponseFormat != nil {
		c.ResponseFormat = other.ResponseFormat
	}
	if c.SafeMode == nil && other.SafeMode != nil {
		c.SafeMode = other.SafeMode
	}

	// Merge extensions
	if other.Extensions != nil {
		if c.Extensions == nil {
			c.Extensions = make(map[string]interface{})
		}
		for provider, config := range other.Extensions {
			if _, exists := c.Extensions[provider]; !exists {
				c.Extensions[provider] = config
			}
		}
	}
}

// Clone creates a deep copy of the configuration
func (c *LLMConfiguration) Clone() *LLMConfiguration {
	if c == nil {
		return nil
	}

	clone := &LLMConfiguration{
		Model:       c.Model,
		Provider:    c.Provider,
		ServiceID:   c.ServiceID,
		ConfigID:    c.ConfigID,
		ValidatedAt: c.ValidatedAt,
	}

	// Clone pointers
	if c.Temperature != nil {
		temp := *c.Temperature
		clone.Temperature = &temp
	}
	if c.MaxTokens != nil {
		tokens := *c.MaxTokens
		clone.MaxTokens = &tokens
	}
	if c.TopP != nil {
		topP := *c.TopP
		clone.TopP = &topP
	}
	if c.TopK != nil {
		topK := *c.TopK
		clone.TopK = &topK
	}
	if c.FrequencyPenalty != nil {
		freq := *c.FrequencyPenalty
		clone.FrequencyPenalty = &freq
	}
	if c.PresencePenalty != nil {
		pres := *c.PresencePenalty
		clone.PresencePenalty = &pres
	}
	if c.SafeMode != nil {
		safe := *c.SafeMode
		clone.SafeMode = &safe
	}

	// Clone response format
	if c.ResponseFormat != nil {
		clone.ResponseFormat = &ResponseFormat{
			Type: c.ResponseFormat.Type,
		}
	}

	// Clone extensions
	if c.Extensions != nil {
		clone.Extensions = make(map[string]interface{})
		for k, v := range c.Extensions {
			clone.Extensions[k] = v
		}
	}

	return clone
}

// ToAdapterOptions converts the configuration to adapter-specific options
func (c *LLMConfiguration) ToAdapterOptions() map[string]interface{} {
	options := make(map[string]interface{})

	if c.Model != "" {
		options["model"] = c.Model
	}
	if c.Temperature != nil {
		options["temperature"] = *c.Temperature
	}
	if c.MaxTokens != nil {
		options["max_tokens"] = *c.MaxTokens
	}
	if c.TopP != nil {
		options["top_p"] = *c.TopP
	}
	if c.TopK != nil {
		options["top_k"] = *c.TopK
	}
	if c.FrequencyPenalty != nil {
		options["frequency_penalty"] = *c.FrequencyPenalty
	}
	if c.PresencePenalty != nil {
		options["presence_penalty"] = *c.PresencePenalty
	}
	if c.ResponseFormat != nil {
		options["response_format"] = c.ResponseFormat
	}
	if c.SafeMode != nil {
		options["safe_mode"] = *c.SafeMode
	}

	// Add provider-specific extensions
	for provider, config := range c.Extensions {
		options[provider] = config
	}

	return options
}

// LogConfiguration logs the configuration for debugging (without sensitive data)
func (c *LLMConfiguration) LogConfiguration(logger *slog.Logger) {
	logger.Info("LLM Configuration",
		"model", c.Model,
		"provider", c.Provider,
		"service_id", c.ServiceID,
		"config_id", c.ConfigID,
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
