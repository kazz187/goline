package tui

import (
	"fmt"
	"github.com/abiosoft/ishell/v2"
	"io"
	"log/slog"
	"strings"

	ui "github.com/gizak/termui/v3"
)

// InputHandler handles input for the TUI
type InputHandler struct {
	ui            *UI
	integration   *REPLIntegration
	currentInput  string
	cursorPos     int
	historyIndex  int
	inputHistory  []string
	commandActive bool
	shell         *ishell.Shell
	shellInput    io.Writer
}

// GetCursorPosition returns the current cursor position
func (h *InputHandler) GetCursorPosition() int {
	return h.cursorPos
}

// NewInputHandler creates a new input handler
func NewInputHandler(ui *UI, integration *REPLIntegration, shell *ishell.Shell, shellInput io.Writer) *InputHandler {
	return &InputHandler{
		ui:           ui,
		integration:  integration,
		currentInput: "",
		cursorPos:    0,
		historyIndex: -1,
		inputHistory: []string{},
		shell:        shell,
		shellInput:   shellInput,
	}
}

// HandleKeyEvent handles a key event
func (h *InputHandler) HandleKeyEvent(e ui.Event) bool {
	switch e.ID {
	case "<C-c>":
		// Ctrl+C to exit
		return true
	case "<Enter>":
		// Enter to submit
		return h.handleEnter()
	case "<Backspace>":
		// Backspace to delete
		h.handleBackspace()
	case "<Delete>":
		// Delete to delete
		h.handleDelete()
	case "<Left>":
		// Left arrow to move cursor left
		h.handleLeft()
	case "<Right>":
		// Right arrow to move cursor right
		h.handleRight()
	case "<Home>":
		// Home to move cursor to beginning
		h.handleHome()
	case "<End>":
		// End to move cursor to end
		h.handleEnd()
	case "<Up>":
		// Up arrow to navigate history
		h.handleUp()
	case "<Down>":
		// Down arrow to navigate history
		h.handleDown()
	case "<C-a>":
		// Ctrl+A to move cursor to beginning
		h.handleHome()
	case "<C-e>":
		// Ctrl+E to move cursor to end
		h.handleEnd()
	case "<C-k>":
		// Ctrl+K to delete to end
		h.handleDeleteToEnd()
	case "<C-u>":
		// Ctrl+U to delete to beginning
		h.handleDeleteToBeginning()
	case "<Tab>":
		// Tab for auto-completion (not implemented yet)
		h.handleTab()
	case "<Space>":
		// Space character
		h.handleCharInput(" ")
	case "<C-d>":
		// Ctrl+D (EOF)
		if h.commandActive {
			// If we're in a multi-line input mode, submit the current input
			h.commandActive = false

			rootCmd := h.shell.RootCmd()
			// Get the command from the prompt
			cmdName := strings.TrimSuffix(rootCmd.Name, "> ")

			// Process the multi-line input based on the command
			if cmdName == "ask" {
				if h.currentInput == "" {
					h.integration.AddSystemMessage("Error: question is required")
				} else {
					// For the ask command, just add the user's question directly as a user message
					// without any system messages
					h.integration.AddUserInput(fmt.Sprintf("ask\n%s", h.currentInput))
				}
			} else {
				// For other commands, display the multi-line input completed message and the input content
				h.integration.AddSystemMessage("Multi-line input completed")

				// Display the input content as system messages
				if h.currentInput != "" {
					h.integration.AddSystemMessage("Input content:")
					lines := strings.Split(h.currentInput, "\n")
					for _, line := range lines {
						h.integration.AddSystemMessage(line)
					}
				}
			}

			// Clear the input
			h.currentInput = ""
			h.cursorPos = 0
			h.ui.UpdateREPLInput(h.currentInput)

			// Reset the prompt
			h.ui.UpdateREPLPrompt("goline> ")

			return false
		} else if h.currentInput == "" {
			// If input is empty, treat as exit command (common behavior in REPLs)
			h.integration.AddSystemMessage("EOF received, exiting...")
			return true
		} else {
			// Otherwise, just log it
			slog.Info("Ctrl+D (EOF) received")
		}
	default:
		// Regular character input
		if len(e.ID) == 1 {
			h.handleCharInput(e.ID)
		} else if e.ID == " " {
			// Alternative space representation
			h.handleCharInput(" ")
		}
	}

	// Update the UI
	h.ui.UpdateREPLInput(h.currentInput)

	return false
}

// handleEnter handles the Enter key
func (h *InputHandler) handleEnter() bool {
	if h.currentInput == "" {
		return false
	}

	// If we're in multi-line input mode, add a newline instead of submitting
	if h.commandActive {
		// Insert a newline at the cursor position
		before := h.currentInput[:h.cursorPos]
		after := h.currentInput[h.cursorPos:]
		h.currentInput = before + "\n" + after
		h.cursorPos = h.cursorPos + 1 // Move cursor after the newline

		// Update the UI
		h.ui.UpdateREPLInput(h.currentInput)
		return false
	}

	// Add to history
	h.inputHistory = append(h.inputHistory, h.currentInput)
	h.historyIndex = -1

	// Process the command
	command := strings.TrimSpace(h.currentInput)
	h.integration.AddUserInput(command)

	// Check for exit command
	if command == "exit" {
		return true
	}

	// Process the command
	h.processCommand(command)

	// Clear the input
	h.currentInput = ""
	h.cursorPos = 0
	h.ui.UpdateREPLInput(h.currentInput)

	// Reset the prompt if it was changed
	rootCmd := h.shell.RootCmd()
	h.ui.UpdateREPLPrompt(rootCmd.Name + "> ")

	return false
}

// processCommand processes a command
func (h *InputHandler) processCommand(command string) {
	// Split the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	// Get the command name
	cmdName := parts[0]

	// Process built-in commands
	switch cmdName {
	case "help":
		h.integration.AddSystemMessage("Available commands:")
		h.integration.AddSystemMessage("  help - Display help for REPL commands")
		h.integration.AddSystemMessage("  exit - Exit the REPL")
		h.integration.AddSystemMessage("  ask [question] - Ask the AI agent a question")
		h.integration.AddSystemMessage("  apply - Apply the AI agent's suggestion")
		h.integration.AddSystemMessage("  cancel - Cancel the AI agent's suggestion")
		h.integration.AddSystemMessage("  checkpoint save - Save the current task state as a checkpoint")
		h.integration.AddSystemMessage("  checkpoint restore [checkpointID] - Restore a previously saved checkpoint")
		h.integration.AddSystemMessage("  diff [checkpointID] - Show the difference between the current state and a checkpoint")
		h.integration.AddSystemMessage("  debug - Show debug information about the current input")
	case "debug":
		// Display debug information about the current input
		h.integration.AddSystemMessage("Debug information:")
		h.integration.AddSystemMessage(fmt.Sprintf("Input length: %d", len(h.currentInput)))
		h.integration.AddSystemMessage(fmt.Sprintf("Cursor position: %d", h.cursorPos))

		// Display the input with line numbers and cursor position
		lines := strings.Split(h.currentInput, "\n")
		for i, line := range lines {
			h.integration.AddSystemMessage(fmt.Sprintf("Line %d (%d chars): %s", i+1, len(line), line))
		}

		// Find which line the cursor is on
		pos := 0
		cursorLine := 0
		cursorCol := 0
		for i, line := range lines {
			lineLength := len(line)
			if pos+lineLength >= h.cursorPos {
				cursorLine = i
				cursorCol = h.cursorPos - pos
				break
			}
			pos += lineLength + 1 // +1 for the newline character
		}
		h.integration.AddSystemMessage(fmt.Sprintf("Cursor at line %d, column %d", cursorLine+1, cursorCol+1))

		// Display the input as a hex dump for debugging
		h.integration.AddSystemMessage("Input as hex:")
		hexDump := ""
		for i, c := range h.currentInput {
			if i == h.cursorPos {
				hexDump += "[CURSOR]"
			}
			hexDump += fmt.Sprintf("%02x ", c)
		}
		h.integration.AddSystemMessage(hexDump)
	case "ask":
		question := strings.TrimSpace(strings.TrimPrefix(command, "ask"))
		if question == "" {
			// Start multi-line input mode
			h.startMultiLineInput("ask")
			return
		}
		h.integration.AddSystemMessage("Sending question to AI agent...")
		h.integration.AddSystemMessage("TODO: Implement ask logic")
	case "apply":
		h.integration.AddSystemMessage("Applying AI agent's suggestion...")
		h.integration.AddSystemMessage("TODO: Implement apply logic")
	case "cancel":
		h.integration.AddSystemMessage("Cancelling AI agent's suggestion...")
		h.integration.AddSystemMessage("TODO: Implement cancel logic")
	case "checkpoint":
		if len(parts) < 2 {
			h.integration.AddSystemMessage("Error: checkpoint subcommand is required")
			return
		}
		switch parts[1] {
		case "save":
			h.integration.AddSystemMessage("Saving checkpoint...")
			h.integration.AddSystemMessage("TODO: Implement checkpoint save logic")
			h.integration.AddSystemMessage("Checkpoint ID: checkpoint-123")
		case "restore":
			if len(parts) < 3 {
				h.integration.AddSystemMessage("Error: checkpoint ID is required")
				return
			}
			checkpointID := parts[2]
			h.integration.AddSystemMessage(fmt.Sprintf("Restoring checkpoint %s...", checkpointID))
			h.integration.AddSystemMessage("TODO: Implement checkpoint restore logic")
		default:
			h.integration.AddSystemMessage(fmt.Sprintf("Error: unknown checkpoint subcommand: %s", parts[1]))
		}
	case "diff":
		if len(parts) < 2 {
			h.integration.AddSystemMessage("Error: checkpoint ID is required")
			return
		}
		checkpointID := parts[1]
		h.integration.AddSystemMessage(fmt.Sprintf("Showing diff for checkpoint %s...", checkpointID))
		h.integration.AddSystemMessage("TODO: Implement diff logic")
	default:
		h.integration.AddSystemMessage(fmt.Sprintf("Error: unknown command: %s", cmdName))
	}
}

// handleBackspace handles the Backspace key
func (h *InputHandler) handleBackspace() {
	if h.cursorPos > 0 {
		h.currentInput = h.currentInput[:h.cursorPos-1] + h.currentInput[h.cursorPos:]
		h.cursorPos--
	}
}

// handleDelete handles the Delete key
func (h *InputHandler) handleDelete() {
	if h.cursorPos < len(h.currentInput) {
		h.currentInput = h.currentInput[:h.cursorPos] + h.currentInput[h.cursorPos+1:]
	}
}

// handleLeft handles the Left arrow key
func (h *InputHandler) handleLeft() {
	if h.cursorPos > 0 {
		h.cursorPos--
	}
}

// handleRight handles the Right arrow key
func (h *InputHandler) handleRight() {
	if h.cursorPos < len(h.currentInput) {
		h.cursorPos++
	}
}

// handleHome handles the Home key
func (h *InputHandler) handleHome() {
	h.cursorPos = 0
}

// handleEnd handles the End key
func (h *InputHandler) handleEnd() {
	h.cursorPos = len(h.currentInput)
}

// handleUp handles the Up arrow key
func (h *InputHandler) handleUp() {
	// In multi-line input mode, move cursor up one line
	if h.commandActive && strings.Contains(h.currentInput, "\n") {
		// Get all lines
		allLines := strings.Split(h.currentInput, "\n")

		// Calculate line start positions
		lineStartPositions := make([]int, len(allLines))
		pos := 0
		for i := range allLines {
			lineStartPositions[i] = pos
			pos += len(allLines[i]) + 1 // +1 for the newline character
		}

		// Find which line the cursor is on
		currentLineIndex := -1
		for i := 0; i < len(allLines); i++ {
			if i < len(allLines)-1 {
				// Not the last line
				if lineStartPositions[i] <= h.cursorPos && h.cursorPos < lineStartPositions[i+1] {
					currentLineIndex = i
					break
				}
			} else {
				// Last line
				if lineStartPositions[i] <= h.cursorPos {
					currentLineIndex = i
					break
				}
			}
		}

		// If we couldn't determine the line or we're already at the first line, do nothing
		if currentLineIndex <= 0 {
			return
		}

		// Calculate column position on current line
		currentColPos := h.cursorPos - lineStartPositions[currentLineIndex]

		// Try to maintain the same column position on the previous line
		prevLineLength := len(allLines[currentLineIndex-1])
		newColPos := currentColPos
		if newColPos > prevLineLength {
			newColPos = prevLineLength
		}

		// Calculate the new absolute cursor position
		newPos := lineStartPositions[currentLineIndex-1] + newColPos

		// Update cursor position
		h.cursorPos = newPos

		return
	}

	// In normal mode, navigate command history
	if len(h.inputHistory) == 0 {
		return
	}

	if h.historyIndex == -1 {
		h.historyIndex = len(h.inputHistory) - 1
	} else if h.historyIndex > 0 {
		h.historyIndex--
	}

	h.currentInput = h.inputHistory[h.historyIndex]
	h.cursorPos = len(h.currentInput)
}

// handleDown handles the Down arrow key
func (h *InputHandler) handleDown() {
	// In multi-line input mode, move cursor down one line
	if h.commandActive && strings.Contains(h.currentInput, "\n") {
		// Get all lines
		allLines := strings.Split(h.currentInput, "\n")

		// Calculate line start positions
		lineStartPositions := make([]int, len(allLines))
		pos := 0
		for i := range allLines {
			lineStartPositions[i] = pos
			pos += len(allLines[i]) + 1 // +1 for the newline character
		}

		// Find which line the cursor is on
		currentLineIndex := -1
		for i := 0; i < len(allLines); i++ {
			if i < len(allLines)-1 {
				// Not the last line
				if lineStartPositions[i] <= h.cursorPos && h.cursorPos < lineStartPositions[i+1] {
					currentLineIndex = i
					break
				}
			} else {
				// Last line
				if lineStartPositions[i] <= h.cursorPos {
					currentLineIndex = i
					break
				}
			}
		}

		// If we couldn't determine the line or we're already at the last line, do nothing
		if currentLineIndex < 0 || currentLineIndex >= len(allLines)-1 {
			return
		}

		// Calculate column position on current line
		currentColPos := h.cursorPos - lineStartPositions[currentLineIndex]

		// Try to maintain the same column position on the next line
		nextLineLength := len(allLines[currentLineIndex+1])
		newColPos := currentColPos
		if newColPos > nextLineLength {
			newColPos = nextLineLength
		}

		// Calculate the new absolute cursor position
		newPos := lineStartPositions[currentLineIndex+1] + newColPos

		// Update cursor position
		h.cursorPos = newPos

		return
	}

	// In normal mode, navigate command history
	if h.historyIndex == -1 {
		return
	}

	if h.historyIndex < len(h.inputHistory)-1 {
		h.historyIndex++
		h.currentInput = h.inputHistory[h.historyIndex]
	} else {
		h.historyIndex = -1
		h.currentInput = ""
	}

	h.cursorPos = len(h.currentInput)
}

// handleDeleteToEnd handles Ctrl+K
func (h *InputHandler) handleDeleteToEnd() {
	if h.cursorPos < len(h.currentInput) {
		h.currentInput = h.currentInput[:h.cursorPos]
	}
}

// handleDeleteToBeginning handles Ctrl+U
func (h *InputHandler) handleDeleteToBeginning() {
	if h.cursorPos > 0 {
		h.currentInput = h.currentInput[h.cursorPos:]
		h.cursorPos = 0
	}
}

// handleTab handles the Tab key
func (h *InputHandler) handleTab() {
	// TODO: Implement auto-completion
	slog.Info("Tab pressed, auto-completion not implemented yet")
}

// handleCharInput handles character input
func (h *InputHandler) handleCharInput(char string) {
	h.currentInput = h.currentInput[:h.cursorPos] + char + h.currentInput[h.cursorPos:]
	h.cursorPos++
}

// startMultiLineInput starts multi-line input mode for a command
func (h *InputHandler) startMultiLineInput(command string) {
	h.commandActive = true

	// Only show the instruction message for commands other than "ask"
	if command != "ask" {
		h.integration.AddSystemMessage(fmt.Sprintf("Enter multi-line input for '%s' command (press Ctrl+D when done):", command))
	}

	// Update the prompt to indicate multi-line input mode
	h.ui.UpdateREPLPrompt(fmt.Sprintf("%s> ", command))
}
