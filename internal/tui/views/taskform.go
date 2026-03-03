package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tenlisboa/tempo/internal/store"
	"github.com/tenlisboa/tempo/internal/tui/components"
	"github.com/tenlisboa/tempo/internal/tui/shared"
)

type fieldIndex int

const (
	fieldName fieldIndex = iota
	fieldPrompt
	fieldScheduleType
	fieldScheduleExpr
	fieldWorkDir
	fieldSkipPerms
	fieldEnabled
	fieldCount
)

type TaskFormView struct {
	task            *store.Task
	focused         fieldIndex
	name            textinput.Model
	prompt          textarea.Model
	scheduleType    int
	scheduleExpr    textinput.Model
	workDir         textinput.Model
	skipPerms       bool
	enabled         bool
	showHelp        bool
	width           int
	height          int
	err             string
	pathSuggestions []string
	suggestionIdx   int
}

var scheduleTypes = []store.ScheduleType{
	store.ScheduleInterval,
	store.ScheduleDaily,
	store.ScheduleWeekly,
	store.ScheduleCron,
	store.ScheduleOnce,
}

func NewTaskForm(task *store.Task, width, height int, initialWorkDir string) TaskFormView {
	name := textinput.New()
	name.Placeholder = "Task name"
	name.Focus()

	prompt := textarea.New()
	prompt.Placeholder = "Claude prompt..."
	prompt.SetWidth(width - 6)
	prompt.SetHeight(5)

	exprInput := textinput.New()
	exprInput.Placeholder = "e.g. 30m, 09:00, mon:09:00, * * * * *"

	workDir := textinput.New()
	workDir.Placeholder = "/path/to/workdir (optional)"

	v := TaskFormView{
		name:         name,
		prompt:       prompt,
		scheduleExpr: exprInput,
		workDir:      workDir,
		skipPerms:    true,
		enabled:      true,
		width:        width,
		height:       height,
	}

	if task != nil {
		v.task = task
		v.name.SetValue(task.Name)
		v.prompt.SetValue(task.Prompt)
		v.scheduleExpr.SetValue(task.ScheduleExpr)
		v.workDir.SetValue(task.WorkDir)
		v.skipPerms = task.SkipPermissions
		v.enabled = task.Enabled
		for i, st := range scheduleTypes {
			if st == task.ScheduleType {
				v.scheduleType = i
				break
			}
		}
	} else if initialWorkDir != "" {
		v.workDir.SetValue(initialWorkDir)
	}
	return v
}

func (v TaskFormView) Update(msg tea.KeyMsg) (TaskFormView, tea.Cmd) {
	if v.showHelp {
		if msg.String() == "esc" || msg.String() == "?" {
			v.showHelp = false
		}
		return v, nil
	}

	if v.focused == fieldWorkDir && len(v.pathSuggestions) > 0 {
		switch msg.String() {
		case "down", "ctrl+n":
			v.suggestionIdx = (v.suggestionIdx + 1) % len(v.pathSuggestions)
			return v, nil
		case "up", "ctrl+p":
			v.suggestionIdx = (v.suggestionIdx - 1 + len(v.pathSuggestions)) % len(v.pathSuggestions)
			return v, nil
		case "tab":
			v.workDir.SetValue(v.pathSuggestions[v.suggestionIdx] + "/")
			v.workDir.CursorEnd()
			v.pathSuggestions = getPathSuggestions(v.workDir.Value())
			v.suggestionIdx = 0
			return v, nil
		}
	}

	switch {
	case key.Matches(msg, shared.FormKeys.Help):
		v.showHelp = true
		return v, nil

	case key.Matches(msg, shared.FormKeys.Cancel):
		return v, func() tea.Msg { return FormCancelMsg{} }

	case key.Matches(msg, shared.FormKeys.Save):
		return v.save()

	case key.Matches(msg, shared.FormKeys.Next):
		v = v.blur()
		v.focused = (v.focused + 1) % fieldCount
		v.pathSuggestions = nil
		v = v.focus()
		return v, nil

	case key.Matches(msg, shared.FormKeys.Prev):
		v = v.blur()
		v.focused = (v.focused - 1 + fieldCount) % fieldCount
		v.pathSuggestions = nil
		v = v.focus()
		return v, nil
	}

	switch v.focused {
	case fieldName:
		var cmd tea.Cmd
		v.name, cmd = v.name.Update(msg)
		return v, cmd
	case fieldPrompt:
		var cmd tea.Cmd
		v.prompt, cmd = v.prompt.Update(msg)
		return v, cmd
	case fieldScheduleType:
		switch msg.String() {
		case "left", "h":
			if v.scheduleType > 0 {
				v.scheduleType--
			}
		case "right", "l":
			if v.scheduleType < len(scheduleTypes)-1 {
				v.scheduleType++
			}
		}
	case fieldScheduleExpr:
		var cmd tea.Cmd
		v.scheduleExpr, cmd = v.scheduleExpr.Update(msg)
		return v, cmd
	case fieldWorkDir:
		var cmd tea.Cmd
		v.workDir, cmd = v.workDir.Update(msg)
		v.pathSuggestions = getPathSuggestions(v.workDir.Value())
		if v.suggestionIdx >= len(v.pathSuggestions) {
			v.suggestionIdx = 0
		}
		return v, cmd
	case fieldSkipPerms:
		if msg.String() == " " || msg.String() == "enter" {
			v.skipPerms = !v.skipPerms
		}
	case fieldEnabled:
		if msg.String() == " " || msg.String() == "enter" {
			v.enabled = !v.enabled
		}
	}
	return v, nil
}

func (v TaskFormView) blur() TaskFormView {
	switch v.focused {
	case fieldName:
		v.name.Blur()
	case fieldPrompt:
		v.prompt.Blur()
	case fieldScheduleExpr:
		v.scheduleExpr.Blur()
	case fieldWorkDir:
		v.workDir.Blur()
	}
	return v
}

func (v TaskFormView) focus() TaskFormView {
	switch v.focused {
	case fieldName:
		v.name.Focus()
	case fieldPrompt:
		v.prompt.Focus()
	case fieldScheduleExpr:
		v.scheduleExpr.Focus()
	case fieldWorkDir:
		v.workDir.Focus()
	}
	return v
}

func (v TaskFormView) save() (TaskFormView, tea.Cmd) {
	name := strings.TrimSpace(v.name.Value())
	prompt := strings.TrimSpace(v.prompt.Value())
	expr := strings.TrimSpace(v.scheduleExpr.Value())

	if name == "" {
		v.err = "name is required"
		return v, nil
	}
	if prompt == "" {
		v.err = "prompt is required"
		return v, nil
	}
	if expr == "" {
		v.err = "schedule expression is required"
		return v, nil
	}

	var id string
	if v.task != nil {
		id = v.task.ID
	}

	t := &store.Task{
		ID:              id,
		Name:            name,
		Prompt:          prompt,
		ScheduleType:    scheduleTypes[v.scheduleType],
		ScheduleExpr:    expr,
		WorkDir:         strings.TrimSpace(v.workDir.Value()),
		SkipPermissions: v.skipPerms,
		Enabled:         v.enabled,
	}
	v.err = ""
	return v, func() tea.Msg { return FormSaveMsg{Task: t} }
}

func (v TaskFormView) View() string {
	if v.showHelp {
		return components.NewHelpModal().View(v.width, v.height)
	}

	var sb strings.Builder

	title := "New Task"
	if v.task != nil {
		title = fmt.Sprintf("Edit: %s", v.task.Name)
	}
	sb.WriteString(shared.StyleTitle.Render(title))
	sb.WriteString("\n\n")

	sb.WriteString(fieldRow("Name", v.name.View(), v.focused == fieldName))
	sb.WriteString("\n\n")
	sb.WriteString(fieldRow("Prompt", v.prompt.View(), v.focused == fieldPrompt))
	sb.WriteString("\n\n")
	sb.WriteString(fieldRow("Schedule Type", v.scheduleTypeSelector(), v.focused == fieldScheduleType))
	sb.WriteString("\n\n")
	sb.WriteString(fieldRow("Schedule Expr", v.scheduleExpr.View(), v.focused == fieldScheduleExpr))
	sb.WriteString("\n\n")
	sb.WriteString(fieldRow("Work Dir", v.workDir.View(), v.focused == fieldWorkDir))
	if v.focused == fieldWorkDir && len(v.pathSuggestions) > 0 {
		sb.WriteString("\n")
		sb.WriteString(v.renderSuggestions())
	}
	sb.WriteString("\n\n")
	sb.WriteString(fieldRow("Skip Permissions", checkbox(v.skipPerms), v.focused == fieldSkipPerms))
	sb.WriteString("\n\n")
	sb.WriteString(fieldRow("Enabled", checkbox(v.enabled), v.focused == fieldEnabled))
	sb.WriteString("\n\n")

	if v.err != "" {
		sb.WriteString(shared.StyleBadgeError.Render("Error: " + v.err))
		sb.WriteString("\n\n")
	}

	sb.WriteString(shared.StyleHelp.Render("tab/shift+tab navigate  ctrl+s save  esc cancel  ? schedule help"))
	return sb.String()
}

func fieldRow(label, input string, focused bool) string {
	labelStyle := lipgloss.NewStyle().Width(18)
	if focused {
		labelStyle = labelStyle.Foreground(shared.ColorPrimary).Bold(true)
	} else {
		labelStyle = labelStyle.Foreground(shared.ColorMuted)
	}
	return fmt.Sprintf("%s %s", labelStyle.Render(label+":"), input)
}

func (v TaskFormView) scheduleTypeSelector() string {
	var parts []string
	for i, st := range scheduleTypes {
		s := string(st)
		if i == v.scheduleType {
			parts = append(parts, shared.StyleBadgeEnabled.Render("["+s+"]"))
		} else {
			parts = append(parts, shared.StyleSubtle.Render(" "+s+" "))
		}
	}
	return strings.Join(parts, " ")
}

func checkbox(v bool) string {
	if v {
		return shared.StyleBadgeEnabled.Render("[x]")
	}
	return shared.StyleSubtle.Render("[ ]")
}

func getPathSuggestions(partial string) []string {
	if partial == "" {
		return nil
	}
	dir, prefix := filepath.Split(partial)
	if dir == "" {
		dir = "."
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var suggestions []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if prefix == "" || strings.HasPrefix(e.Name(), prefix) {
			suggestions = append(suggestions, filepath.Join(dir, e.Name()))
		}
	}
	return suggestions
}

func (v TaskFormView) renderSuggestions() string {
	const maxShow = 5
	start := 0
	if v.suggestionIdx >= maxShow {
		start = v.suggestionIdx - maxShow + 1
	}
	end := min(start+maxShow, len(v.pathSuggestions))
	indent := strings.Repeat(" ", 19)
	var lines []string
	for i := start; i < end; i++ {
		s := v.pathSuggestions[i]
		if i == v.suggestionIdx {
			lines = append(lines, indent+shared.StyleBadgeEnabled.Render("▸ "+s))
		} else {
			lines = append(lines, indent+shared.StyleSubtle.Render("  "+s))
		}
	}
	return strings.Join(lines, "\n")
}

type FormSaveMsg struct{ Task *store.Task }
type FormCancelMsg struct{}
