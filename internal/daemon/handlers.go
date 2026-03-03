package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tenlisboa/tempo/internal/config"
	"github.com/tenlisboa/tempo/internal/ipc"
	"github.com/tenlisboa/tempo/internal/scheduler"
	"github.com/tenlisboa/tempo/internal/store"
)

type handlers struct {
	st      store.Store
	sched   *scheduler.Scheduler
	startAt time.Time
}

func (h *handlers) register(srv *ipc.Server) {
	srv.Handle("daemon.ping", h.ping)
	srv.Handle("task.list", h.taskList)
	srv.Handle("task.create", h.taskCreate)
	srv.Handle("task.update", h.taskUpdate)
	srv.Handle("task.delete", h.taskDelete)
	srv.Handle("task.run", h.taskRun)
	srv.Handle("log.list", h.logList)
}

func (h *handlers) ping(req *ipc.Request) (any, error) {
	return ipc.PingResponse{
		Version:       config.Version,
		UptimeSeconds: int64(time.Since(h.startAt).Seconds()),
	}, nil
}

func (h *handlers) taskList(req *ipc.Request) (any, error) {
	return h.st.ListTasks()
}

func (h *handlers) taskCreate(req *ipc.Request) (any, error) {
	t, err := decodeTask(req.Params)
	if err != nil {
		return nil, err
	}
	t.ID = uuid.NewString()
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	if err := h.st.CreateTask(t); err != nil {
		return nil, err
	}
	if t.Enabled {
		_ = h.sched.Add(t)
	}
	return t, nil
}

func (h *handlers) taskUpdate(req *ipc.Request) (any, error) {
	t, err := decodeTask(req.Params)
	if err != nil {
		return nil, err
	}
	if err := h.st.UpdateTask(t); err != nil {
		return nil, err
	}
	_ = h.sched.Remove(t.ID)
	if t.Enabled {
		_ = h.sched.Add(t)
	}
	return t, nil
}

func (h *handlers) taskDelete(req *ipc.Request) (any, error) {
	id, ok := req.Params["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id required")
	}
	_ = h.sched.Remove(id)
	if err := h.st.DeleteTask(id); err != nil {
		return nil, err
	}
	return ipc.TaskDeleteResponse{OK: true}, nil
}

func (h *handlers) taskRun(req *ipc.Request) (any, error) {
	id, ok := req.Params["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id required")
	}
	task, err := h.st.GetTask(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return h.sched.RunNow(ctx, task)
}

func (h *handlers) logList(req *ipc.Request) (any, error) {
	taskID, _ := req.Params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id required")
	}
	limit := 50
	offset := 0
	if v, ok := req.Params["limit"].(float64); ok {
		limit = int(v)
	}
	if v, ok := req.Params["offset"].(float64); ok {
		offset = int(v)
	}
	return h.st.ListRunLogs(taskID, limit, offset)
}

func decodeTask(params map[string]any) (*store.Task, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	var raw struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Prompt          string `json:"prompt"`
		ScheduleType    string `json:"schedule_type"`
		ScheduleExpr    string `json:"schedule_expr"`
		Enabled         bool   `json:"enabled"`
		WorkDir         string `json:"work_dir"`
		SkipPermissions bool   `json:"skip_permissions"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	return &store.Task{
		ID:              raw.ID,
		Name:            raw.Name,
		Prompt:          raw.Prompt,
		ScheduleType:    store.ScheduleType(raw.ScheduleType),
		ScheduleExpr:    raw.ScheduleExpr,
		Enabled:         raw.Enabled,
		WorkDir:         raw.WorkDir,
		SkipPermissions: raw.SkipPermissions,
	}, nil
}
