package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/abiosoft/ishell/v2"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/mattn/go-runewidth"
)

// TaskInfo represents the information about a task
type TaskInfo struct {
	ID        string
	Status    string
	StartTime time.Time
	Provider  string
	Engine    string
}

// HistoryEntry represents an entry in the task history
type HistoryEntry struct {
	Timestamp time.Time
	Type      string // "user", "agent", "system"
	Content   string
}

// InputHandlerInterface defines the interface for input handlers
type InputHandlerInterface interface {
	HandleKeyEvent(e ui.Event) bool
	GetCursorPosition() int
}

// UI represents the TUI
type UI struct {
	taskInfo      *widgets.Paragraph
	historyList   *widgets.List
	replParagraph *widgets.Paragraph
	shell         *ishell.Shell
	grid          *ui.Grid
	taskInfoData  TaskInfo
	historyData   []HistoryEntry
	replInput     string
	inputHandler  InputHandlerInterface
}

// NewUI creates a new TUI
func NewUI(shell *ishell.Shell) (*UI, error) {
	if err := ui.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize termui: %w", err)
	}

	// Create task info widget (optimized for single line display)
	taskInfo := widgets.NewParagraph()
	taskInfo.Title = "Task Information"
	taskInfo.BorderStyle.Fg = ui.ColorYellow
	taskInfo.PaddingTop = 0
	taskInfo.PaddingBottom = 0

	// Create history widget
	historyList := widgets.NewList()
	historyList.Title = "Task History"
	historyList.BorderStyle.Fg = ui.ColorCyan
	historyList.TextStyle = ui.NewStyle(ui.ColorWhite)
	historyList.WrapText = true

	// Ensure we have a reasonable minimum width for the history list
	// This helps prevent vertical text rendering before the UI is fully initialized
	termWidth, termHeight := ui.TerminalDimensions()
	if termWidth < 20 {
		termWidth = 80 // Default to 80 columns if terminal width is too small
	}

	// Create REPL widget
	replParagraph := widgets.NewParagraph()
	replParagraph.Title = "Command Input"
	replParagraph.BorderStyle.Fg = ui.ColorGreen

	// Create grid layout
	grid := ui.NewGrid()
	// Use the already declared variables
	termWidth, termHeight = ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	// Set up the grid with three rows
	// Task info: just enough for a single line (3 fixed height units)
	// History: takes up most of the remaining space
	// REPL: initially 15% of height, but will be adjusted dynamically
	grid.Set(
		ui.NewRow(3.0/float64(termHeight),
			ui.NewCol(1.0, taskInfo),
		),
		ui.NewRow(0.85-(3.0/float64(termHeight)),
			ui.NewCol(1.0, historyList),
		),
		ui.NewRow(0.15,
			ui.NewCol(1.0, replParagraph),
		),
	)

	return &UI{
		taskInfo:      taskInfo,
		historyList:   historyList,
		replParagraph: replParagraph,
		shell:         shell,
		grid:          grid,
		taskInfoData: TaskInfo{
			ID:        "task-123",
			Status:    "Active",
			StartTime: time.Now(),
			Provider:  "anthropic",
			Engine:    "claude-3-opus",
		},
		historyData: []HistoryEntry{},
		replInput:   "",
	}, nil
}

// UpdateTaskInfo updates the task info widget
func (u *UI) UpdateTaskInfo(taskInfo TaskInfo) {
	u.taskInfoData = taskInfo
	u.Render() // Render the entire UI
}

// AddHistoryEntry adds an entry to the history widget
func (u *UI) AddHistoryEntry(entry HistoryEntry) {
	u.historyData = append(u.historyData, entry)
	u.Render() // Render the entire UI
}

// UpdateREPLInput updates the REPL input widget
func (u *UI) UpdateREPLInput(input string) {
	u.replInput = input
	u.Render() // Render the entire UI
}

// UpdateREPLPrompt updates the REPL prompt
func (u *UI) UpdateREPLPrompt(prompt string) {
	u.replParagraph.Title = prompt
	u.Render() // Render the entire UI
}

// renderTaskInfo updates the task info widget content
func (u *UI) renderTaskInfo() {
	elapsed := time.Since(u.taskInfoData.StartTime).Round(time.Second)
	text := fmt.Sprintf("ID: %s | Status: %s | Elapsed: %s | Provider: %s | Engine: %s",
		u.taskInfoData.ID,
		u.taskInfoData.Status,
		elapsed,
		u.taskInfoData.Provider,
		u.taskInfoData.Engine,
	)

	// Ensure the text fits within the available width
	availableWidth := u.taskInfo.Inner.Dx()
	if runewidth.StringWidth(text) > availableWidth {
		// Truncate the text to fit within the available width
		// Leave room for "..." at the end
		truncated := ""
		currentWidth := 0
		for _, char := range text {
			charWidth := runewidth.RuneWidth(char)
			if currentWidth+charWidth+3 > availableWidth { // +3 for "..."
				break
			}
			truncated += string(char)
			currentWidth += charWidth
		}
		text = truncated + "..."
	}

	u.taskInfo.Text = text
	// No individual rendering here, will be rendered by Render()
}

// renderHistory updates the history widget content
func (u *UI) renderHistory() {
	u.historyList.Rows = []string{}

	// Get the width, ensuring it's at least a reasonable minimum
	width := u.historyList.Inner.Dx()
	if width < 20 {
		// Use a reasonable default if width is too small
		width = 80
	}

	for _, entry := range u.historyData {
		timestamp := entry.Timestamp.Format("15:04:05")
		prefix := ""
		switch entry.Type {
		case "user":
			prefix = "[User]"
		case "agent":
			prefix = "[Agent]"
		case "system":
			prefix = "[System]"
		}
		text := fmt.Sprintf("[%s] %s %s", timestamp, prefix, entry.Content)
		wrappedText := wrapText(text, width)
		u.historyList.Rows = append(u.historyList.Rows, wrappedText...)
	}
	// No individual rendering here, will be rendered by Render()
}

// renderREPL updates the REPL widget content
func (u *UI) renderREPL() {
	// Default prompt is "goline> ", but can be changed with UpdateREPLPrompt
	prompt := "goline> "
	if u.replParagraph.Title != "Command Input" {
		// If the title has been changed, use it as the prompt base
		prompt = u.replParagraph.Title
	}

	// Get the available width for text
	availableWidth := u.replParagraph.Inner.Dx()
	if availableWidth < 10 {
		availableWidth = 80 // Default to a reasonable width if too small
	}

	// Account for prompt length in available width
	promptWidth := runewidth.StringWidth(prompt)
	textWidth := availableWidth - promptWidth

	// If we have an input handler with a cursor position, display the cursor
	if u.inputHandler != nil {
		// Get the cursor position from the input handler
		cursorPos := u.inputHandler.GetCursorPosition()

		if cursorPos <= len(u.replInput) {
			// Split the input at the cursor position
			before := u.replInput[:cursorPos]
			after := ""
			if cursorPos < len(u.replInput) {
				after = u.replInput[cursorPos:]
			}

			// Handle multi-line input and cursor positioning
			lines := []string{}

			// Process text before cursor
			beforeLines := strings.Split(before, "\n")

			// Process all lines except the last one (where cursor is)
			for i := 0; i < len(beforeLines)-1; i++ {
				// Wrap each line to fit within available width
				wrappedLines := wrapTextForREPL(beforeLines[i], textWidth)
				lines = append(lines, wrappedLines...)
			}

			// Process the line with cursor
			var currentLine string
			if len(beforeLines) > 0 {
				currentLine = beforeLines[len(beforeLines)-1]
			} else {
				currentLine = ""
			}

			// If the current line is too long, we need to scroll horizontally
			currentLineWidth := runewidth.StringWidth(currentLine)

			// Determine visible portion of the current line
			visibleCurrentLine := currentLine
			if currentLineWidth > textWidth {
				// Show the end of the line with the cursor
				// Calculate how many characters we can show
				visibleStart := 0
				visibleWidth := 0
				for i, char := range currentLine {
					charWidth := runewidth.RuneWidth(char)
					if visibleWidth+charWidth > textWidth-5 { // Leave space for cursor and some context
						visibleStart = i
						break
					}
					visibleWidth += charWidth
				}

				// If we need to truncate, add an indicator
				if visibleStart > 0 {
					visibleCurrentLine = "..." + currentLine[visibleStart:]
				}
			}

			// Add the cursor line (with cursor)
			cursorLine := visibleCurrentLine + "█"

			// If there's text after the cursor, add it
			if after != "" {
				afterLines := strings.Split(after, "\n")

				// Add the first line of 'after' to the cursor line
				cursorLine += afterLines[0]

				// Process remaining lines after cursor
				for i := 1; i < len(afterLines); i++ {
					// Wrap each line to fit within available width
					wrappedLines := wrapTextForREPL(afterLines[i], textWidth)
					lines = append(lines, wrappedLines...)
				}
			}

			// Add the cursor line to our lines
			lines = append(lines, cursorLine)

			// Join all lines with newlines
			displayText := strings.Join(lines, "\n")

			// Set the text with proper prompt
			if len(lines) > 1 {
				// For multi-line, only add prompt to the first line
				firstNewline := strings.Index(displayText, "\n")
				if firstNewline >= 0 {
					u.replParagraph.Text = prompt + displayText[:firstNewline] + "\n" + displayText[firstNewline+1:]
				} else {
					u.replParagraph.Text = prompt + displayText
				}
			} else {
				// Single line case
				u.replParagraph.Text = prompt + displayText
			}
		} else {
			// Fallback if cursor position is invalid
			u.replParagraph.Text = prompt + u.replInput + "█"
		}
	} else {
		// No input handler, just show the input without cursor
		// Still need to handle multi-line display
		lines := strings.Split(u.replInput, "\n")

		// Wrap each line to fit within available width
		var wrappedLines []string
		for _, line := range lines {
			wrapped := wrapTextForREPL(line, textWidth)
			wrappedLines = append(wrappedLines, wrapped...)
		}

		// Join all lines with newlines
		displayText := strings.Join(wrappedLines, "\n")

		// Set the text with proper prompt
		if len(wrappedLines) > 1 {
			// For multi-line, only add prompt to the first line
			firstNewline := strings.Index(displayText, "\n")
			if firstNewline >= 0 {
				u.replParagraph.Text = prompt + displayText[:firstNewline] + "\n" + displayText[firstNewline+1:]
			} else {
				u.replParagraph.Text = prompt + displayText
			}
		} else {
			// Single line case
			u.replParagraph.Text = prompt + displayText
		}
	}

	// No individual rendering here, will be rendered by Render()
}

// wrapTextForREPL wraps text specifically for the REPL display
// This is a simplified version of wrapText that doesn't add each line to an array
func wrapTextForREPL(text string, width int) []string {
	var lines []string

	// Ensure width is reasonable
	if width <= 0 {
		width = 80 // Default to 80 columns if width is invalid
	}

	totalWidth := runewidth.StringWidth(text)
	if totalWidth <= width {
		return []string{text}
	}

	// Wrapping algorithm
	current := ""
	currentWidth := 0

	for _, char := range text {
		charWidth := runewidth.RuneWidth(char)

		// If adding this character would exceed the width, start a new line
		if currentWidth+charWidth > width {
			lines = append(lines, current)
			current = string(char)
			currentWidth = charWidth
		} else {
			current += string(char)
			currentWidth += charWidth
		}
	}

	// Don't forget the last line
	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// Render renders the UI
func (u *UI) Render() {
	// Adjust the grid layout based on the number of input lines
	u.adjustGridLayout()

	// Update the content of each component
	u.renderTaskInfo()
	u.renderHistory()
	u.renderREPL()

	// Render the entire grid at once instead of individual components
	ui.Render(u.grid)
}

// adjustGridLayout adjusts the grid layout based on the number of input lines
func (u *UI) adjustGridLayout() {
	// Get the terminal dimensions
	termWidth, termHeight := ui.TerminalDimensions()

	// Count the number of lines in the input
	lineCount := 1
	if u.replInput != "" {
		lineCount = len(strings.Split(u.replInput, "\n"))
	}

	// Calculate the height ratio for the REPL input area
	// Base height is 15% of the terminal height
	// Add 2% for each additional line, up to a maximum of 40%
	replHeightRatio := 0.15
	if lineCount > 1 {
		additionalHeight := float64(lineCount-1) * 0.02
		replHeightRatio = math.Min(0.40, replHeightRatio+additionalHeight)
	}

	// Calculate the history height ratio
	historyHeightRatio := 0.85 - (3.0 / float64(termHeight)) - replHeightRatio

	// Update the grid layout
	u.grid.SetRect(0, 0, termWidth, termHeight)
	u.grid.Set(
		ui.NewRow(3.0/float64(termHeight),
			ui.NewCol(1.0, u.taskInfo),
		),
		ui.NewRow(historyHeightRatio,
			ui.NewCol(1.0, u.historyList),
		),
		ui.NewRow(replHeightRatio,
			ui.NewCol(1.0, u.replParagraph),
		),
	)
}

// Close closes the UI
func (u *UI) Close() {
	ui.Close()
}

// Run runs the UI
func (u *UI) Run() error {
	// Initial render
	u.Render()

	// Set up event handling
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				if u.inputHandler != nil {
					// Let the input handler process the event
					if exit := u.inputHandler.HandleKeyEvent(e); exit {
						return nil
					}
				} else {
					// Default keyboard handling if no input handler
					switch e.ID {
					case "q", "<C-c>":
						return nil
					}
				}
			} else if e.Type == ui.ResizeEvent {
				payload := e.Payload.(ui.Resize)
				u.grid.SetRect(0, 0, payload.Width, payload.Height)
				u.Render()
			}
		case <-ticker.C:
			// Update elapsed time every second and re-render the entire UI
			u.Render()
		}
	}
}

// SetInputHandler sets the input handler for the UI
func (u *UI) SetInputHandler(handler InputHandlerInterface) {
	// This will be called by the REPLIntegration
	u.inputHandler = handler
}

// wrapText wraps text to fit within a given width
func wrapText(text string, width int) []string {
	var lines []string

	// Ensure width is reasonable
	if width <= 0 {
		width = 80 // Default to 80 columns if width is invalid
	}

	totalWidth := runewidth.StringWidth(text)
	if totalWidth <= width {
		return []string{text}
	}

	// Improved wrapping algorithm
	current := ""
	currentWidth := 0

	for _, char := range text {
		charWidth := runewidth.RuneWidth(char)

		// If adding this character would exceed the width, start a new line
		if currentWidth+charWidth > width {
			lines = append(lines, current)
			current = string(char)
			currentWidth = charWidth
		} else {
			current += string(char)
			currentWidth += charWidth
		}

		// If we hit a newline character, respect it
		if char == '\n' {
			lines = append(lines, current)
			current = ""
			currentWidth = 0
		}
	}

	// Don't forget the last line
	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// The StartTUI function has been removed as it was causing duplicate UI instances
