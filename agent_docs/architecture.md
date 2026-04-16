# Architecture

## IPC layer (`internal/ipc/`)

Client and daemon communicate over a Unix socket (`~/.local/share/tempo/daemon.sock`) using newline-delimited JSON. Each message is a `Request` (with `method` + `params`) or `Response` (with `data` or `error`). The server dispatches on method name using registered `HandlerFunc`s.

## Daemon (`internal/daemon/`)

`daemon.Run` wires everything together: opens SQLite store → creates `Runner` + `Scheduler` → loads existing enabled tasks → starts the gocron scheduler → starts the IPC server.

Handlers are registered in `handlers.go` for these IPC methods:
- `daemon.ping`
- `task.create`, `task.update`, `task.delete`, `task.list`, `task.run`
- `log.list`

## Scheduler (`internal/scheduler/`)

`Scheduler` wraps `gocron.Scheduler` and maps task IDs to live `gocron.Job` handles. `Runner.Run` dispatches to the appropriate agent via `agent.Registry.Get(task.Agent)`, then calls `BuildCommand` to construct the subprocess, writing a `RunLog` before and after.

### Agent abstraction (`internal/agent/`)

`agent.Registry` holds one implementation per `Kind` (`"claude"`, `"codex"`). `Runner` holds a `*Registry` and calls `Get(agent.Kind(task.Agent))` at run time — unknown or empty values fall back to the Claude agent. Each implementation's `BuildCommand` constructs the correct `exec.Cmd` for that CLI tool.

Schedule types and their `ScheduleExpr` formats:
| Type       | Expression format          | Example        |
|------------|----------------------------|----------------|
| `once`     | RFC3339 datetime           | `2025-06-01T09:00:00Z` |
| `interval` | Go duration string         | `1h30m`        |
| `daily`    | `HH:MM`                    | `09:00`        |
| `weekly`   | `day:HH:MM`                | `mon:09:00`    |
| `cron`     | standard 5-field cron expr | `0 9 * * 1`    |

Scheduled jobs run with a 10-minute `context.WithTimeout`. Manual runs (`task.run`) also go through `Runner.Run` with `triggered="manual"`.

## Store (`internal/store/`)

`Store` interface backed by `modernc.org/sqlite` (pure Go, no CGO required). Schema migrations run sequentially at startup from `migrations.go`. Two tables: `tasks` and `run_logs`.

## TUI (`internal/tui/`)

Bubbletea model (`App`) with three view states:
- `stateList` — task list with status, last run, enable/disable toggle
- `stateForm` — create or edit a task (name, prompt, schedule, work dir, skip-permissions)
- `stateLogs` — paginated run log viewer for a selected task

`App.Init()` kicks off a 2-second polling loop (`tick()`) that refreshes the task list and daemon reachability. All IPC calls are wrapped as Bubbletea `tea.Cmd` functions.

Sub-packages:
- `internal/tui/views/` — `TaskListView`, `TaskFormView`, `LogViewerView`
- `internal/tui/shared/` — shared keymap and lipgloss styles
- `internal/tui/components/` — reusable status bar, help modal, confirm dialog

## Config (`internal/config/`)

Data dir: `$XDG_DATA_HOME/tempo` → falls back to `~/.local/share/tempo`.

Env overrides:
- `XDG_DATA_HOME` — changes the data directory root
- `CLAUDE_BIN` — path to the `claude` binary (default: `"claude"`)
- `CODEX_BIN` — path to the `codex` binary (default: `"codex"`)
