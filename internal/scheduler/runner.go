package scheduler

import (
	"context"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/tenlisboa/tempo/internal/store"
)

type Runner struct {
	st        store.Store
	claudeBin string
}

func NewRunner(st store.Store, claudeBin string) *Runner {
	return &Runner{st: st, claudeBin: claudeBin}
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

	args := []string{"--print", task.Prompt}
	if task.SkipPermissions {
		args = append(args, "--dangerously-skip-permissions")
	}

	cmd := exec.CommandContext(ctx, r.claudeBin, args...)
	if task.WorkDir != "" {
		cmd.Dir = task.WorkDir
	}

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
