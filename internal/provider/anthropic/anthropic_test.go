package anthropic

import (
	"testing"

	"github.com/kazz187/goline/internal/provider"
)

func TestProviderRegistration(t *testing.T) {
	// Create a provider with valid parameters
	p, err := NewProvider("test-api-key", "", "")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Check provider name
	if p.Name() != "anthropic" {
		t.Errorf("Expected provider name to be 'anthropic', got '%s'", p.Name())
	}

	// Check default model
	model := p.GetModel()
	if model.Name != string(DefaultModelID) {
		t.Errorf("Expected default model to be '%s', got '%s'", DefaultModelID, model.Name)
	}
}

func TestModelSelection(t *testing.T) {
	testCases := []struct {
		name      string
		modelName string
		expected  ModelID
	}{
		{"Empty model name", "", DefaultModelID},
		{"Valid model name", string(Claude3Opus), Claude3Opus},
		{"Invalid model name", "invalid-model", DefaultModelID},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modelID := getModelID(tc.modelName)
			if modelID != tc.expected {
				t.Errorf("Expected model ID to be '%s', got '%s'", tc.expected, modelID)
			}
		})
	}
}

func TestProviderFactory(t *testing.T) {
	// Get the factory from the registry
	factory, ok := provider.GetFactory("anthropic")
	if !ok {
		t.Fatal("Anthropic provider factory not registered")
	}

	// Create a provider using the factory
	p, err := factory("test-api-key", "", "")
	if err != nil {
		t.Fatalf("Failed to create provider using factory: %v", err)
	}

	// Check provider name
	if p.Name() != "anthropic" {
		t.Errorf("Expected provider name to be 'anthropic', got '%s'", p.Name())
	}
}
