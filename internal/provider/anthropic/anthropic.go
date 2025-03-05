package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kazz187/goline/internal/provider"
)

// DefaultEndpoint is the default Anthropic API endpoint
const DefaultEndpoint = "https://api.anthropic.com/v1"

// Provider implements the provider.Provider interface for Anthropic
type Provider struct {
	client    *http.Client
	apiKey    string
	endpoint  string
	modelID   ModelID
	modelInfo provider.ModelInfo
}

// NewProvider creates a new Anthropic provider
func NewProvider(apiKey, endpoint, modelName string) (provider.Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}

	if endpoint == "" {
		endpoint = DefaultEndpoint
	}

	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Determine model ID
	modelID := getModelID(modelName)
	modelInfo, ok := Models[modelID]
	if !ok {
		return nil, fmt.Errorf("unknown Anthropic model: %s", modelID)
	}

	return &Provider{
		client:    client,
		apiKey:    apiKey,
		endpoint:  endpoint,
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

// Message represents an Anthropic message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MessageRequest represents an Anthropic message request
type MessageRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	System      string    `json:"system"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature *float64  `json:"temperature,omitempty"`
	Thinking    *Thinking `json:"thinking,omitempty"`
}

// Thinking represents the thinking configuration for Anthropic models
type Thinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}

// StreamEvent represents an event in the Anthropic stream response
type StreamEvent struct {
	Type         string        `json:"type"`
	Message      *MessageEvent `json:"message,omitempty"`
	ContentBlock *ContentBlock `json:"content_block,omitempty"`
	Delta        *DeltaEvent   `json:"delta,omitempty"`
	Usage        *UsageEvent   `json:"usage,omitempty"`
	Index        int           `json:"index,omitempty"`
}

// MessageEvent represents a message event
type MessageEvent struct {
	Usage *Usage `json:"usage,omitempty"`
}

// ContentBlock represents a content block
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// DeltaEvent represents a delta event
type DeltaEvent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// UsageEvent represents a usage event
type UsageEvent struct {
	OutputTokens int `json:"output_tokens"`
}

// CreateMessage sends a message to the Anthropic API and returns a stream of events
func (p *Provider) CreateMessage(ctx context.Context, systemPrompt string, messages []provider.Message) (chan provider.StreamEvent, error) {
	eventCh := make(chan provider.StreamEvent)

	// Convert messages to Anthropic format
	anthropicMessages := make([]Message, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "assistant"
		}

		anthropicMessages = append(anthropicMessages, Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Check if we're using a model that supports thinking
	supportsThinking := strings.Contains(string(p.modelID), "3-7")
	var thinkingBudget int
	if supportsThinking {
		thinkingBudget = 10000 // Default thinking budget
	}

	// Set temperature to 0 for deterministic responses
	var temperature *float64
	if !supportsThinking {
		temp := 0.0
		temperature = &temp
	}

	// Create message request
	req := &MessageRequest{
		Model:       string(p.modelID),
		MaxTokens:   p.modelInfo.MaxTokens,
		System:      systemPrompt,
		Messages:    anthropicMessages,
		Stream:      true,
		Temperature: temperature,
	}

	// Enable thinking for models that support it
	if supportsThinking && thinkingBudget > 0 {
		req.Thinking = &Thinking{
			Type:         "enabled",
			BudgetTokens: thinkingBudget,
		}
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", p.apiKey)
	httpReq.Header.Set("Anthropic-Version", "2023-06-01")

	// Enable prompt caching for supported models
	if isCachingSupported(p.modelID) {
		httpReq.Header.Set("Anthropic-Beta", "prompt-caching-2024-07-31")
	}

	// Start streaming in a goroutine
	go func() {
		defer close(eventCh)

		// Send request
		resp, err := p.client.Do(httpReq)
		if err != nil {
			slog.Error("Failed to send request", "error", err)
			eventCh <- provider.StreamEvent{
				Type: "error",
				Text: fmt.Sprintf("Error: %v", err),
			}
			return
		}
		defer resp.Body.Close()

		// Check for error response
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			slog.Error("API error", "status", resp.Status, "body", string(body))
			eventCh <- provider.StreamEvent{
				Type: "error",
				Text: fmt.Sprintf("API error: %s - %s", resp.Status, string(body)),
			}
			return
		}

		// Process the stream
		reader := bufio.NewReader(resp.Body)
		for {
			// Read line from stream
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "stream closed") {
					// Stream closed normally
					return
				}
				slog.Error("Error reading from stream", "error", err)
				eventCh <- provider.StreamEvent{
					Type: "error",
					Text: fmt.Sprintf("Stream error: %v", err),
				}
				return
			}

			// Skip empty lines
			line = strings.TrimSpace(line)
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			// Parse event data
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			// Parse JSON
			var event StreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				slog.Error("Failed to parse event", "error", err, "data", data)
				continue
			}

			// Process the event based on its type
			switch event.Type {
			case "message_start":
				// Handle message start event (includes usage information)
				if event.Message != nil && event.Message.Usage != nil {
					usage := event.Message.Usage
					eventCh <- provider.StreamEvent{
						Type: "usage",
						Usage: &provider.Usage{
							InputTokens:      usage.InputTokens,
							OutputTokens:     usage.OutputTokens,
							CacheReadTokens:  usage.CacheReadInputTokens,
							CacheWriteTokens: usage.CacheCreationInputTokens,
							TotalCost:        calculateCost(p.modelInfo, usage.InputTokens, usage.OutputTokens, usage.CacheCreationInputTokens, usage.CacheReadInputTokens),
						},
					}
				}

			case "message_delta":
				// Handle message delta event (includes token usage updates)
				if event.Usage != nil {
					eventCh <- provider.StreamEvent{
						Type: "usage",
						Usage: &provider.Usage{
							InputTokens:  0, // Delta only includes output tokens
							OutputTokens: event.Usage.OutputTokens,
						},
					}
				}

			case "content_block_start":
				// Handle content block start
				if event.ContentBlock != nil {
					switch event.ContentBlock.Type {
					case "thinking":
						// Handle thinking block
						if event.ContentBlock.Thinking != "" {
							eventCh <- provider.StreamEvent{
								Type:      "reasoning",
								Reasoning: event.ContentBlock.Thinking,
							}
						}
					case "text":
						// Handle text block
						if event.ContentBlock.Text != "" {
							// Add a line break between text blocks if this isn't the first one
							if event.Index > 0 {
								eventCh <- provider.StreamEvent{
									Type: "text",
									Text: "\n",
								}
							}
							eventCh <- provider.StreamEvent{
								Type: "text",
								Text: event.ContentBlock.Text,
							}
						}
					}
				}

			case "content_block_delta":
				// Handle content block delta
				if event.Delta != nil {
					switch event.Delta.Type {
					case "thinking_delta":
						// Handle thinking delta
						if event.Delta.Thinking != "" {
							eventCh <- provider.StreamEvent{
								Type:      "reasoning",
								Reasoning: event.Delta.Thinking,
							}
						}
					case "text_delta":
						// Handle text delta
						if event.Delta.Text != "" {
							eventCh <- provider.StreamEvent{
								Type: "text",
								Text: event.Delta.Text,
							}
						}
					}
				}
			}
		}
	}()

	return eventCh, nil
}

// isCachingSupported returns true if the model supports prompt caching
func isCachingSupported(modelID ModelID) bool {
	switch modelID {
	case Claude3Opus, Claude3Haiku, Claude35Sonnet, Claude35Haiku, Claude37Sonnet:
		return true
	default:
		return false
	}
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
	return "anthropic"
}

// init registers the Anthropic provider factory
func init() {
	provider.Register("anthropic", NewProvider)
}
