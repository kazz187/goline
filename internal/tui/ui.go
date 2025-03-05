package tui

import (
	"fmt"
	"log/slog"
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
	inputHandler  *InputHandler
}

// NewUI creates a new TUI
func NewUI(shell *ishell.Shell) (*UI, error) {
	if err := ui.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize termui: %w", err)
	}

	// Create task info widget
	taskInfo := widgets.NewParagraph()
	taskInfo.Title = "Task Information"
	taskInfo.BorderStyle.Fg = ui.ColorYellow

	// Create history widget
	historyList := widgets.NewList()
	historyList.Title = "Task History"
	historyList.BorderStyle.Fg = ui.ColorCyan
	historyList.TextStyle = ui.NewStyle(ui.ColorWhite)
	historyList.WrapText = true

	// Create REPL widget
	replParagraph := widgets.NewParagraph()
	replParagraph.Title = "Command Input"
	replParagraph.BorderStyle.Fg = ui.ColorGreen

	// Create grid layout
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	// Set up the grid with three rows
	// Task info: 15% of height
	// History: 70% of height
	// REPL: 15% of height
	grid.Set(
		ui.NewRow(0.15,
			ui.NewCol(1.0, taskInfo),
		),
		ui.NewRow(0.70,
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
	u.renderTaskInfo()
}

// AddHistoryEntry adds an entry to the history widget
func (u *UI) AddHistoryEntry(entry HistoryEntry) {
	u.historyData = append(u.historyData, entry)
	u.renderHistory()
}

// UpdateREPLInput updates the REPL input widget
func (u *UI) UpdateREPLInput(input string) {
	u.replInput = input
	u.renderREPL()
}

// renderTaskInfo renders the task info widget
func (u *UI) renderTaskInfo() {
	elapsed := time.Since(u.taskInfoData.StartTime).Round(time.Second)
	text := fmt.Sprintf("ID: %s | Status: %s | Elapsed: %s | Provider: %s | Engine: %s",
		u.taskInfoData.ID,
		u.taskInfoData.Status,
		elapsed,
		u.taskInfoData.Provider,
		u.taskInfoData.Engine,
	)
	u.taskInfo.Text = text
	ui.Render(u.taskInfo)
}

// renderHistory renders the history widget
func (u *UI) renderHistory() {
	u.historyList.Rows = []string{}
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
		wrappedText := wrapText(text, u.historyList.Inner.Dx())
		u.historyList.Rows = append(u.historyList.Rows, wrappedText...)
	}
	ui.Render(u.historyList)
}

// renderREPL renders the REPL widget
func (u *UI) renderREPL() {
	prompt := "goline> "
	u.replParagraph.Text = prompt + u.replInput
	ui.Render(u.replParagraph)
}

// Render renders the UI
func (u *UI) Render() {
	u.renderTaskInfo()
	u.renderHistory()
	u.renderREPL()
	ui.Render(u.grid)
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
			// Update elapsed time every second
			u.renderTaskInfo()
		}
	}
}

// SetInputHandler sets the input handler for the UI
func (u *UI) SetInputHandler(handler *InputHandler) {
	// This will be called by the REPLIntegration
	u.inputHandler = handler
}

// wrapText wraps text to fit within a given width
func wrapText(text string, width int) []string {
	var lines []string
	words := runewidth.StringWidth(text)
	if words <= width {
		return []string{text}
	}

	// Simple wrapping algorithm
	current := ""
	currentWidth := 0
	for _, char := range text {
		charWidth := runewidth.RuneWidth(char)
		if currentWidth+charWidth > width {
			lines = append(lines, current)
			current = string(char)
			currentWidth = charWidth
		} else {
			current += string(char)
			currentWidth += charWidth
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// StartTUI starts the TUI
func StartTUI(shell *ishell.Shell) error {
	ui, err := NewUI(shell)
	if err != nil {
		return fmt.Errorf("failed to create UI: %w", err)
	}
	defer ui.Close()

	// Add some sample history entries
	ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "system",
		Content:   "Task started",
	})
	ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "user",
		Content:   "Hello, I need help with implementing a feature",
	})
	ui.AddHistoryEntry(HistoryEntry{
		Timestamp: time.Now(),
		Type:      "agent",
		Content:   "I'll help you implement that feature. What are the requirements?",
	})

	// Run the UI
	if err := ui.Run(); err != nil {
		slog.Error("UI error", "error", err)
		return err
	}

	return nil
}
