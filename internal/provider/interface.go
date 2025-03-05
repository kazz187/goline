package provider

import (
	"context"
	"io"
)

// Message represents a message in a conversation
type Message struct {
	// Role of the message sender (e.g., "user", "assistant", "system")
	Role string
	// Content of the message
	Content string
	// Optional reasoning content for models that support it
	ReasoningContent string
}

// Usage represents token usage information
type Usage struct {
	// Number of input tokens
	InputTokens int
	// Number of output tokens
	OutputTokens int
	// Number of tokens read from cache (if supported)
	CacheReadTokens int
	// Number of tokens written to cache (if supported)
	CacheWriteTokens int
	// Estimated cost of the API call
	TotalCost float64
}

// StreamEvent represents an event in the response stream
type StreamEvent struct {
	// Type of event ("text", "reasoning", "usage")
	Type string
	// Text content (for "text" events)
	Text string
	// Reasoning content (for "reasoning" events)
	Reasoning string
	// Usage information (for "usage" events)
	Usage *Usage
}

// ModelInfo represents information about a model
type ModelInfo struct {
	// Name of the model
	Name string
	// Maximum number of tokens the model can process
	MaxTokens int
	// Cost per 1K input tokens
	InputCostPer1K float64
	// Cost per 1K output tokens
	OutputCostPer1K float64
	// Cost per 1K cache write tokens (if supported)
	CacheWriteCostPer1K float64
	// Cost per 1K cache read tokens (if supported)
	CacheReadCostPer1K float64
}

// Provider defines the interface for AI providers
type Provider interface {
	// CreateMessage sends a message to the AI provider and returns a stream of events
	CreateMessage(ctx context.Context, systemPrompt string, messages []Message) (chan StreamEvent, error)

	// GetModel returns information about the current model
	GetModel() ModelInfo

	// Name returns the name of the provider
	Name() string
}

// Factory creates a provider instance from configuration
type Factory func(apiKey, endpoint, modelName string) (Provider, error)

// registry of provider factories
var providerFactories = make(map[string]Factory)

// Register registers a provider factory
func Register(name string, factory Factory) {
	providerFactories[name] = factory
}

// Create creates a provider instance
func Create(name, apiKey, endpoint, modelName string) (Provider, error) {
	factory, ok := providerFactories[name]
	if !ok {
		return nil, ErrProviderNotFound
	}
	return factory(apiKey, endpoint, modelName)
}

// GetFactory returns a provider factory by name
func GetFactory(name string) (Factory, bool) {
	factory, ok := providerFactories[name]
	return factory, ok
}

// ErrProviderNotFound is returned when a provider is not found
var ErrProviderNotFound = io.EOF
