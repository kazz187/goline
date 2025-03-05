package subcmd

import (
	"errors"
	"fmt"

	"github.com/kazz187/goline/internal/tui"
)

// Start starts a new Goline task
func Start() error {
	fmt.Println("Starting a new Goline task...")

	// Initialize the REPL shell
	shell := initREPL()

	// Start the TUI with the REPL
	return tui.StartREPLWithTUI(shell)
}

// Resume resumes a paused task
func Resume(taskID string) error {
	fmt.Printf("Resuming task %s...\n", taskID)

	// Initialize the REPL shell
	shell := initREPL()

	// TODO: Load task data from storage

	// Start the TUI with the REPL
	return tui.StartREPLWithTUI(shell)
}

// ListTasks lists all tasks
func ListTasks() error {
	fmt.Println("Listing all tasks...")
	// TODO: Implement task listing logic
	return errors.New("not implemented yet")
}

// Attach attaches to a terminal
func Attach(terminalID string) error {
	fmt.Printf("Attaching to terminal %s...\n", terminalID)
	// TODO: Implement terminal attachment logic
	return errors.New("not implemented yet")
}
