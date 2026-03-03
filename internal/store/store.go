package store

import "time"

type ScheduleType string

const (
	ScheduleOnce     ScheduleType = "once"
	ScheduleInterval ScheduleType = "interval"
	ScheduleDaily    ScheduleType = "daily"
	ScheduleWeekly   ScheduleType = "weekly"
	ScheduleCron     ScheduleType = "cron"
)

type Task struct {
	ID               string
	Name             string
	Prompt           string
	ScheduleType     ScheduleType
	ScheduleExpr     string
	Enabled          bool
	WorkDir          string
	SkipPermissions  bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastRunAt        *time.Time
	LastExitCode     *int
}

type RunLog struct {
	ID        string
	TaskID    string
	StartedAt time.Time
	EndedAt   *time.Time
	ExitCode  *int
	Output    string
	Triggered string
}

type Store interface {
	CreateTask(t *Task) error
	UpdateTask(t *Task) error
	DeleteTask(id string) error
	GetTask(id string) (*Task, error)
	ListTasks() ([]*Task, error)
	UpdateTaskLastRun(id string, at time.Time, exitCode int) error

	CreateRunLog(r *RunLog) error
	UpdateRunLog(r *RunLog) error
	ListRunLogs(taskID string, limit, offset int) ([]*RunLog, error)

	Close() error
}
