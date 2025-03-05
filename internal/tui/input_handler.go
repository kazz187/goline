package tui

import (
	"fmt"
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
}

// NewInputHandler creates a new input handler
func NewInputHandler(ui *UI, integration *REPLIntegration) *InputHandler {
	return &InputHandler{
		ui:           ui,
		integration:  integration,
		currentInput: "",
		cursorPos:    0,
		historyIndex: -1,
		inputHistory: []string{},
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
	default:
		// Regular character input
		if len(e.ID) == 1 {
			h.handleCharInput(e.ID)
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
	case "ask":
		question := strings.TrimSpace(strings.TrimPrefix(command, "ask"))
		if question == "" {
			h.integration.AddSystemMessage("Error: question is required")
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
