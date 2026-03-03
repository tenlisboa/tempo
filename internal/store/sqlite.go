package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type sqliteStore struct {
	db *sql.DB
}

func NewSQLite(path string) (Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &sqliteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *sqliteStore) migrate() error {
	var current int
	_ = s.db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&current)

	for i, m := range migrations {
		if i < current {
			continue
		}
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
		if _, err := s.db.Exec(`UPDATE schema_version SET version = ?`, i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

func (s *sqliteStore) CreateTask(t *Task) error {
	_, err := s.db.Exec(
		`INSERT INTO tasks (id,name,prompt,schedule_type,schedule_expr,enabled,work_dir,skip_permissions,created_at,updated_at,last_run_at,last_exit_code)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.ID, t.Name, t.Prompt, t.ScheduleType, t.ScheduleExpr,
		boolInt(t.Enabled), t.WorkDir, boolInt(t.SkipPermissions),
		t.CreatedAt.Unix(), t.UpdatedAt.Unix(),
		timePtr(t.LastRunAt), t.LastExitCode,
	)
	return err
}

func (s *sqliteStore) UpdateTask(t *Task) error {
	t.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE tasks SET name=?,prompt=?,schedule_type=?,schedule_expr=?,enabled=?,work_dir=?,skip_permissions=?,updated_at=? WHERE id=?`,
		t.Name, t.Prompt, t.ScheduleType, t.ScheduleExpr,
		boolInt(t.Enabled), t.WorkDir, boolInt(t.SkipPermissions),
		t.UpdatedAt.Unix(), t.ID,
	)
	return err
}

func (s *sqliteStore) DeleteTask(id string) error {
	_, err := s.db.Exec(`DELETE FROM tasks WHERE id=?`, id)
	return err
}

func (s *sqliteStore) GetTask(id string) (*Task, error) {
	row := s.db.QueryRow(`SELECT * FROM tasks WHERE id=?`, id)
	return scanTask(row)
}

func (s *sqliteStore) ListTasks() ([]*Task, error) {
	rows, err := s.db.Query(`SELECT * FROM tasks ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []*Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *sqliteStore) UpdateTaskLastRun(id string, at time.Time, exitCode int) error {
	_, err := s.db.Exec(
		`UPDATE tasks SET last_run_at=?,last_exit_code=?,updated_at=? WHERE id=?`,
		at.Unix(), exitCode, time.Now().Unix(), id,
	)
	return err
}

func (s *sqliteStore) CreateRunLog(r *RunLog) error {
	_, err := s.db.Exec(
		`INSERT INTO run_logs (id,task_id,started_at,ended_at,exit_code,output,triggered) VALUES (?,?,?,?,?,?,?)`,
		r.ID, r.TaskID, r.StartedAt.Unix(),
		timePtr(r.EndedAt), r.ExitCode,
		truncate(r.Output, 512*1024), r.Triggered,
	)
	return err
}

func (s *sqliteStore) UpdateRunLog(r *RunLog) error {
	_, err := s.db.Exec(
		`UPDATE run_logs SET ended_at=?,exit_code=?,output=? WHERE id=?`,
		timePtr(r.EndedAt), r.ExitCode, truncate(r.Output, 512*1024), r.ID,
	)
	return err
}

func (s *sqliteStore) ListRunLogs(taskID string, limit, offset int) ([]*RunLog, error) {
	rows, err := s.db.Query(
		`SELECT id,task_id,started_at,ended_at,exit_code,output,triggered FROM run_logs WHERE task_id=? ORDER BY started_at DESC LIMIT ? OFFSET ?`,
		taskID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*RunLog
	for rows.Next() {
		var r RunLog
		var startedAt, endedAt sql.NullInt64
		var exitCode sql.NullInt64
		if err := rows.Scan(&r.ID, &r.TaskID, &startedAt, &endedAt, &exitCode, &r.Output, &r.Triggered); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			r.StartedAt = time.Unix(startedAt.Int64, 0)
		}
		if endedAt.Valid {
			t := time.Unix(endedAt.Int64, 0)
			r.EndedAt = &t
		}
		if exitCode.Valid {
			v := int(exitCode.Int64)
			r.ExitCode = &v
		}
		logs = append(logs, &r)
	}
	return logs, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (*Task, error) {
	var t Task
	var createdAt, updatedAt int64
	var lastRunAt sql.NullInt64
	var lastExitCode sql.NullInt64
	var enabled, skipPerms int
	err := s.Scan(
		&t.ID, &t.Name, &t.Prompt,
		&t.ScheduleType, &t.ScheduleExpr,
		&enabled, &t.WorkDir, &skipPerms,
		&createdAt, &updatedAt,
		&lastRunAt, &lastExitCode,
	)
	if err != nil {
		return nil, err
	}
	t.Enabled = enabled == 1
	t.SkipPermissions = skipPerms == 1
	t.CreatedAt = time.Unix(createdAt, 0)
	t.UpdatedAt = time.Unix(updatedAt, 0)
	if lastRunAt.Valid {
		ts := time.Unix(lastRunAt.Int64, 0)
		t.LastRunAt = &ts
	}
	if lastExitCode.Valid {
		v := int(lastExitCode.Int64)
		t.LastExitCode = &v
	}
	return &t, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func timePtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	v := t.Unix()
	return &v
}

func truncate(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes]
}
