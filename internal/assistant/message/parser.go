package message

import (
	"fmt"
	"strings"
)

// ParseAssistantMessage parses an assistant message into content blocks
func ParseAssistantMessage(assistantMessage string) []interface{} {
	var contentBlocks []interface{}
	var currentTextContent *TextContent
	var currentTextContentStartIndex int
	var currentToolUse *ToolUse
	var currentToolUseStartIndex int
	var currentParamName ToolParamName
	var currentParamValueStartIndex int
	var accumulator string

	for i, char := range assistantMessage {
		accumulator += string(char)

		// There should not be a param without a tool use
		if currentToolUse != nil && currentParamName != "" {
			currentParamValue := accumulator[currentParamValueStartIndex:]
			paramClosingTag := fmt.Sprintf("</%s>", currentParamName)
			if strings.HasSuffix(currentParamValue, paramClosingTag) {
				// End of param value
				paramValue := currentParamValue[:len(currentParamValue)-len(paramClosingTag)]
				currentToolUse.Params[currentParamName] = strings.TrimSpace(paramValue)
				currentParamName = ""
				continue
			} else {
				// Partial param value is accumulating
				continue
			}
		}

		// No currentParamName

		if currentToolUse != nil {
			currentToolValue := accumulator[currentToolUseStartIndex:]
			toolUseClosingTag := fmt.Sprintf("</%s>", currentToolUse.Name)
			if strings.HasSuffix(currentToolValue, toolUseClosingTag) {
				// End of a tool use
				currentToolUse.Content.Partial = false
				contentBlocks = append(contentBlocks, *currentToolUse)
				currentToolUse = nil
				continue
			} else {
				// Check for parameter opening tags
				for _, paramName := range AllToolParamNames() {
					paramOpeningTag := fmt.Sprintf("<%s>", paramName)
					if strings.HasSuffix(accumulator, paramOpeningTag) {
						// Start of a new parameter
						currentParamName = paramName
						currentParamValueStartIndex = len(accumulator)
						break
					}
				}

				// Special case for write_to_file where file contents could contain the closing tag
				if currentToolUse.Name == WriteToFileToolName && strings.HasSuffix(accumulator, fmt.Sprintf("</%s>", ContentParam)) {
					toolContent := accumulator[currentToolUseStartIndex:]
					contentStartTag := fmt.Sprintf("<%s>", ContentParam)
					contentEndTag := fmt.Sprintf("</%s>", ContentParam)
					contentStartIndex := strings.Index(toolContent, contentStartTag) + len(contentStartTag)
					contentEndIndex := strings.LastIndex(toolContent, contentEndTag)
					if contentStartIndex != -1 && contentEndIndex != -1 && contentEndIndex > contentStartIndex {
						currentToolUse.Params[ContentParam] = strings.TrimSpace(toolContent[contentStartIndex:contentEndIndex])
					}
				}

				// Partial tool value is accumulating
				continue
			}
		}

		// No currentToolUse

		didStartToolUse := false
		for _, toolName := range AllToolUseNames() {
			toolUseOpeningTag := fmt.Sprintf("<%s>", toolName)
			if strings.HasSuffix(accumulator, toolUseOpeningTag) {
				// Start of a new tool use
				newToolUse := NewToolUse(toolName, true)
				currentToolUse = &newToolUse
				currentToolUseStartIndex = len(accumulator)

				// This also indicates the end of the current text content
				if currentTextContent != nil {
					currentTextContent.Content.Partial = false
					// Remove the partially accumulated tool use tag from the end of text
					content := currentTextContent.Content.Content
					tagStart := len(content) - len(toolUseOpeningTag) + 1
					if tagStart > 0 {
						currentTextContent.Content.Content = strings.TrimSpace(content[:tagStart])
					}
					contentBlocks = append(contentBlocks, *currentTextContent)
					currentTextContent = nil
				}

				didStartToolUse = true
				break
			}
		}

		if !didStartToolUse {
			// No tool use, so it must be text either at the beginning or between tools
			if currentTextContent == nil {
				currentTextContentStartIndex = i
				newTextContent := NewTextContent(accumulator[currentTextContentStartIndex:], true)
				currentTextContent = &newTextContent
			} else {
				currentTextContent.Content.Content = strings.TrimSpace(accumulator[currentTextContentStartIndex:])
			}
		}
	}

	if currentToolUse != nil {
		// Stream did not complete tool call, add it as partial
		if currentParamName != "" {
			// Tool call has a parameter that was not completed
			currentToolUse.Params[currentParamName] = strings.TrimSpace(accumulator[currentParamValueStartIndex:])
		}
		contentBlocks = append(contentBlocks, *currentToolUse)
	}

	// Note: it doesn't matter if check for currentToolUse or currentTextContent, only one of them will be defined since only one can be partial at a time
	if currentTextContent != nil {
		// Stream did not complete text content, add it as partial
		contentBlocks = append(contentBlocks, *currentTextContent)
	}

	return contentBlocks
}
