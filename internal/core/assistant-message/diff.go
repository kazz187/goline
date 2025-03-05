package assistantmessage

import (
	"errors"
	"regexp"
	"strings"
)

// Markers for diff blocks
const (
	SearchMarker  = "<<<<<<< SEARCH"
	DividerMarker = "======="
	ReplaceMarker = ">>>>>>> REPLACE"
)

// LineTrimmedFallbackMatch attempts a line-trimmed fallback match for the given search content in the original content.
// It returns the start and end indices of the match if found, or an error if not found.
func LineTrimmedFallbackMatch(originalContent, searchContent string, startIndex int) (int, int, error) {
	// Split both contents into lines
	originalLines := strings.Split(originalContent, "\n")
	searchLines := strings.Split(searchContent, "\n")

	// Trim trailing empty line if exists (from the trailing \n in searchContent)
	if len(searchLines) > 0 && searchLines[len(searchLines)-1] == "" {
		searchLines = searchLines[:len(searchLines)-1]
	}

	// Find the line number where startIndex falls
	startLineNum := 0
	currentIndex := 0
	for currentIndex < startIndex && startLineNum < len(originalLines) {
		currentIndex += len(originalLines[startLineNum]) + 1 // +1 for \n
		startLineNum++
	}

	// For each possible starting position in original content
	for i := startLineNum; i <= len(originalLines)-len(searchLines); i++ {
		matches := true

		// Try to match all search lines from this position
		for j := 0; j < len(searchLines); j++ {
			originalTrimmed := strings.TrimSpace(originalLines[i+j])
			searchTrimmed := strings.TrimSpace(searchLines[j])

			if originalTrimmed != searchTrimmed {
				matches = false
				break
			}
		}

		// If we found a match, calculate the exact character positions
		if matches {
			// Find start character index
			matchStartIndex := 0
			for k := 0; k < i; k++ {
				matchStartIndex += len(originalLines[k]) + 1 // +1 for \n
			}

			// Find end character index
			matchEndIndex := matchStartIndex
			for k := 0; k < len(searchLines); k++ {
				matchEndIndex += len(originalLines[i+k]) + 1 // +1 for \n
			}

			return matchStartIndex, matchEndIndex, nil
		}
	}

	return 0, 0, errors.New("no line-trimmed match found")
}

// BlockAnchorFallbackMatch attempts to match blocks of code by using the first and last lines as anchors.
// It returns the start and end indices of the match if found, or an error if not found.
func BlockAnchorFallbackMatch(originalContent, searchContent string, startIndex int) (int, int, error) {
	originalLines := strings.Split(originalContent, "\n")
	searchLines := strings.Split(searchContent, "\n")

	// Only use this approach for blocks of 3+ lines
	if len(searchLines) < 3 {
		return 0, 0, errors.New("search content too short for block anchor match")
	}

	// Trim trailing empty line if exists
	if len(searchLines) > 0 && searchLines[len(searchLines)-1] == "" {
		searchLines = searchLines[:len(searchLines)-1]
	}

	firstLineSearch := strings.TrimSpace(searchLines[0])
	lastLineSearch := strings.TrimSpace(searchLines[len(searchLines)-1])
	searchBlockSize := len(searchLines)

	// Find the line number where startIndex falls
	startLineNum := 0
	currentIndex := 0
	for currentIndex < startIndex && startLineNum < len(originalLines) {
		currentIndex += len(originalLines[startLineNum]) + 1
		startLineNum++
	}

	// Look for matching start and end anchors
	for i := startLineNum; i <= len(originalLines)-searchBlockSize; i++ {
		// Check if first line matches
		if strings.TrimSpace(originalLines[i]) != firstLineSearch {
			continue
		}

		// Check if last line matches at the expected position
		if strings.TrimSpace(originalLines[i+searchBlockSize-1]) != lastLineSearch {
			continue
		}

		// Calculate exact character positions
		matchStartIndex := 0
		for k := 0; k < i; k++ {
			matchStartIndex += len(originalLines[k]) + 1
		}

		matchEndIndex := matchStartIndex
		for k := 0; k < searchBlockSize; k++ {
			matchEndIndex += len(originalLines[i+k]) + 1
		}

		return matchStartIndex, matchEndIndex, nil
	}

	return 0, 0, errors.New("no block anchor match found")
}

// ConstructNewFileContent reconstructs the file content by applying a streamed diff to the original file content.
func ConstructNewFileContent(diffContent, originalContent string, isFinal bool) (string, error) {
	result := ""
	lastProcessedIndex := 0

	currentSearchContent := ""
	currentReplaceContent := ""
	inSearch := false
	inReplace := false

	searchMatchIndex := -1
	searchEndIndex := -1

	lines := strings.Split(diffContent, "\n")

	// If the last line looks like a partial marker but isn't recognized,
	// remove it because it might be incomplete.
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if (strings.HasPrefix(lastLine, "<") || strings.HasPrefix(lastLine, "=") || strings.HasPrefix(lastLine, ">")) &&
			lastLine != SearchMarker && lastLine != DividerMarker && lastLine != ReplaceMarker {
			lines = lines[:len(lines)-1]
		}
	}

	for _, line := range lines {
		if line == SearchMarker {
			inSearch = true
			currentSearchContent = ""
			currentReplaceContent = ""
			continue
		}

		if line == DividerMarker {
			inSearch = false
			inReplace = true

			if currentSearchContent == "" {
				// Empty search block
				if len(originalContent) == 0 {
					// New file scenario: nothing to match, just start inserting
					searchMatchIndex = 0
					searchEndIndex = 0
				} else {
					// Complete file replacement scenario: treat the entire file as matched
					searchMatchIndex = 0
					searchEndIndex = len(originalContent)
				}
			} else {
				// Exact search match scenario
				exactIndex := strings.Index(originalContent[lastProcessedIndex:], currentSearchContent)
				if exactIndex != -1 {
					searchMatchIndex = lastProcessedIndex + exactIndex
					searchEndIndex = searchMatchIndex + len(currentSearchContent)
				} else {
					// Attempt fallback line-trimmed matching
					matchStart, matchEnd, err := LineTrimmedFallbackMatch(originalContent, currentSearchContent, lastProcessedIndex)
					if err == nil {
						searchMatchIndex = matchStart
						searchEndIndex = matchEnd
					} else {
						// Try block anchor fallback for larger blocks
						matchStart, matchEnd, err := BlockAnchorFallbackMatch(originalContent, currentSearchContent, lastProcessedIndex)
						if err == nil {
							searchMatchIndex = matchStart
							searchEndIndex = matchEnd
						} else {
							return "", errors.New("the SEARCH block does not match anything in the file")
						}
					}
				}
			}

			// Output everything up to the match location
			result += originalContent[lastProcessedIndex:searchMatchIndex]
			continue
		}

		if line == ReplaceMarker {
			// Finished one replace block

			// Advance lastProcessedIndex to after the matched section
			lastProcessedIndex = searchEndIndex

			// Reset for next block
			inSearch = false
			inReplace = false
			currentSearchContent = ""
			currentReplaceContent = ""
			searchMatchIndex = -1
			searchEndIndex = -1
			continue
		}

		// Accumulate content for search or replace
		if inSearch {
			currentSearchContent += line + "\n"
		} else if inReplace {
			currentReplaceContent += line + "\n"
			// Output replacement lines immediately if we know the insertion point
			if searchMatchIndex != -1 {
				result += line + "\n"
			}
		}
	}

	// If this is the final chunk, append any remaining original content
	if isFinal && lastProcessedIndex < len(originalContent) {
		result += originalContent[lastProcessedIndex:]
	}

	return result, nil
}

// ParseDiff parses a diff string into search and replace blocks
func ParseDiff(diffContent string) ([]map[string]string, error) {
	var blocks []map[string]string

	// Regular expression to match SEARCH/REPLACE blocks
	re := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(SearchMarker) + `\n(.*?)\n` + regexp.QuoteMeta(DividerMarker) + `\n(.*?)\n` + regexp.QuoteMeta(ReplaceMarker))

	matches := re.FindAllStringSubmatch(diffContent, -1)

	for _, match := range matches {
		if len(match) == 3 {
			block := map[string]string{
				"search":  match[1],
				"replace": match[2],
			}
			blocks = append(blocks, block)
		}
	}

	if len(blocks) == 0 {
		return nil, errors.New("no valid SEARCH/REPLACE blocks found in diff")
	}

	return blocks, nil
}
