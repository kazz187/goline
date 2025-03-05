# DeepSeek Provider for Goline

This package implements the DeepSeek provider for the Goline project, allowing interaction with DeepSeek's AI models through their API.

## Features

- Support for multiple DeepSeek models:
  - `deepseek-chat`: General-purpose chat model
  - `deepseek-coder`: Code-specialized model
  - `deepseek-reasoner`: Model with reasoning capabilities
  - `deepseek-lite`: Lighter, faster model
  - `deepseek-lite-reason`: Lighter model with reasoning capabilities

- Streaming responses for real-time interaction
- Token usage tracking and cost estimation
- Support for custom API endpoints

## Usage

### Configuration

To use the DeepSeek provider, you need to configure it with your API key. You can do this using the Goline configuration commands:

```bash
# Set up the DeepSeek provider with your API key
goline config provider set deepseek --api-key YOUR_API_KEY

# Set DeepSeek as the default provider
goline config default-provider set deepseek

# Set a specific model for DeepSeek
goline config provider set deepseek --model deepseek-coder
```

### In Code

To use the DeepSeek provider in your Go code:

```go
import (
    "github.com/kazz187/goline/internal/provider"
    _ "github.com/kazz187/goline/internal/provider/deepseek" // Import for side effects (init registration)
)

// Create a DeepSeek provider
p, err := provider.Create("deepseek", apiKey, endpoint, modelName)
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
        // Handle reasoning content (for reasoner models)
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

- The DeepSeek API is compatible with the OpenAI API format, so we use the `github.com/sashabaranov/go-openai` client library.
- DeepSeek models support context caching, but the current implementation doesn't fully utilize this feature due to limitations in the OpenAI client library.
- For reasoner models, the reasoning content is not currently accessible through the OpenAI client library. A custom implementation would be needed to fully support this feature.

## API Documentation

For more information about the DeepSeek API, see the [official documentation](https://api-docs.deepseek.com/).
