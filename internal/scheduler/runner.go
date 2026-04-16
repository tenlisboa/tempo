package scheduler

import (
	"context"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/tenlisboa/tempo/internal/agent"
	"github.com/tenlisboa/tempo/internal/store"
)

type Runner struct {
	st       store.Store
	registry *agent.Registry
}

func NewRunner(st store.Store, registry *agent.Registry) *Runner {
	return &Runner{st: st, registry: registry}
}

func (r *Runner) Run(ctx context.Context, task *store.Task, triggered string) (*store.RunLog, error) {
	log := &store.RunLog{
		ID:        uuid.NewString(),
		TaskID:    task.ID,
		StartedAt: time.Now(),
		Triggered: triggered,
	}
	if err := r.st.CreateRunLog(log); err != nil {
		return nil, err
	}

	ag := r.registry.Get(agent.Kind(task.Agent))
	cmd := ag.BuildCommand(ctx, task.Prompt, task.SkipPermissions, task.WorkDir)

	out, err := cmd.CombinedOutput()
	now := time.Now()
	log.EndedAt = &now
	log.Output = string(out)

	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			exitCode = -1
		}
	}
	log.ExitCode = &exitCode

	_ = r.st.UpdateRunLog(log)
	_ = r.st.UpdateTaskLastRun(task.ID, now, exitCode)

	return log, nil
}
