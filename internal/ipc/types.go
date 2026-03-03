package ipc

import "github.com/tenlisboa/tempo/internal/store"

type Request struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type Response struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

type PingResponse struct {
	Version       string `json:"version"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type TaskListResponse = []*store.Task

type TaskCreateParams = store.Task

type TaskUpdateParams = store.Task

type TaskDeleteParams struct {
	ID string `json:"id"`
}

type TaskDeleteResponse struct {
	OK bool `json:"ok"`
}

type TaskRunParams struct {
	ID string `json:"id"`
}

type LogListParams struct {
	TaskID string `json:"task_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type LogListResponse = []*store.RunLog
