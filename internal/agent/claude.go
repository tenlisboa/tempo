package agent

import (
	"context"
	"os/exec"
)

type Claude struct {
	Bin string
}

func (c *Claude) Kind() Kind {
	return KindClaude
}

func (c *Claude) BuildCommand(ctx context.Context, prompt string, yolo bool, workDir string) *exec.Cmd {
	args := []string{"--print", prompt}
	if yolo {
		args = append(args, "--dangerously-skip-permissions")
	}
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	return cmd
}
