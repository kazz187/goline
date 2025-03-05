package subcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell/v2"
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
	checkpoint := &ishell.Cmd{
		Name: "checkpoint",
		Help: "Manage task checkpoints",
	}

	checkpoint.AddCmd(&ishell.Cmd{
		Name: "save",
		Help: "Save the current task state as a checkpoint",
		Func: func(c *ishell.Context) {
			c.Println("Saving checkpoint...")
			c.Println("TODO: Implement checkpoint save logic")
			c.Println("Checkpoint ID: checkpoint-123")
		},
	})

	checkpoint.AddCmd(&ishell.Cmd{
		Name: "restore",
		Help: "Restore a previously saved checkpoint",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Error: checkpoint ID is required")
				return
			}
			checkpointID := c.Args[0]
			c.Printf("Restoring checkpoint %s...\n", checkpointID)
			c.Println("TODO: Implement checkpoint restore logic")
		},
	})

	shell.AddCmd(checkpoint)
}

// registerDiffCommand registers the diff command
func registerDiffCommand(shell *ishell.Shell) {
	shell.AddCmd(&ishell.Cmd{
		Name: "diff",
		Help: "Show the difference between the current state and a checkpoint",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Error: checkpoint ID is required")
				return
			}
			checkpointID := c.Args[0]
			c.Printf("Showing diff for checkpoint %s...\n", checkpointID)
			c.Println("TODO: Implement diff logic")
		},
	})
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
