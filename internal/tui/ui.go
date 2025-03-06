package tui

import (
	"bytes"
	"fmt"
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

// InputHandlerInterface defines the interface for input handlers
type InputHandlerInterface interface {
	HandleKeyEvent(e ui.Event) bool
	GetCursorPosition() int
}

// UI represents the TUI
type UI struct {
	shell        *ishell.Shell
	shellInput   *bytes.Buffer
	replUI       *ReplUI
	inputHandler InputHandlerInterface
	termWidth    int
	termHeight   int
}

type ReplUI struct {
	taskInfo    *Block[*widgets.Paragraph, *TaskInfo]
	historyList *Block[*widgets.List, []HistoryEntry]
	repl        *Block[*widgets.Paragraph, string]
}

type Block[T ui.Drawable, S any] struct {
	Widget       T
	data         S
	updateSignal chan struct{}
}

func NewBlock[T ui.Drawable, S any](widget T, data S) *Block[T, S] {
	return &Block[T, S]{
		Widget:       widget,
		data:         data,
		updateSignal: make(chan struct{}, 1),
	}
}

func (b *Block[T, S]) SetData(data S) {
	b.data = data
	select {
	case b.updateSignal <- struct{}{}:
	default:
	}
}

func (b *Block[T, S]) GetData() S {
	return b.data
}

func (b *Block[T, S]) UpdateSignal() <-chan struct{} {
	return b.updateSignal
}

func (b *Block[T, S]) Render() {
	ui.Render(b.Widget)
}

func NewReplUI() *ReplUI {
	taskInfoData := &TaskInfo{
		ID:        "task-123",
		Status:    "Active",
		StartTime: time.Now(),
		Provider:  "anthropic",
		Engine:    "claude-3-opus",
	}

	taskInfo := widgets.NewParagraph()
	taskInfo.Title = "Task Information"
	taskInfo.BorderStyle.Fg = ui.ColorYellow
	taskInfo.PaddingTop = 0
	taskInfo.PaddingBottom = 0

	historyList := widgets.NewList()
	historyList.Title = "Task History"
	historyList.BorderStyle.Fg = ui.ColorCyan
	historyList.TextStyle = ui.NewStyle(ui.ColorWhite)
	historyList.WrapText = true

	repl := widgets.NewParagraph()
	repl.Title = "Command Input"
	repl.BorderStyle.Fg = ui.ColorGreen
	repl.Text = ""

	g := &ReplUI{
		taskInfo:    NewBlock(taskInfo, taskInfoData),
		historyList: NewBlock(historyList, []HistoryEntry{}),
		repl:        NewBlock(repl, ""),
	}
	return g
}

func (gu *ReplUI) Render(termWidth, termHeight int) {
	grid := ui.NewGrid()
	grid.SetRect(0, 0, termWidth, termHeight)

	taskInfoCol := ui.NewCol(1.0, gu.taskInfo.Widget)
	historyListCol := ui.NewCol(1.0, gu.historyList.Widget)
	replCol := ui.NewCol(1.0, gu.repl.Widget)

	taskInfoHeight := float64(3) / float64(termHeight)
	historyListHeight := 0.7
	replHeight := 1.0 - taskInfoHeight - historyListHeight

	taskInfoRow := ui.NewRow(taskInfoHeight, taskInfoCol)
	historyListRow := ui.NewRow(historyListHeight, historyListCol)
	replRow := ui.NewRow(replHeight, replCol)

	grid.Set(
		taskInfoRow,
		historyListRow,
		replRow,
	)
	ui.Render(grid)
}

// NewUI creates a new TUI instance.
func NewUI(shell *ishell.Shell, shellInput *bytes.Buffer) (*UI, error) {
	// ishell のデフォルトプロンプトを無効化する
	shell.SetPrompt("")

	if err := ui.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize termui: %w", err)
	}

	return &UI{
		shell:      shell,
		shellInput: shellInput,
		replUI:     NewReplUI(),
	}, nil
}

// UpdateTaskInfo updates the task info widget
func (u *UI) UpdateTaskInfo(taskInfo *TaskInfo) {
	u.replUI.taskInfo.SetData(taskInfo)
}

// AddHistoryEntry adds an entry to the history widget
func (u *UI) AddHistoryEntry(entry HistoryEntry) {
	u.replUI.historyList.SetData(append(u.replUI.historyList.GetData(), entry))
}

// UpdateREPLInput updates the REPL input widget
func (u *UI) UpdateREPLInput(input string) {
	u.replUI.repl.SetData(input)
}

// UpdateREPLPrompt updates the REPL prompt
func (u *UI) UpdateREPLPrompt(prompt string) {
	//u.replUI.repl.Data = prompt
}

// renderTaskInfo updates the task info widget content.
func (u *UI) prerenderTaskInfo() {
	taskInfo := u.replUI.taskInfo.GetData()
	elapsed := time.Since(taskInfo.StartTime).Round(time.Second)
	text := fmt.Sprintf("ID: %s | Status: %s | Elapsed: %s | Provider: %s | Engine: %s",
		taskInfo.ID,
		taskInfo.Status,
		elapsed,
		taskInfo.Provider,
		taskInfo.Engine,
	)
	availableWidth := u.replUI.taskInfo.Widget.Inner.Dx()
	if runewidth.StringWidth(text) > availableWidth {
		// 短縮表示
		truncated := ""
		currentWidth := 0
		for _, char := range text {
			charWidth := runewidth.RuneWidth(char)
			if currentWidth+charWidth+3 > availableWidth {
				break
			}
			truncated += string(char)
			currentWidth += charWidth
		}
		text = truncated + "..."
	}
	u.replUI.taskInfo.Widget.Text = text
}

// renderHistory updates the history widget content.
func (u *UI) prerenderHistory() {
	u.replUI.historyList.Widget.Rows = []string{}
	width := u.replUI.historyList.Widget.Inner.Dx()
	if width < 80 {
		width = 80
	}
	for _, entry := range u.replUI.historyList.GetData() {
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
		line := fmt.Sprintf("[%s] %s %s", timestamp, prefix, entry.Content)
		// termui 側で自動改行させるため、そのまま設定
		u.replUI.historyList.Widget.Rows = append(u.replUI.historyList.Widget.Rows, line)
	}
}

// renderREPL updates the REPL widget content.
// ※ishell のプロンプトを u.replInput に含めないようにし、ここで一度だけプロンプトを先頭に追加します。
func (u *UI) prerenderREPL() {
	in := u.shellInput.String()
	u.replUI.repl.Widget.Text = in
}

// adjustGridLayout adjusts the replUI layout based on terminal size and input lines.
func (u *UI) adjustGridLayout(termWidth, termHeight int) bool {
	if u.termWidth != termWidth || u.termHeight != termHeight {
		u.termWidth = termWidth
		u.termHeight = termHeight
		u.prerenderTaskInfo()
		u.prerenderHistory()
		u.prerenderREPL()
		u.replUI.Render(termWidth, termHeight)
		return true
	}
	return false
}

// Close closes the UI.
func (u *UI) Close() {
	ui.Close()
}

// Run runs the UI.
func (u *UI) Run() error {
	termWidth, termHeight := ui.TerminalDimensions()
	u.adjustGridLayout(termWidth, termHeight)
	uiEvents := ui.PollEvents()

	for {
		select {
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				if u.inputHandler != nil {
					if exit := u.inputHandler.HandleKeyEvent(e); exit {
						return nil
					}
				} else {
					switch e.ID {
					case "q", "<C-c>":
						return nil
					}
				}
			} else if e.Type == ui.ResizeEvent {
				time.Sleep(10 * time.Millisecond)
				//payload := e.Payload.(ui.Resize)
				termWidth, termHeight := ui.TerminalDimensions()
				u.adjustGridLayout(termWidth, termHeight)
			}
		case <-u.replUI.taskInfo.UpdateSignal():
			u.prerenderTaskInfo()
			u.replUI.taskInfo.Render()
		case <-u.replUI.historyList.UpdateSignal():
			u.prerenderHistory()
			u.replUI.historyList.Render()
		case <-u.replUI.repl.UpdateSignal():
			u.prerenderREPL()
			u.replUI.repl.Render()
		}
	}
}

// SetInputHandler sets the input handler for the UI.
func (u *UI) SetInputHandler(handler InputHandlerInterface) {
	u.inputHandler = handler
}
