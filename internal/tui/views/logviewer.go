package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tenlisboa/tempo/internal/store"
	"github.com/tenlisboa/tempo/internal/tui/shared"
)

type LogViewerView struct {
	task     *store.Task
	logs     []*store.RunLog
	cursor   int
	viewport viewport.Model
	width    int
	height   int
}

func NewLogViewer(task *store.Task, logs []*store.RunLog, width, height int) LogViewerView {
	vp := viewport.New(width/2-2, height-8)
	v := LogViewerView{
		task:     task,
		logs:     logs,
		viewport: vp,
		width:    width,
		height:   height,
	}
	v.updateViewport()
	return v
}

func (v LogViewerView) WithLogs(logs []*store.RunLog) LogViewerView {
	v.logs = logs
	if v.cursor >= len(logs) && len(logs) > 0 {
		v.cursor = len(logs) - 1
	}
	v.updateViewport()
	return v
}

func (v *LogViewerView) updateViewport() {
	if len(v.logs) == 0 || v.cursor >= len(v.logs) {
		v.viewport.SetContent(shared.StyleSubtle.Render("no output"))
		return
	}
	v.viewport.SetContent(v.logs[v.cursor].Output)
}

func (v LogViewerView) Update(msg tea.KeyMsg) (LogViewerView, tea.Cmd) {
	switch {
	case key.Matches(msg, shared.LogKeys.Back):
		return v, func() tea.Msg { return LogBackMsg{} }
	case key.Matches(msg, shared.LogKeys.Up):
		if v.cursor > 0 {
			v.cursor--
			v.updateViewport()
		}
	case key.Matches(msg, shared.LogKeys.Down):
		if v.cursor < len(v.logs)-1 {
			v.cursor++
			v.updateViewport()
		}
	case key.Matches(msg, shared.LogKeys.PgUp):
		v.viewport.HalfViewUp()
	case key.Matches(msg, shared.LogKeys.PgDown):
		v.viewport.HalfViewDown()
	default:
		var cmd tea.Cmd
		v.viewport, cmd = v.viewport.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v LogViewerView) View() string {
	listWidth := v.width/2 - 1
	outputWidth := v.width - listWidth - 3

	title := shared.StyleTitle.Render(fmt.Sprintf("Logs: %s", v.task.Name))

	var list strings.Builder
	for i, r := range v.logs {
		var badge string
		if r.ExitCode != nil && *r.ExitCode == 0 {
			badge = shared.StyleBadgeSuccess.Render("✓")
		} else if r.ExitCode != nil {
			badge = shared.StyleBadgeError.Render("✗")
		} else {
			badge = shared.StyleSubtle.Render("…")
		}

		triggered := shared.StyleSubtle.Render(r.Triggered)
		ts := shared.StyleSubtle.Render(r.StartedAt.Format("01/02 15:04:05"))
		row := fmt.Sprintf("%s %s %s", badge, ts, triggered)
		if i == v.cursor {
			row = shared.StyleSelected.Width(listWidth - 2).Render(row)
		}
		list.WriteString(row)
		list.WriteString("\n")
	}

	if len(v.logs) == 0 {
		list.WriteString(shared.StyleSubtle.Render("no runs yet"))
	}

	v.viewport.Width = outputWidth
	v.viewport.Height = v.height - 8

	leftPane := lipgloss.NewStyle().
		Width(listWidth).
		Height(v.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.ColorBorder).
		Render(list.String())

	rightPane := lipgloss.NewStyle().
		Width(outputWidth).
		Height(v.height - 8).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.ColorBorder).
		Render(v.viewport.View())

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	help := shared.StyleHelp.Render("j/k select run  pgup/pgdn scroll output  esc back")

	return title + "\n\n" + body + "\n" + help
}

type LogBackMsg struct{}
