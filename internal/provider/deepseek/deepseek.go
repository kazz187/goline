package deepseek

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/kazz187/goline/internal/provider"
	"github.com/sashabaranov/go-openai"
)

// DefaultEndpoint is the default DeepSeek API endpoint
const DefaultEndpoint = "https://api.deepseek.com/v1"

// Provider implements the provider.Provider interface for DeepSeek
type Provider struct {
	client    *openai.Client
	modelID   ModelID
	modelInfo provider.ModelInfo
}

// NewProvider creates a new DeepSeek provider
func NewProvider(apiKey, endpoint, modelName string) (provider.Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("DeepSeek API key is required")
	}

	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	// Create OpenAI client with DeepSeek endpoint
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = endpoint
	client := openai.NewClientWithConfig(config)

	// Determine model ID
	modelID := getModelID(modelName)
	modelInfo, ok := Models[modelID]
	if !ok {
		return nil, fmt.Errorf("unknown DeepSeek model: %s", modelID)
	}

	return &Provider{
		client:    client,
		modelID:   modelID,
		modelInfo: modelInfo,
	}, nil
}

// getModelID returns a ModelID from a model name string
func getModelID(modelName string) ModelID {
	if modelName == "" {
		return DefaultModelID
	}

	// Check if the model name is a valid ModelID
	for id := range Models {
		if string(id) == modelName {
			return id
		}
	}

	// If not found, return the default model ID
	return DefaultModelID
}

// CreateMessage sends a message to the DeepSeek API and returns a stream of events
func (p *Provider) CreateMessage(ctx context.Context, systemPrompt string, messages []provider.Message) (chan provider.StreamEvent, error) {
	eventCh := make(chan provider.StreamEvent)

	// Convert messages to OpenAI format
	openAIMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// Add user and assistant messages
	for _, msg := range messages {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}

		openAIMessages = append(openAIMessages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Check if we're using the reasoner model
	isReasoner := strings.Contains(string(p.modelID), "reasoner")

	// Create request
	req := openai.ChatCompletionRequest{
		Model:     string(p.modelID),
		Messages:  openAIMessages,
		Stream:    true,
		MaxTokens: p.modelInfo.MaxTokens,
	}

	// Only set temperature for non-reasoner models
	if !isReasoner {
		req.Temperature = 0
	}

	// Start streaming in a goroutine
	go func() {
		defer close(eventCh)

		stream, err := p.client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			slog.Error("Failed to create chat completion stream", "error", err)
			eventCh <- provider.StreamEvent{
				Type: "error",
				Text: fmt.Sprintf("Error: %v", err),
			}
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "stream closed") {
					// Stream closed normally
					return
				}
				slog.Error("Error receiving from stream", "error", err)
				eventCh <- provider.StreamEvent{
					Type: "error",
					Text: fmt.Sprintf("Stream error: %v", err),
				}
				return
			}

			// Process the response
			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta

				// Handle text content
				if delta.Content != "" {
					eventCh <- provider.StreamEvent{
						Type: "text",
						Text: delta.Content,
					}
				}

				// Note: The go-openai library doesn't directly expose reasoning_content
				// For DeepSeek reasoner models, we would need to extend the library
				// or use a custom implementation to access this field
			}

			// Handle usage information
			if response.Usage != nil {
				// DeepSeek reports total input AND cache reads/writes
				// See context caching: https://api-docs.deepseek.com/guides/kv_cache
				inputTokens := response.Usage.PromptTokens
				outputTokens := response.Usage.CompletionTokens

				// Note: The go-openai library doesn't directly expose cache token fields
				// We'll use only the standard fields available in the Usage struct
				var cacheReadTokens, cacheWriteTokens int
				// Cache tokens are not directly accessible with the current library

				// Calculate cost
				totalCost := calculateCost(p.modelInfo, inputTokens, outputTokens, cacheWriteTokens, cacheReadTokens)

				eventCh <- provider.StreamEvent{
					Type: "usage",
					Usage: &provider.Usage{
						InputTokens:      inputTokens,
						OutputTokens:     outputTokens,
						CacheReadTokens:  cacheReadTokens,
						CacheWriteTokens: cacheWriteTokens,
						TotalCost:        totalCost,
					},
				}
			}
		}
	}()

	return eventCh, nil
}

// calculateCost calculates the cost of an API call
func calculateCost(info provider.ModelInfo, inputTokens, outputTokens, cacheWriteTokens, cacheReadTokens int) float64 {
	inputCost := float64(inputTokens) * info.InputCostPer1K / 1000
	outputCost := float64(outputTokens) * info.OutputCostPer1K / 1000
	cacheWriteCost := float64(cacheWriteTokens) * info.CacheWriteCostPer1K / 1000
	cacheReadCost := float64(cacheReadTokens) * info.CacheReadCostPer1K / 1000

	return inputCost + outputCost + cacheWriteCost + cacheReadCost
}

// GetModel returns information about the current model
func (p *Provider) GetModel() provider.ModelInfo {
	return p.modelInfo
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return "deepseek"
}

// init registers the DeepSeek provider factory
func init() {
	provider.Register("deepseek", NewProvider)
}
