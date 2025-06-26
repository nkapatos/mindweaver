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
	config AdapterConfig
}

// NewOpenAIAdapter creates a new OpenAI-compatible adapter
func NewOpenAIAdapter(config AdapterConfig) (LLMProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required for OpenAI adapter")
	}

	// Create client options
	var opts []option.RequestOption

	// Add API key
	opts = append(opts, option.WithAPIKey(config.APIKey))

	// Set custom base URL if provided
	if config.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(config.BaseURL))
	}

	// Set organization if provided
	if org, exists := config.Settings["organization"]; exists && org != "" {
		opts = append(opts, option.WithOrganization(org))
	}

	// Create client
	client := openai.NewClient(opts...)

	return &OpenAIAdapter{
		BaseAdapter: BaseAdapter{
			Name: config.Name,
		},
		client: client,
		config: config,
	}, nil
}

func (a *OpenAIAdapter) Generate(ctx context.Context, prompt string, options GenerateOptions) (*GenerateResponse, error) {
	// Set default model if not provided
	model := options.Model
	if model == "" {
		model = "gpt-3.5-turbo-instruct" // Default model for completions
	}

	// Create completion parameters
	params := openai.CompletionNewParams{
		Prompt: openai.CompletionNewParamsPromptUnion{
			OfString: openai.String(prompt),
		},
		Model:       openai.CompletionNewParamsModelGPT3_5TurboInstruct, // TODO: update to make it dynamic
		Temperature: openai.Float(options.Temperature),
		MaxTokens:   openai.Int(options.MaxTokens),
	}

	// Add optional parameters if they have non-zero values
	if options.Temperature == 0 {
		params.Temperature = openai.Float(0.7) // Use default temperature
	}
	if options.MaxTokens == 0 {
		params.MaxTokens = openai.Int(12) // Use default max tokens
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
