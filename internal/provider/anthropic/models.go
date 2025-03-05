package anthropic

import (
	"github.com/kazz187/goline/internal/provider"
)

// ModelID represents an Anthropic model ID
type ModelID string

// Anthropic model IDs
const (
	Claude3Opus    ModelID = "claude-3-opus-20240229"
	Claude3Haiku   ModelID = "claude-3-haiku-20240307"
	Claude35Sonnet ModelID = "claude-3-5-sonnet-20241022"
	Claude35Haiku  ModelID = "claude-3-5-haiku-20241022"
	Claude37Sonnet ModelID = "claude-3-7-sonnet-20250219"
)

// DefaultModelID is the default Anthropic model ID
const DefaultModelID = Claude37Sonnet

// Models is a map of Anthropic model IDs to model information
var Models = map[ModelID]provider.ModelInfo{
	Claude3Opus: {
		Name:                string(Claude3Opus),
		MaxTokens:           200000,
		InputCostPer1K:      0.015,
		OutputCostPer1K:     0.075,
		CacheWriteCostPer1K: 0.01875,
		CacheReadCostPer1K:  0.0015,
	},
	Claude3Haiku: {
		Name:                string(Claude3Haiku),
		MaxTokens:           200000,
		InputCostPer1K:      0.00025,
		OutputCostPer1K:     0.00125,
		CacheWriteCostPer1K: 0.0003,
		CacheReadCostPer1K:  0.00003,
	},
	Claude35Sonnet: {
		Name:                string(Claude35Sonnet),
		MaxTokens:           200000,
		InputCostPer1K:      0.003,
		OutputCostPer1K:     0.015,
		CacheWriteCostPer1K: 0.00375,
		CacheReadCostPer1K:  0.0003,
	},
	Claude35Haiku: {
		Name:                string(Claude35Haiku),
		MaxTokens:           200000,
		InputCostPer1K:      0.0008,
		OutputCostPer1K:     0.004,
		CacheWriteCostPer1K: 0.001,
		CacheReadCostPer1K:  0.00008,
	},
	Claude37Sonnet: {
		Name:                string(Claude37Sonnet),
		MaxTokens:           200000,
		InputCostPer1K:      0.003,
		OutputCostPer1K:     0.015,
		CacheWriteCostPer1K: 0.00375,
		CacheReadCostPer1K:  0.0003,
	},
}
