package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kazz187/goline/internal/provider"
	_ "github.com/kazz187/goline/internal/provider/anthropic" // Import for side effects (init registration)
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create an Anthropic provider
	p, err := provider.Create("anthropic", apiKey, "", "")
	if err != nil {
		log.Fatalf("Failed to create Anthropic provider: %v", err)
	}

	// Print provider and model information
	fmt.Printf("Provider: %s\n", p.Name())
	model := p.GetModel()
	fmt.Printf("Model: %s (max tokens: %d)\n", model.Name, model.MaxTokens)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// System prompt and messages
	systemPrompt := "You are a helpful AI assistant."
	messages := []provider.Message{
		{
			Role:    "user",
			Content: "Hello! Can you tell me about the Go programming language?",
		},
	}

	// Send a message to the Anthropic API
	fmt.Println("\nSending message to Anthropic API...")
	eventCh, err := p.CreateMessage(ctx, systemPrompt, messages)
	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	// Process the response stream
	var fullResponse string
	for event := range eventCh {
		switch event.Type {
		case "text":
			fmt.Print(event.Text)
			fullResponse += event.Text
		case "reasoning":
			fmt.Printf("\n[Reasoning: %s]\n", event.Reasoning)
		case "usage":
			fmt.Printf("\n\nUsage: %d input tokens, %d output tokens\n",
				event.Usage.InputTokens, event.Usage.OutputTokens)
			fmt.Printf("Estimated cost: $%.6f\n", event.Usage.TotalCost)
		case "error":
			fmt.Printf("\nError: %s\n", event.Text)
		}
	}

	fmt.Println("\n\nDone!")
}
