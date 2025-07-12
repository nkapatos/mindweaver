package services

import (
	"fmt"
	"strings"
)

// OpenAIValidator implements ProviderValidator for OpenAI and OpenAI-compatible services
type OpenAIValidator struct {
	SupportedModels []string
}

// NewOpenAIValidator creates a new OpenAI validator
func NewOpenAIValidator() *OpenAIValidator {
	return &OpenAIValidator{
		SupportedModels: []string{
			"gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini",
			"gpt-3.5-turbo", "gpt-3.5-turbo-instruct",
			"gpt-4-1106-preview", "gpt-4-0125-preview",
		},
	}
}

// ValidateConfig validates OpenAI-specific configuration
func (v *OpenAIValidator) ValidateConfig(config *LLMConfiguration) ValidationResult {
	result := ValidationResult{Valid: true}

	// Check if model is supported
	if !v.isModelSupported(config.Model) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("model '%s' is not supported by OpenAI", config.Model))
	}

	// OpenAI-specific validations
	if config.MaxTokens != nil && *config.MaxTokens > 4096 {
		result.Warnings = append(result.Warnings, "max_tokens exceeds 4096, which may cause issues with some OpenAI models")
	}

	// Check for OpenAI-specific extensions
	if openaiConfig := config.GetProviderSpecificConfig("openai"); openaiConfig != nil {
		if org, exists := openaiConfig["organization"]; exists {
			if orgStr, ok := org.(string); !ok || orgStr == "" {
				result.Errors = append(result.Errors, "openai.organization must be a non-empty string")
				result.Valid = false
			}
		}
	}

	return result
}

// GetSupportedModels returns the list of supported models
func (v *OpenAIValidator) GetSupportedModels() []string {
	return v.SupportedModels
}

// GetDefaultConfig returns a default configuration for the given model
func (v *OpenAIValidator) GetDefaultConfig(model string) *LLMConfiguration {
	config := DefaultConfiguration(model)
	config.Provider = "openai"

	// Set OpenAI-specific defaults
	if strings.Contains(model, "gpt-4") {
		temp := 0.7
		maxTokens := 4000
		config.Temperature = &temp
		config.MaxTokens = &maxTokens
	} else if strings.Contains(model, "gpt-3.5") {
		temp := 0.7
		maxTokens := 2000
		config.Temperature = &temp
		config.MaxTokens = &maxTokens
	}

	return config
}

// isModelSupported checks if the model is supported by OpenAI
func (v *OpenAIValidator) isModelSupported(model string) bool {
	for _, supported := range v.SupportedModels {
		if supported == model {
			return true
		}
	}
	return false
}

// AnthropicValidator implements ProviderValidator for Anthropic services
type AnthropicValidator struct {
	SupportedModels []string
}

// NewAnthropicValidator creates a new Anthropic validator
func NewAnthropicValidator() *AnthropicValidator {
	return &AnthropicValidator{
		SupportedModels: []string{
			"claude-3-opus", "claude-3-sonnet", "claude-3-haiku",
			"claude-3-5-sonnet", "claude-3-5-haiku",
		},
	}
}

// ValidateConfig validates Anthropic-specific configuration
func (v *AnthropicValidator) ValidateConfig(config *LLMConfiguration) ValidationResult {
	result := ValidationResult{Valid: true}

	// Check if model is supported
	if !v.isModelSupported(config.Model) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("model '%s' is not supported by Anthropic", config.Model))
	}

	// Anthropic-specific validations
	if config.MaxTokens != nil && *config.MaxTokens > 4096 {
		result.Warnings = append(result.Warnings, "max_tokens exceeds 4096, which may cause issues with some Anthropic models")
	}

	// Anthropic doesn't support top_k
	if config.TopK != nil {
		result.Warnings = append(result.Warnings, "top_k is not supported by Anthropic models")
	}

	return result
}

// GetSupportedModels returns the list of supported models
func (v *AnthropicValidator) GetSupportedModels() []string {
	return v.SupportedModels
}

// GetDefaultConfig returns a default configuration for the given model
func (v *AnthropicValidator) GetDefaultConfig(model string) *LLMConfiguration {
	config := DefaultConfiguration(model)
	config.Provider = "anthropic"

	// Set Anthropic-specific defaults
	if strings.Contains(model, "opus") {
		temp := 0.7
		maxTokens := 4096
		config.Temperature = &temp
		config.MaxTokens = &maxTokens
	} else if strings.Contains(model, "sonnet") {
		temp := 0.7
		maxTokens := 4096
		config.Temperature = &temp
		config.MaxTokens = &maxTokens
	} else if strings.Contains(model, "haiku") {
		temp := 0.7
		maxTokens := 4096
		config.Temperature = &temp
		config.MaxTokens = &maxTokens
	}

	return config
}

// isModelSupported checks if the model is supported by Anthropic
func (v *AnthropicValidator) isModelSupported(model string) bool {
	for _, supported := range v.SupportedModels {
		if supported == model {
			return true
		}
	}
	return false
}

// OllamaValidator implements ProviderValidator for Ollama services
type OllamaValidator struct {
	SupportedModels []string
}

// NewOllamaValidator creates a new Ollama validator
func NewOllamaValidator() *OllamaValidator {
	return &OllamaValidator{
		SupportedModels: []string{
			"llama2", "llama2:7b", "llama2:13b", "llama2:70b",
			"llama3", "llama3:8b", "llama3:70b",
			"mistral", "mistral:7b", "mistral:instruct",
			"codellama", "codellama:7b", "codellama:13b", "codellama:34b",
		},
	}
}

// ValidateConfig validates Ollama-specific configuration
func (v *OllamaValidator) ValidateConfig(config *LLMConfiguration) ValidationResult {
	result := ValidationResult{Valid: true}

	// Check if model is supported (Ollama is more flexible, so we'll be lenient)
	if !v.isModelSupported(config.Model) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("model '%s' may not be available in your Ollama installation", config.Model))
	}

	// Ollama-specific validations
	if config.MaxTokens != nil && *config.MaxTokens > 8192 {
		result.Warnings = append(result.Warnings, "max_tokens exceeds 8192, which may cause issues with some Ollama models")
	}

	return result
}

// GetSupportedModels returns the list of supported models
func (v *OllamaValidator) GetSupportedModels() []string {
	return v.SupportedModels
}

// GetDefaultConfig returns a default configuration for the given model
func (v *OllamaValidator) GetDefaultConfig(model string) *LLMConfiguration {
	config := DefaultConfiguration(model)
	config.Provider = "ollama"

	// Set Ollama-specific defaults
	temp := 0.7
	maxTokens := 2048
	config.Temperature = &temp
	config.MaxTokens = &maxTokens

	return config
}

// isModelSupported checks if the model is supported by Ollama
func (v *OllamaValidator) isModelSupported(model string) bool {
	for _, supported := range v.SupportedModels {
		if supported == model {
			return true
		}
	}
	return false
}

// ValidatorRegistry manages all available validators
type ValidatorRegistry struct {
	validators map[string]ProviderValidator
}

// NewValidatorRegistry creates a new validator registry
func NewValidatorRegistry() *ValidatorRegistry {
	registry := &ValidatorRegistry{
		validators: make(map[string]ProviderValidator),
	}

	// Register default validators
	registry.Register("openai", NewOpenAIValidator())
	registry.Register("anthropic", NewAnthropicValidator())
	registry.Register("ollama", NewOllamaValidator())

	return registry
}

// Register adds a validator for a specific provider
func (r *ValidatorRegistry) Register(provider string, validator ProviderValidator) {
	r.validators[provider] = validator
}

// GetValidator returns the validator for a specific provider
func (r *ValidatorRegistry) GetValidator(provider string) (ProviderValidator, bool) {
	validator, exists := r.validators[provider]
	return validator, exists
}

// ValidateConfig validates a configuration with the appropriate provider validator
func (r *ValidatorRegistry) ValidateConfig(config *LLMConfiguration) ValidationResult {
	if config.Provider == "" {
		return ValidationResult{
			Valid:  false,
			Errors: []string{"provider is required for validation"},
		}
	}

	validator, exists := r.GetValidator(config.Provider)
	if !exists {
		return ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("no validator found for provider '%s'", config.Provider)},
		}
	}

	return config.ValidateWithProvider(validator)
}

// GetSupportedModels returns all supported models for a provider
func (r *ValidatorRegistry) GetSupportedModels(provider string) []string {
	validator, exists := r.GetValidator(provider)
	if !exists {
		return []string{}
	}
	return validator.GetSupportedModels()
}

// GetDefaultConfig returns a default configuration for a provider and model
func (r *ValidatorRegistry) GetDefaultConfig(provider, model string) *LLMConfiguration {
	validator, exists := r.GetValidator(provider)
	if !exists {
		// Return basic default if no validator found
		config := DefaultConfiguration(model)
		config.Provider = provider
		return config
	}
	return validator.GetDefaultConfig(model)
}
