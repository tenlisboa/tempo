package agent

import (
	"context"
	"os/exec"
)

type Kind string

const (
	KindClaude Kind = "claude"
	KindCodex  Kind = "codex"
)

func (k Kind) Valid() bool {
	return k == KindClaude || k == KindCodex
}

type Agent interface {
	Kind() Kind
	BuildCommand(ctx context.Context, prompt string, yolo bool, workDir string) *exec.Cmd
}

type Registry struct {
	byKind map[Kind]Agent
}

func NewRegistry(claudeBin, codexBin string) *Registry {
	r := &Registry{byKind: make(map[Kind]Agent)}
	r.byKind[KindClaude] = &Claude{Bin: claudeBin}
	r.byKind[KindCodex] = &Codex{Bin: codexBin}
	return r
}

func (r *Registry) Get(k Kind) Agent {
	if a, ok := r.byKind[k]; ok {
		return a
	}
	return r.byKind[KindClaude]
}
