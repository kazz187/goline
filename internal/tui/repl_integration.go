package tui

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/abiosoft/ishell/v2"
)

// REPLIntegration represents the integration between the TUI and the REPL
type REPLIntegration struct {
	ui    *UI
	shell *ishell.Shell
	mutex sync.Mutex
}

// NewREPLIntegration creates a new REPL integration
func NewREPLIntegration(shell *ishell.Shell) (*REPLIntegration, error) {
	ui, err := NewUI(shell)
	if err != nil {
		return nil, fmt.Errorf("failed to create UI: %w", err)
	}

	return &REPLIntegration{
		ui:    ui,
		shell: shell,
		mutex: sync.Mutex{},
	}, nil
}

// Start starts the REPL integration
func (r *REPLIntegration) Start() error {
	// Note: ishell doesn't provide direct methods to set input/output
	// We'll use a different approach to capture input/output

	// Create and set the input handler
	inputHandler := NewInputHandler(r.ui, r)
	r.ui.SetInputHandler(inputHandler)

	// Add system history entry
	r.ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "system",
		Content:   "Task started",
	})

	// Set up command processing
	r.setupCommandProcessing()

	// Start the UI in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := r.ui.Run(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for the UI to exit
	err := <-errCh
	if err != nil {
		slog.Error("UI error", "error", err)
		return err
	}

	return nil
}

// setupCommandProcessing sets up command processing
func (r *REPLIntegration) setupCommandProcessing() {
	// Since we can't directly access the shell's commands,
	// we'll use a different approach to capture command execution.

	// We'll register a custom process function for each command we know about
	// from the REPLCommands list in subcmd/repl.go

	// For now, we'll just add a message to the history
	r.AddSystemMessage("Command processing set up")
	r.AddSystemMessage("Type 'help' to see available commands")
}

// Close closes the REPL integration
func (r *REPLIntegration) Close() {
	r.ui.Close()
}

// AddUserInput adds user input to the history
func (r *REPLIntegration) AddUserInput(input string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "user",
		Content:   input,
	})
}

// AddAgentOutput adds agent output to the history
func (r *REPLIntegration) AddAgentOutput(output string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "agent",
		Content:   output,
	})
}

// AddSystemMessage adds a system message to the history
func (r *REPLIntegration) AddSystemMessage(message string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "system",
		Content:   message,
	})
}

// UpdateREPLInput updates the REPL input
func (r *REPLIntegration) UpdateREPLInput(input string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.ui.UpdateREPLInput(input)
}

// REPLReader is a custom reader for the REPL
type REPLReader struct {
	integration *REPLIntegration
	buffer      string
}

// Read implements the io.Reader interface
func (r *REPLReader) Read(p []byte) (n int, err error) {
	// This is a placeholder implementation
	// In a real implementation, this would read from the TUI input
	return 0, io.EOF
}

// REPLWriter is a custom writer for the REPL
type REPLWriter struct {
	integration *REPLIntegration
}

// Write implements the io.Writer interface
func (w *REPLWriter) Write(p []byte) (n int, err error) {
	// Convert the bytes to a string
	output := string(p)

	// Skip empty output
	if strings.TrimSpace(output) == "" {
		return len(p), nil
	}

	// Add the output to the history
	w.integration.AddSystemMessage(output)

	return len(p), nil
}

// StartREPLWithTUI starts the REPL with the TUI
func StartREPLWithTUI(shell *ishell.Shell) error {
	integration, err := NewREPLIntegration(shell)
	if err != nil {
		return fmt.Errorf("failed to create REPL integration: %w", err)
	}
	defer integration.Close()

	return integration.Start()
}
