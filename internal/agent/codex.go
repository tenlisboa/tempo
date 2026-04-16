package agent

import (
	"context"
	"os/exec"
)

type Codex struct {
	Bin string
}

func (c *Codex) Kind() Kind {
	return KindCodex
}

func (c *Codex) BuildCommand(ctx context.Context, prompt string, yolo bool, workDir string) *exec.Cmd {
	args := []string{"exec", "--skip-git-repo-check"}
	if yolo {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}
	args = append(args, prompt)
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	return cmd
}
