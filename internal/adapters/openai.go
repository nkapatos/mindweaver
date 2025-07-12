package adapters

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIAdapter implements LLMProvider for OpenAI and OpenAI-compatible providers
type OpenAIAdapter struct {
	BaseAdapter
	client openai.Client
}

// NewOpenAIAdapter creates a new OpenAI-compatible adapter
func NewOpenAIAdapter(apiKey, baseURL string) (LLMProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI adapter")
	}

	// Create client options
	var opts []option.RequestOption

	// Add API key
	opts = append(opts, option.WithAPIKey(apiKey))

	// Set custom base URL if provided
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	// Create client
	client := openai.NewClient(opts...)

	return &OpenAIAdapter{
		BaseAdapter: BaseAdapter{
			Name:    "openai",
			APIKey:  apiKey,
			BaseURL: baseURL,
		},
		client: client,
	}, nil
}

// GetModels fetches available models from OpenAI API
func (a *OpenAIAdapter) GetModels(ctx context.Context) ([]Model, error) {
	// Fetch models from OpenAI API
	resp, err := a.client.Models.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models from OpenAI: %w", err)
	}

	// Convert OpenAI models to our unified Model struct
	var models []Model
	for _, openaiModel := range resp.Data {
		model := Model{
			ID:       openaiModel.ID,
			Name:     openaiModel.ID, // Use ID as name for consistency
			Provider: a.GetName(),
			Created:  openaiModel.Created,
			OwnedBy:  openaiModel.OwnedBy,
		}
		models = append(models, model)
	}

	return models, nil
}

// Generate generates a response using OpenAI's completion API
func (a *OpenAIAdapter) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	// Model is required
	if options.Model == "" {
		return nil, fmt.Errorf("model is required for generation")
	}

	// Create completion parameters
	params := openai.CompletionNewParams{
		Prompt: openai.CompletionNewParamsPromptUnion{
			OfString: openai.String(prompt),
		},
		Model:       openai.CompletionNewParamsModel(options.Model),
		Temperature: openai.Float(options.Temperature),
		MaxTokens:   openai.Int(options.MaxTokens),
	}

	// Call the OpenAI API
	res, err := a.client.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to generate completion: %w", err)
	}

	// Extract the generated text from the first choice
	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	content := res.Choices[0].Text

	return &GenerateResponse{
		Content: content,
	}, nil
}

// GenerateResponse is a convenience method that combines prompt and system prompt
func (a *OpenAIAdapter) GenerateResponse(ctx context.Context, prompt, systemPrompt string, options map[string]interface{}) (string, error) {
	// Combine system prompt and user prompt
	fullPrompt := prompt
	if systemPrompt != "" {
		fullPrompt = systemPrompt + "\n\n" + prompt
	}

	// Extract required options
	model, ok := options["model"].(string)
	if !ok || model == "" {
		return "", fmt.Errorf("model is required in options")
	}

	temperature := 0.7 // default
	if temp, exists := options["temperature"]; exists {
		if tempFloat, ok := temp.(float64); ok {
			temperature = tempFloat
		}
	}

	maxTokens := int64(1000) // default
	if max, exists := options["max_tokens"]; exists {
		if maxInt, ok := max.(int); ok {
			maxTokens = int64(maxInt)
		}
	}

	// Generate response
	response, err := a.Generate(ctx, fullPrompt, GenerateOptions{
		Model:       model,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return "", err
	}

	return response.Content, nil
}
