# Anthropic Provider for Goline

This package implements the Anthropic provider for the Goline project, allowing interaction with Anthropic's Claude AI models through their API.

## Features

- Support for multiple Claude models:
  - `claude-3-opus-20240229`: Most powerful Claude model
  - `claude-3-sonnet-20240229`: Balanced performance and cost
  - `claude-3-haiku-20240307`: Fast and cost-effective
  - `claude-3-5-sonnet-20241022`: Improved Claude 3.5 Sonnet model
  - `claude-3-5-haiku-20241022`: Improved Claude 3.5 Haiku model
  - `claude-3-7-sonnet-20250219`: Latest Claude 3.7 model with thinking capabilities

- Streaming responses for real-time interaction
- Support for Claude's thinking/reasoning capabilities (for Claude 3.7 models)
- Token usage tracking and cost estimation
- Prompt caching support for all Claude 3 models
- Support for custom API endpoints

## Usage

### Configuration

To use the Anthropic provider, you need to configure it with your API key. You can do this using the Goline configuration commands:

```bash
# Set up the Anthropic provider with your API key
goline config provider set anthropic --api-key YOUR_API_KEY

# Set Anthropic as the default provider
goline config default-provider set anthropic

# Set a specific model for Anthropic
goline config provider set anthropic --model claude-3-opus-20240229
```

### In Code

To use the Anthropic provider in your Go code:

```go
import (
    "github.com/kazz187/goline/internal/provider"
    _ "github.com/kazz187/goline/internal/provider/anthropic" // Import for side effects (init registration)
)

// Create an Anthropic provider
p, err := provider.Create("anthropic", apiKey, endpoint, modelName)
if err != nil {
    // Handle error
}

// Send a message
eventCh, err := p.CreateMessage(ctx, systemPrompt, messages)
if err != nil {
    // Handle error
}

// Process the response stream
for event := range eventCh {
    switch event.Type {
    case "text":
        // Handle text content
        fmt.Print(event.Text)
    case "reasoning":
        // Handle reasoning content (for Claude 3.7 models)
        fmt.Printf("\n[Reasoning: %s]\n", event.Reasoning)
    case "usage":
        // Handle usage information
        fmt.Printf("Tokens: %d input, %d output\n", 
            event.Usage.InputTokens, event.Usage.OutputTokens)
    case "error":
        // Handle errors
        fmt.Printf("Error: %s\n", event.Text)
    }
}
```

See the [example](./example/main.go) for a complete working example.

## Implementation Notes

- The implementation uses a direct HTTP client to interact with the Anthropic API, following the [Anthropic API documentation](https://docs.anthropic.com/claude/reference/getting-started-with-the-api).
- Claude 3.7 models support "thinking" capabilities, which allow the model to show its reasoning process. This is enabled by default for Claude 3.7 models.
- All Claude 3 models support prompt caching, which can reduce token usage and costs for repeated prompts.
- The provider handles SSE (Server-Sent Events) streaming for real-time responses.

## API Documentation

For more information about the Anthropic API, see the [official documentation](https://docs.anthropic.com/claude/reference/getting-started-with-the-api).
