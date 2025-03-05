package subcmd

import (
	"errors"
	"fmt"
)

// Start starts a new Goline task
func Start() error {
	fmt.Println("Starting a new Goline task...")
	return startREPL()
}

// Resume resumes a paused task
func Resume(taskID string) error {
	fmt.Printf("Resuming task %s...\n", taskID)
	// TODO: Implement task resuming logic
	return errors.New("not implemented yet")
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
