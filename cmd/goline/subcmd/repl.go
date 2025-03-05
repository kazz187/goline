package subcmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abiosoft/ishell/v2"
	"github.com/kazz187/goline/internal/core/checkpoint"
)

// REPLCommands defines the available commands in the REPL
var REPLCommands = []struct {
	Name        string
	Description string
	Usage       string
}{
	{
		Name:        "help",
		Description: "Display help for REPL commands",
		Usage:       "help",
	},
	{
		Name:        "exit",
		Description: "Exit the REPL",
		Usage:       "exit",
	},
	{
		Name:        "ask",
		Description: "Ask the AI agent a question",
		Usage:       "ask [question]",
	},
	{
		Name:        "apply",
		Description: "Apply the AI agent's suggestion",
		Usage:       "apply",
	},
	{
		Name:        "cancel",
		Description: "Cancel the AI agent's suggestion",
		Usage:       "cancel",
	},
	{
		Name:        "checkpoint save",
		Description: "Save the current task state as a checkpoint",
		Usage:       "checkpoint save",
	},
	{
		Name:        "checkpoint restore",
		Description: "Restore a previously saved checkpoint",
		Usage:       "checkpoint restore [checkpointID]",
	},
	{
		Name:        "diff",
		Description: "Show the difference between the current state and a checkpoint",
		Usage:       "diff [checkpointID]",
	},
}

// initREPL initializes the REPL shell
func initREPL() *ishell.Shell {
	shell := ishell.New()
	shell.SetPrompt("goline> ")

	// Register commands
	registerHelpCommand(shell)
	registerExitCommand(shell)
	registerAskCommand(shell)
	registerApplyCommand(shell)
	registerCancelCommand(shell)
	registerCheckpointCommands(shell)
	registerDiffCommand(shell)

	return shell
}

// registerHelpCommand registers the help command
func registerHelpCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "help",
		Help: "Display help for REPL commands",
		Func: func(c *ishell.Context) {
			c.Println("Goline REPL Commands:")
			c.Println()

			maxNameLen := 0
			maxUsageLen := 0
			for _, cmd := range REPLCommands {
				if len(cmd.Name) > maxNameLen {
					maxNameLen = len(cmd.Name)
				}
				if len(cmd.Usage) > maxUsageLen {
					maxUsageLen = len(cmd.Usage)
				}
			}

			for _, cmd := range REPLCommands {
				name := cmd.Name + strings.Repeat(" ", maxNameLen-len(cmd.Name))
				usage := cmd.Usage + strings.Repeat(" ", maxUsageLen-len(cmd.Usage))
				c.Printf("  %s  %s  %s\n", name, usage, cmd.Description)
			}
			c.Println()
		},
	})
}

// registerExitCommand registers the exit command
func registerExitCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "exit",
		Help: "Exit the REPL",
		Func: func(c *ishell.Context) {
			c.Println("Exiting Goline...")
			os.Exit(0)
		},
	})
}

// registerAskCommand registers the ask command
func registerAskCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "ask",
		Help: "Ask the AI agent a question",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				// If no arguments, open an editor for multi-line input
				c.Println("Enter your question (press Ctrl+D when done):")
				question := c.ReadMultiLines(">")
				c.Printf("Question: %s\n", question)
				c.Println("TODO: Send question to AI agent")
			} else {
				// If arguments are provided, use them as the question
				question := strings.Join(c.Args, " ")
				c.Printf("Question: %s\n", question)
				c.Println("TODO: Send question to AI agent")
			}
		},
	})
}

// registerApplyCommand registers the apply command
func registerApplyCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "apply",
		Help: "Apply the AI agent's suggestion",
		Func: func(c *ishell.Context) {
			c.Println("Applying AI agent's suggestion...")
			c.Println("TODO: Implement apply logic")
		},
	})
}

// registerCancelCommand registers the cancel command
func registerCancelCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "cancel",
		Help: "Cancel the AI agent's suggestion",
		Func: func(c *ishell.Context) {
			c.Println("Cancelling AI agent's suggestion...")
			c.Println("TODO: Implement cancel logic")
		},
	})
}

// registerCheckpointCommands registers the checkpoint commands
func registerCheckpointCommands(shell *ishell.Shell) {
	checkpointCmd := &ishell.Cmd{
		Name: "checkpoint",
		Help: "Manage task checkpoints",
	}

	checkpointCmd.AddCmd(&ishell.Cmd{
		Name: "save",
		Help: "Save the current task state as a checkpoint",
		Func: func(c *ishell.Context) {
			// Get task context
			taskID := getCurrentTaskID()
			if taskID == "" {
				c.Println("Error: No active task")
				return
			}
			workingDir, err := os.Getwd()
			if err != nil {
				c.Printf("Error: Failed to get working directory: %v\n", err)
				return
			}

			// Get checkpoint name
			var name string
			if len(c.Args) > 0 {
				name = strings.Join(c.Args, " ")
			} else {
				c.Print("Enter checkpoint name: ")
				name = c.ReadLine()
			}
			if name == "" {
				name = fmt.Sprintf("Checkpoint %s", time.Now().Format(time.RFC3339))
			}

			// Create checkpoint service
			service := checkpoint.NewService()

			// Save checkpoint
			c.Println("Saving checkpoint...")
			event, err := service.SaveCheckpoint(taskID, workingDir, name, "")
			if err != nil {
				c.Printf("Error: Failed to save checkpoint: %v\n", err)
				return
			}

			// Display result
			c.Printf("Checkpoint saved: %s\n", event.CheckpointId[:8])

			// TODO: Add checkpoint event to task history
		},
	})

	checkpointCmd.AddCmd(&ishell.Cmd{
		Name: "restore",
		Help: "Restore a previously saved checkpoint",
		Func: func(c *ishell.Context) {
			// Get task context
			taskID := getCurrentTaskID()
			if taskID == "" {
				c.Println("Error: No active task")
				return
			}
			workingDir, err := os.Getwd()
			if err != nil {
				c.Printf("Error: Failed to get working directory: %v\n", err)
				return
			}

			// Create checkpoint service
			service := checkpoint.NewService()

			// Get checkpoints
			checkpoints, err := service.GetCheckpoints(taskID, workingDir)
			if err != nil {
				c.Printf("Error: Failed to get checkpoints: %v\n", err)
				return
			}
			if len(checkpoints) == 0 {
				c.Println("No checkpoints available")
				return
			}

			// Get checkpoint ID
			var checkpointID string
			if len(c.Args) > 0 {
				checkpointID = c.Args[0]
			} else {
				// Display checkpoints
				c.Println(service.FormatCheckpointList(checkpoints))

				// Prompt for checkpoint ID
				c.Print("Enter checkpoint ID: ")
				checkpointID = c.ReadLine()
			}
			if checkpointID == "" {
				c.Println("Error: checkpoint ID is required")
				return
			}

			// Confirm restore
			c.Printf("Are you sure you want to restore checkpoint %s? This will overwrite your current workspace. (y/n): ", checkpointID)
			confirm := c.ReadLine()
			if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
				c.Println("Restore cancelled")
				return
			}

			// Restore checkpoint
			c.Printf("Restoring checkpoint %s...\n", checkpointID)
			_, err = service.RestoreCheckpoint(taskID, workingDir, checkpointID)
			if err != nil {
				c.Printf("Error: Failed to restore checkpoint: %v\n", err)
				return
			}

			c.Println("Checkpoint restored successfully")

			// TODO: Add checkpoint event to task history
		},
	})

	checkpointCmd.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "List all checkpoints for the current task",
		Func: func(c *ishell.Context) {
			// Get task context
			taskID := getCurrentTaskID()
			if taskID == "" {
				c.Println("Error: No active task")
				return
			}
			workingDir, err := os.Getwd()
			if err != nil {
				c.Printf("Error: Failed to get working directory: %v\n", err)
				return
			}

			// Create checkpoint service
			service := checkpoint.NewService()

			// Get checkpoints
			checkpoints, err := service.GetCheckpoints(taskID, workingDir)
			if err != nil {
				c.Printf("Error: Failed to get checkpoints: %v\n", err)
				return
			}

			// Display checkpoints
			c.Println(service.FormatCheckpointList(checkpoints))
		},
	})

	shell.AddCmd(checkpointCmd)
}

// registerDiffCommand registers the diff command
func registerDiffCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "diff",
		Help: "Show the difference between the current state and a checkpoint",
		Func: func(c *ishell.Context) {
			// Get task context
			taskID := getCurrentTaskID()
			if taskID == "" {
				c.Println("Error: No active task")
				return
			}
			workingDir, err := os.Getwd()
			if err != nil {
				c.Printf("Error: Failed to get working directory: %v\n", err)
				return
			}

			// Create checkpoint service
			service := checkpoint.NewService()

			// Get checkpoints
			checkpoints, err := service.GetCheckpoints(taskID, workingDir)
			if err != nil {
				c.Printf("Error: Failed to get checkpoints: %v\n", err)
				return
			}
			if len(checkpoints) == 0 {
				c.Println("No checkpoints available")
				return
			}

			// Get checkpoint ID
			var fromCheckpointID string
			if len(c.Args) > 0 {
				fromCheckpointID = c.Args[0]
			} else {
				// Display checkpoints
				c.Println(service.FormatCheckpointList(checkpoints))

				// Prompt for checkpoint ID
				c.Print("Enter checkpoint ID to compare from: ")
				fromCheckpointID = c.ReadLine()
			}
			if fromCheckpointID == "" {
				c.Println("Error: checkpoint ID is required")
				return
			}

			// Get diff
			c.Printf("Showing diff for checkpoint %s...\n", fromCheckpointID)
			diffs, err := service.GetDiff(taskID, workingDir, fromCheckpointID, "")
			if err != nil {
				c.Printf("Error: Failed to get diff: %v\n", err)
				return
			}

			// Display diff
			c.Println(service.FormatDiff(diffs))
		},
	})
}

// getCurrentTaskID returns the ID of the current task
// TODO: Implement this function to get the actual task ID
func getCurrentTaskID() string {
	// For now, return a dummy task ID
	return "task-123"
}

// startREPL starts the REPL shell
func startREPL() error {
	shell := initREPL()

	// Display welcome message
	fmt.Println("Welcome to Goline!")
	fmt.Println("Type 'help' to see available commands.")

	// Start the shell
	shell.Run()

	return nil
}
