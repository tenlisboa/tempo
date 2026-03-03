package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tenlisboa/tempo/internal/store"
	"github.com/tenlisboa/tempo/internal/tui/components"
	"github.com/tenlisboa/tempo/internal/tui/shared"
)

type TaskListView struct {
	tasks      []*store.Task
	cursor     int
	width      int
	height     int
	daemonOK   bool
	confirming *components.Confirm
}

func NewTaskList(tasks []*store.Task, width, height int, daemonOK bool) TaskListView {
	return TaskListView{tasks: tasks, width: width, height: height, daemonOK: daemonOK}
}

func (v TaskListView) WithTasks(tasks []*store.Task) TaskListView {
	v.tasks = tasks
	if v.cursor >= len(tasks) && len(tasks) > 0 {
		v.cursor = len(tasks) - 1
	}
	return v
}

func (v TaskListView) WithSize(w, h int) TaskListView {
	v.width = w
	v.height = h
	return v
}

func (v TaskListView) WithDaemon(ok bool) TaskListView {
	v.daemonOK = ok
	return v
}

func (v TaskListView) SelectedTask() *store.Task {
	if len(v.tasks) == 0 || v.cursor >= len(v.tasks) {
		return nil
	}
	return v.tasks[v.cursor]
}

func (v TaskListView) Update(msg tea.KeyMsg) (TaskListView, tea.Cmd) {
	if v.confirming != nil {
		updated, result := v.confirming.Update(msg)
		if result != nil {
			v.confirming = nil
			if result.Confirmed {
				return v, func() tea.Msg { return DeleteConfirmedMsg{TaskID: result.Tag} }
			}
		} else {
			v.confirming = &updated
		}
		return v, nil
	}

	switch {
	case msg.String() == "up", msg.String() == "k":
		if v.cursor > 0 {
			v.cursor--
		}
	case msg.String() == "down", msg.String() == "j":
		if v.cursor < len(v.tasks)-1 {
			v.cursor++
		}
	case key.Matches(msg, shared.ListKeys.New):
		return v, func() tea.Msg { return OpenFormMsg{} }
	case key.Matches(msg, shared.ListKeys.Edit):
		if t := v.SelectedTask(); t != nil {
			return v, func() tea.Msg { return OpenFormMsg{Task: t} }
		}
	case key.Matches(msg, shared.ListKeys.Delete):
		if t := v.SelectedTask(); t != nil {
			c := components.NewConfirm(fmt.Sprintf("Delete %q?", t.Name), t.ID)
			v.confirming = &c
		}
	case key.Matches(msg, shared.ListKeys.Run):
		if t := v.SelectedTask(); t != nil {
			return v, func() tea.Msg { return RunTaskMsg{TaskID: t.ID} }
		}
	case key.Matches(msg, shared.ListKeys.Logs):
		if t := v.SelectedTask(); t != nil {
			return v, func() tea.Msg { return OpenLogsMsg{Task: t} }
		}
	case key.Matches(msg, shared.ListKeys.Toggle):
		if t := v.SelectedTask(); t != nil {
			return v, func() tea.Msg { return ToggleTaskMsg{TaskID: t.ID} }
		}
	case key.Matches(msg, shared.ListKeys.Quit):
		return v, tea.Quit
	}
	return v, nil
}

func (v TaskListView) View() string {
	if v.confirming != nil {
		return v.confirming.View()
	}

	var sb strings.Builder

	taskWord := "tasks"
	if len(v.tasks) == 1 {
		taskWord = "task"
	}
	titleStr := shared.StyleTitle.Render("tempo")
	countStr := shared.StyleCount.Render(fmt.Sprintf("%d %s", len(v.tasks), taskWord))
	gap := v.width - lipgloss.Width(titleStr) - lipgloss.Width(countStr) - 2
	if gap < 0 {
		gap = 0
	}
	sb.WriteString(titleStr + strings.Repeat(" ", gap) + countStr)
	sb.WriteString("\n")
	sb.WriteString(shared.StyleDivider.Render(strings.Repeat("─", v.width)))
	sb.WriteString("\n\n")

	if len(v.tasks) == 0 {
		sb.WriteString(shared.StyleSubtle.Render("  No tasks yet. Press n to create one."))
		sb.WriteString("\n")
	} else {
		for i, t := range v.tasks {
			sb.WriteString(v.renderRow(i, t))
			sb.WriteString("\n")
		}
	}

	listHeight := v.height - 8
	rendered := sb.String()
	lines := strings.Count(rendered, "\n")
	for i := lines; i < listHeight; i++ {
		rendered += "\n"
	}

	helpText := "n new  e edit  d del  r run  l logs  t toggle  q quit"
	divider := shared.StyleDivider.Render(strings.Repeat("─", v.width))
	statusBar := components.StatusBar(v.width, v.daemonOK, helpText)

	return rendered + divider + "\n" + statusBar
}

func (v TaskListView) renderRow(idx int, t *store.Task) string {
	var enabledBadge string
	if t.Enabled {
		enabledBadge = shared.StyleBadgeEnabled.Render("●")
	} else {
		enabledBadge = shared.StyleBadgeDisabled.Render("○")
	}

	var exitBadge string
	if t.LastExitCode != nil {
		if *t.LastExitCode == 0 {
			exitBadge = shared.StyleBadgeSuccess.Render("✓")
		} else {
			exitBadge = shared.StyleBadgeError.Render("✗")
		}
	} else {
		exitBadge = shared.StyleSubtle.Render("·")
	}

	lastRun := shared.StyleSubtle.Render("never")
	if t.LastRunAt != nil {
		lastRun = shared.StyleSubtle.Render(formatRelative(*t.LastRunAt))
	}

	name := t.Name
	schedule := shared.StyleSubtle.Render(fmt.Sprintf("%s · %s", t.ScheduleType, t.ScheduleExpr))

	row := fmt.Sprintf("  %s %s  %s  %s  %s", enabledBadge, exitBadge, name, schedule, lastRun)

	if idx == v.cursor {
		row = shared.StyleSelected.Width(v.width - 2).Render(row)
	}

	return lipgloss.NewStyle().PaddingLeft(1).Render(row)
}

func formatRelative(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

type OpenFormMsg struct{ Task *store.Task }
type OpenLogsMsg struct{ Task *store.Task }
type RunTaskMsg struct{ TaskID string }
type ToggleTaskMsg struct{ TaskID string }
type DeleteConfirmedMsg struct{ TaskID string }
