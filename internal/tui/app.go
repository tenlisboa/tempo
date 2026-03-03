package tui

import (
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tenlisboa/tempo/internal/ipc"
	"github.com/tenlisboa/tempo/internal/store"
	"github.com/tenlisboa/tempo/internal/tui/views"
)

type viewState int

const (
	stateList viewState = iota
	stateForm
	stateLogs
)

type App struct {
	client         *ipc.Client
	state          viewState
	listView       views.TaskListView
	formView       views.TaskFormView
	logView        views.LogViewerView
	daemonOK       bool
	width          int
	height         int
	initialWorkDir string
}

func NewApp(client *ipc.Client, initialWorkDir string) *App {
	return &App{client: client, initialWorkDir: initialWorkDir}
}

type tickMsg struct{}
type tasksLoadedMsg struct{ tasks []*store.Task }
type logsLoadedMsg struct{ logs []*store.RunLog }
type daemonStatusMsg struct{ ok bool }
type errMsg struct{ err error }

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.cmdCheckDaemon(),
		a.cmdLoadTasks(),
		tick(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.listView = a.listView.WithSize(msg.Width, msg.Height)
		return a, nil

	case tickMsg:
		return a, tea.Batch(a.cmdCheckDaemon(), a.cmdLoadTasks(), tick())

	case daemonStatusMsg:
		a.daemonOK = msg.ok
		a.listView = a.listView.WithDaemon(msg.ok)
		return a, nil

	case tasksLoadedMsg:
		a.listView = a.listView.WithTasks(msg.tasks)
		return a, nil

	case logsLoadedMsg:
		if a.state == stateLogs {
			a.logView = a.logView.WithLogs(msg.logs)
		}
		return a, nil

	case views.OpenFormMsg:
		a.formView = views.NewTaskForm(msg.Task, a.width, a.height, a.initialWorkDir)
		a.state = stateForm
		return a, nil

	case views.FormSaveMsg:
		a.state = stateList
		return a, a.cmdSaveTask(msg.Task)

	case views.FormCancelMsg:
		a.state = stateList
		return a, nil

	case views.OpenLogsMsg:
		a.logView = views.NewLogViewer(msg.Task, nil, a.width, a.height)
		a.state = stateLogs
		return a, a.cmdLoadLogs(msg.Task.ID)

	case views.LogBackMsg:
		a.state = stateList
		return a, nil

	case views.RunTaskMsg:
		return a, a.cmdRunTask(msg.TaskID)

	case views.ToggleTaskMsg:
		return a, a.cmdToggleTask(msg.TaskID)

	case views.DeleteConfirmedMsg:
		return a, a.cmdDeleteTask(msg.TaskID)

	case tea.KeyMsg:
		return a.handleKey(msg)
	}
	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.state {
	case stateList:
		updated, cmd := a.listView.Update(msg)
		a.listView = updated
		return a, cmd
	case stateForm:
		updated, cmd := a.formView.Update(msg)
		a.formView = updated
		return a, cmd
	case stateLogs:
		updated, cmd := a.logView.Update(msg)
		a.logView = updated
		return a, cmd
	}
	return a, nil
}

func (a *App) View() string {
	switch a.state {
	case stateForm:
		return a.formView.View()
	case stateLogs:
		return a.logView.View()
	default:
		return a.listView.View()
	}
}

func (a *App) cmdCheckDaemon() tea.Cmd {
	return func() tea.Msg {
		resp, err := a.client.Call("daemon.ping", nil)
		if err != nil || resp.Error != "" {
			return daemonStatusMsg{ok: false}
		}
		return daemonStatusMsg{ok: true}
	}
}

func (a *App) cmdLoadTasks() tea.Cmd {
	return func() tea.Msg {
		resp, err := a.client.Call("task.list", nil)
		if err != nil || resp.Error != "" {
			return tasksLoadedMsg{}
		}
		b, _ := json.Marshal(resp.Data)
		var tasks []*store.Task
		_ = json.Unmarshal(b, &tasks)
		return tasksLoadedMsg{tasks: tasks}
	}
}

func (a *App) cmdLoadLogs(taskID string) tea.Cmd {
	return func() tea.Msg {
		resp, err := a.client.Call("log.list", map[string]any{
			"task_id": taskID,
			"limit":   50,
			"offset":  0,
		})
		if err != nil || resp.Error != "" {
			return logsLoadedMsg{}
		}
		b, _ := json.Marshal(resp.Data)
		var logs []*store.RunLog
		_ = json.Unmarshal(b, &logs)
		return logsLoadedMsg{logs: logs}
	}
}

func (a *App) cmdSaveTask(t *store.Task) tea.Cmd {
	return func() tea.Msg {
		var method string
		params := map[string]any{
			"name":             t.Name,
			"prompt":           t.Prompt,
			"schedule_type":    string(t.ScheduleType),
			"schedule_expr":    t.ScheduleExpr,
			"work_dir":         t.WorkDir,
			"skip_permissions": t.SkipPermissions,
			"enabled":          t.Enabled,
		}
		if t.ID == "" {
			method = "task.create"
		} else {
			method = "task.update"
			params["id"] = t.ID
		}
		resp, err := a.client.Call(method, params)
		if err != nil {
			return errMsg{err}
		}
		if resp.Error != "" {
			return errMsg{fmt.Errorf(resp.Error)}
		}
		return a.cmdLoadTasks()()
	}
}

func (a *App) cmdDeleteTask(id string) tea.Cmd {
	return func() tea.Msg {
		_, _ = a.client.Call("task.delete", map[string]any{"id": id})
		return a.cmdLoadTasks()()
	}
}

func (a *App) cmdRunTask(id string) tea.Cmd {
	return func() tea.Msg {
		_, _ = a.client.Call("task.run", map[string]any{"id": id})
		return a.cmdLoadTasks()()
	}
}

func (a *App) cmdToggleTask(id string) tea.Cmd {
	return func() tea.Msg {
		resp, err := a.client.Call("task.list", nil)
		if err != nil || resp.Error != "" {
			return nil
		}
		b, _ := json.Marshal(resp.Data)
		var tasks []*store.Task
		_ = json.Unmarshal(b, &tasks)
		for _, t := range tasks {
			if t.ID == id {
				t.Enabled = !t.Enabled
				_, _ = a.client.Call("task.update", map[string]any{
					"id":               t.ID,
					"name":             t.Name,
					"prompt":           t.Prompt,
					"schedule_type":    string(t.ScheduleType),
					"schedule_expr":    t.ScheduleExpr,
					"work_dir":         t.WorkDir,
					"skip_permissions": t.SkipPermissions,
					"enabled":          t.Enabled,
				})
				break
			}
		}
		return a.cmdLoadTasks()()
	}
}
