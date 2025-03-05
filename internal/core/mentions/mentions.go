package mentions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MentionType represents the type of mention
type MentionType string

const (
	// FileMention represents a file mention
	FileMention MentionType = "file"
	// FolderMention represents a folder mention
	FolderMention MentionType = "folder"
	// ProblemsMention represents a problems mention
	ProblemsMention MentionType = "problems"
	// TerminalMention represents a terminal mention
	TerminalMention MentionType = "terminal"
	// GitChangesMention represents a git changes mention
	GitChangesMention MentionType = "git-changes"
	// GitCommitMention represents a git commit mention
	GitCommitMention MentionType = "git-commit"
	// URLMention represents a URL mention
	URLMention MentionType = "url"
	// UnknownMention represents an unknown mention
	UnknownMention MentionType = "unknown"
)

// Mention represents a mention in a message
type Mention struct {
	// Type of mention
	Type MentionType
	// Original text of the mention
	Original string
	// Processed text of the mention (e.g., file path without leading slash)
	Processed string
}

// mentionRegex is the regular expression for detecting mentions
var mentionRegex = regexp.MustCompile(`@([^\s]+)`)

// ParseMentions parses mentions in a message
func ParseMentions(text string) []Mention {
	var mentions []Mention

	matches := mentionRegex.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		mentionText := match[1]
		mention := Mention{
			Original: mentionText,
		}

		// Determine the type of mention
		if mentionText == "problems" {
			mention.Type = ProblemsMention
			mention.Processed = "Workspace Problems"
		} else if mentionText == "terminal" {
			mention.Type = TerminalMention
			mention.Processed = "Terminal Output"
		} else if mentionText == "git-changes" {
			mention.Type = GitChangesMention
			mention.Processed = "Working directory changes"
		} else if strings.HasPrefix(mentionText, "http") {
			mention.Type = URLMention
			mention.Processed = mentionText
		} else if strings.HasPrefix(mentionText, "/") {
			if strings.HasSuffix(mentionText, "/") {
				mention.Type = FolderMention
				mention.Processed = mentionText[1:] // Remove leading slash
			} else {
				mention.Type = FileMention
				mention.Processed = mentionText[1:] // Remove leading slash
			}
		} else if isGitCommitHash(mentionText) {
			mention.Type = GitCommitMention
			mention.Processed = fmt.Sprintf("Git commit '%s'", mentionText)
		} else {
			mention.Type = UnknownMention
			mention.Processed = mentionText
		}

		mentions = append(mentions, mention)
	}

	return mentions
}

// isGitCommitHash checks if a string is a git commit hash
func isGitCommitHash(s string) bool {
	// Git commit hashes are hexadecimal and typically 7-40 characters long
	match, _ := regexp.MatchString(`^[a-f0-9]{7,40}$`, s)
	return match
}

// ReplaceMentionsWithContent replaces mentions in a message with their content
func ReplaceMentionsWithContent(text string, cwd string) (string, error) {
	mentions := ParseMentions(text)

	// First, replace mentions in the text with their descriptions
	parsedText := mentionRegex.ReplaceAllStringFunc(text, func(match string) string {
		mentionText := match[1:] // Remove @ symbol

		for _, mention := range mentions {
			if mention.Original == mentionText {
				switch mention.Type {
				case FileMention:
					return fmt.Sprintf("'%s' (see below for file content)", mention.Processed)
				case FolderMention:
					return fmt.Sprintf("'%s' (see below for folder content)", mention.Processed)
				case ProblemsMention:
					return "Workspace Problems (see below for diagnostics)"
				case TerminalMention:
					return "Terminal Output (see below for output)"
				case GitChangesMention:
					return "Working directory changes (see below for details)"
				case GitCommitMention:
					return fmt.Sprintf("%s (see below for commit info)", mention.Processed)
				case URLMention:
					return fmt.Sprintf("'%s' (see below for site content)", mention.Processed)
				default:
					return match
				}
			}
		}

		return match
	})

	// Then, append the content for each mention
	for _, mention := range mentions {
		var content string
		var err error

		switch mention.Type {
		case FileMention:
			content, err = getFileContent(filepath.Join(cwd, mention.Processed))
			if err != nil {
				content = fmt.Sprintf("Error fetching content: %s", err.Error())
			}
			parsedText += fmt.Sprintf("\n\n<file_content path=\"%s\">\n%s\n</file_content>", mention.Processed, content)

		case FolderMention:
			content, err = getFolderContent(filepath.Join(cwd, mention.Processed), cwd)
			if err != nil {
				content = fmt.Sprintf("Error fetching content: %s", err.Error())
			}
			parsedText += fmt.Sprintf("\n\n<folder_content path=\"%s\">\n%s\n</folder_content>", mention.Processed, content)

		case ProblemsMention:
			// In a real implementation, this would fetch workspace diagnostics
			content = "No errors or warnings detected."
			parsedText += fmt.Sprintf("\n\n<workspace_diagnostics>\n%s\n</workspace_diagnostics>", content)

		case TerminalMention:
			// In a real implementation, this would fetch terminal output
			content = "No terminal output available."
			parsedText += fmt.Sprintf("\n\n<terminal_output>\n%s\n</terminal_output>", content)

		case GitChangesMention:
			// In a real implementation, this would fetch git working state
			content = "No git changes detected."
			parsedText += fmt.Sprintf("\n\n<git_working_state>\n%s\n</git_working_state>", content)

		case GitCommitMention:
			// In a real implementation, this would fetch git commit info
			content = fmt.Sprintf("Commit information for '%s' not available.", mention.Original)
			parsedText += fmt.Sprintf("\n\n<git_commit hash=\"%s\">\n%s\n</git_commit>", mention.Original, content)

		case URLMention:
			// In a real implementation, this would fetch URL content
			content = fmt.Sprintf("Content for URL '%s' not available.", mention.Original)
			parsedText += fmt.Sprintf("\n\n<url_content url=\"%s\">\n%s\n</url_content>", mention.Original, content)
		}
	}

	return parsedText, nil
}

// getFileContent reads the content of a file
func getFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// getFolderContent gets the content of a folder
func getFolderContent(path string, cwd string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var folderContent strings.Builder
	var fileContentPromises []string

	for i, entry := range entries {
		isLast := i == len(entries)-1
		linePrefix := "└── "
		if !isLast {
			linePrefix = "├── "
		}

		if entry.IsDir() {
			folderContent.WriteString(fmt.Sprintf("%s%s/\n", linePrefix, entry.Name()))
			// Not recursively getting folder contents
		} else {
			folderContent.WriteString(fmt.Sprintf("%s%s\n", linePrefix, entry.Name()))

			// In a real implementation, we would read file contents here
			// For now, we'll just add placeholders for non-binary files
			filePath := filepath.Join(path, entry.Name())
			relPath, _ := filepath.Rel(cwd, filePath)

			// Check if file is binary (simplified check)
			if !isBinaryFile(filePath) {
				content, err := getFileContent(filePath)
				if err == nil {
					fileContentPromises = append(fileContentPromises,
						fmt.Sprintf("<file_content path=\"%s\">\n%s\n</file_content>",
							filepath.ToSlash(relPath), content))
				}
			}
		}
	}

	if folderContent.Len() == 0 {
		return "(Empty folder)", nil
	}

	result := folderContent.String()
	if len(fileContentPromises) > 0 {
		result += "\n" + strings.Join(fileContentPromises, "\n\n")
	}

	return strings.TrimSpace(result), nil
}

// isBinaryFile checks if a file is binary (simplified implementation)
func isBinaryFile(path string) bool {
	// This is a simplified check - in a real implementation, we would use a more robust method
	ext := strings.ToLower(filepath.Ext(path))
	binaryExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".zip": true, ".tar": true, ".gz": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true,
	}

	return binaryExtensions[ext]
}

// OpenMention opens a mention (e.g., file, folder, URL)
// This is a placeholder for the actual implementation that would integrate with the UI
func OpenMention(mention string, cwd string) error {
	if mention == "" {
		return nil
	}

	if mention == "problems" {
		// Open problems view
		return nil
	} else if mention == "terminal" {
		// Focus terminal
		return nil
	} else if strings.HasPrefix(mention, "http") {
		// Open URL in browser
		return nil
	} else if strings.HasPrefix(mention, "/") {
		// absPath is used in the actual implementation
		// filepath.Join(cwd, mention[1:])

		if strings.HasSuffix(mention, "/") {
			// Reveal folder in explorer
			return nil
		} else {
			// Open file
			return nil
		}
	}

	return fmt.Errorf("unknown mention type: %s", mention)
}
