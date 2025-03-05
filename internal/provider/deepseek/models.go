package deepseek

import (
	"github.com/kazz187/goline/internal/provider"
)

// ModelID represents a DeepSeek model ID
type ModelID string

// DeepSeek model IDs
const (
	DeepSeekChat     ModelID = "deepseek-chat"
	DeepSeekReasoner ModelID = "deepseek-reasoner"
)

// DefaultModelID is the default DeepSeek model ID
const DefaultModelID = DeepSeekChat

// Models is a map of DeepSeek model IDs to model information
var Models = map[ModelID]provider.ModelInfo{
	DeepSeekChat: {
		Name:                string(DeepSeekChat),
		MaxTokens:           64000,
		InputCostPer1K:      0.00027,
		OutputCostPer1K:     0.0011,
		CacheWriteCostPer1K: 0.00027,
		CacheReadCostPer1K:  0.00007,
	},
	DeepSeekReasoner: {
		Name:                string(DeepSeekReasoner),
		MaxTokens:           64000,
		InputCostPer1K:      0.00055,
		OutputCostPer1K:     0.00219,
		CacheWriteCostPer1K: 0.00055,
		CacheReadCostPer1K:  0.00014,
	},
}
