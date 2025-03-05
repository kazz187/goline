package prompts

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// FormatResponse contains functions for formatting responses
type FormatResponse struct{}

// NewFormatResponse creates a new FormatResponse
func NewFormatResponse() *FormatResponse {
	return &FormatResponse{}
}

// ToolDenied returns a message for when a tool is denied
func (f *FormatResponse) ToolDenied() string {
	return "The user denied this operation."
}

// ToolError returns a message for when a tool execution fails
func (f *FormatResponse) ToolError(error string) string {
	return fmt.Sprintf("The tool execution failed with the following error:\n<error>\n%s\n</error>", error)
}

// ClineIgnoreError returns a message for when a file is blocked by .clineignore
func (f *FormatResponse) ClineIgnoreError(path string) string {
	return fmt.Sprintf("Access to %s is blocked by the .clineignore file settings. You must try to continue in the task without using this file, or ask the user to update the .clineignore file.", path)
}

// NoToolsUsed returns a message for when no tools are used
func (f *FormatResponse) NoToolsUsed() string {
	return `[ERROR] You did not use a tool in your previous response! Please retry with a tool use.

` + toolUseInstructionsReminder + `

# Next Steps

If you have completed the user's task, use the attempt_completion tool. 
If you require additional information from the user, use the ask_followup_question tool. 
Otherwise, if you have not completed the task and do not need additional information, then proceed with the next step of the task. 
(This is an automated message, so do not respond to it conversationally.)`
}

// TooManyMistakes returns a message for when there are too many mistakes
func (f *FormatResponse) TooManyMistakes(feedback string) string {
	return fmt.Sprintf("You seem to be having trouble proceeding. The user has provided the following feedback to help guide you:\n<feedback>\n%s\n</feedback>", feedback)
}

// MissingToolParameterError returns a message for when a tool parameter is missing
func (f *FormatResponse) MissingToolParameterError(paramName string) string {
	return fmt.Sprintf("Missing value for required parameter '%s'. Please retry with complete response.\n\n%s", paramName, toolUseInstructionsReminder)
}

// InvalidMcpToolArgumentError returns a message for when an MCP tool argument is invalid
func (f *FormatResponse) InvalidMcpToolArgumentError(serverName, toolName string) string {
	return fmt.Sprintf("Invalid JSON argument used with %s for %s. Please retry with a properly formatted JSON argument.", serverName, toolName)
}

// ToolResult returns a message for a tool result
func (f *FormatResponse) ToolResult(text string) string {
	return text
}

// FormatFilesList formats a list of files
func (f *FormatResponse) FormatFilesList(absolutePath string, files []string, didHitLimit bool) string {
	sorted := make([]string, len(files))
	for i, file := range files {
		// Convert absolute path to relative path
		relativePath, err := filepath.Rel(absolutePath, file)
		if err != nil {
			relativePath = file
		}
		// Convert to forward slashes for consistency
		relativePath = filepath.ToSlash(relativePath)
		if strings.HasSuffix(file, "/") || strings.HasSuffix(file, "\\") {
			relativePath += "/"
		}
		sorted[i] = relativePath
	}

	// Sort files so they are listed under their respective directories
	// This makes it clear what files are children of what directories
	sort.Strings(sorted)

	if didHitLimit {
		return fmt.Sprintf("%s\n\n(File list truncated. Use list_files on specific subdirectories if you need to explore further.)", strings.Join(sorted, "\n"))
	} else if len(sorted) == 0 || (len(sorted) == 1 && sorted[0] == "") {
		return "No files found."
	} else {
		return strings.Join(sorted, "\n")
	}
}

// CreatePrettyPatch creates a pretty patch for a diff
func (f *FormatResponse) CreatePrettyPatch(filename string, oldStr, newStr string) string {
	// This is a simplified version - in a real implementation, we would use a diff library
	if filename == "" {
		filename = "file"
	}
	if oldStr == "" {
		oldStr = ""
	}
	if newStr == "" {
		newStr = ""
	}

	// Simple line-by-line diff
	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	var result strings.Builder
	result.WriteString(fmt.Sprintf("--- %s\n", filename))
	result.WriteString(fmt.Sprintf("+++ %s\n", filename))

	// Very simple diff algorithm - just show removed lines with - and added lines with +
	for _, line := range oldLines {
		if !containsLine(newLines, line) {
			result.WriteString(fmt.Sprintf("-%s\n", line))
		}
	}
	for _, line := range newLines {
		if !containsLine(oldLines, line) {
			result.WriteString(fmt.Sprintf("+%s\n", line))
		}
	}

	return result.String()
}

// containsLine checks if a slice of strings contains a specific line
func containsLine(lines []string, line string) bool {
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}

// toolUseInstructionsReminder is a reminder for tool use instructions
const toolUseInstructionsReminder = `# Reminder: Instructions for Tool Use

Tool uses are formatted using XML-style tags. The tool name is enclosed in opening and closing tags, and each parameter is similarly enclosed within its own set of tags. Here's the structure:

<tool_name>
<parameter1_name>value1</parameter1_name>
<parameter2_name>value2</parameter2_name>
...
</tool_name>

For example:

<attempt_completion>
<result>
I have completed the task...
</result>
</attempt_completion>

Always adhere to this format for all tool uses to ensure proper parsing and execution.`
