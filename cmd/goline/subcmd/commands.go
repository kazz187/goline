package subcmd

import (
	"errors"
	"fmt"

	"github.com/kazz187/goline/internal/tui"
)

// Start starts a new Goline task
func Start() error {
	fmt.Println("Starting a new Goline task...")

	// Start the TUI with the REPL
	return tui.StartREPLWithTUI()
}

// Resume resumes a paused task
func Resume(taskID string) error {
	fmt.Printf("Resuming task %s...\n", taskID)

	// TODO: Load task data from storage

	// Start the TUI with the REPL
	return tui.StartREPLWithTUI()
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
