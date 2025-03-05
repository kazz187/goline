package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/kazz187/goline/cmd/goline/subcmd"
)

var (
	// Create a new application
	app = kingpin.New("goline", "CUI-based AI agent inspired by Cline")

	// Set application details
	_ = app.Version("0.1.0")
	_ = app.Author("kazz187")
	_ = app.UsageWriter(os.Stdout)
	_ = app.HelpFlag.Short('h')

	// REPL commands
	startCmd = app.Command("start", "Start a new Goline task")
	_        = startCmd.Help("Start a new Goline task with an AI agent. This will open a TUI interface where you can interact with the AI agent.")

	resumeCmd = app.Command("resume", "Resume a paused task")
	_         = resumeCmd.Help("Resume a previously paused task. This will reopen the TUI interface for the specified task.")
	taskID    = resumeCmd.Arg("taskID", "ID of the task to resume").String()
	_         = taskID

	// Oneshot commands
	tasksCmd = app.Command("tasks", "List all tasks")
	_        = tasksCmd.Help("List all tasks, including active, paused, and completed tasks. Shows task ID, prompt, and status.")

	attachCmd  = app.Command("attach", "Attach to a terminal")
	_          = attachCmd.Help("Attach to a terminal that was started by a task. This allows you to interact with the terminal directly.")
	terminalID = attachCmd.Arg("terminalID", "ID of the terminal to attach to").Required().String()
	_          = terminalID

	// Help command is automatically provided by kingpin
)

func main() {
	// Set up logging
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Parse command line
	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Execute the appropriate command
	switch cmd {
	case "start":
		if err := subcmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "resume":
		if err := subcmd.Resume(*taskID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "tasks":
		if err := subcmd.ListTasks(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "attach":
		if err := subcmd.Attach(*terminalID); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}
