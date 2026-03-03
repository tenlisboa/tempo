# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## WHY - Purpose

`tempo` is a personal Claude task scheduler: a background daemon (`tempod`) that runs `claude --print <prompt>` on a schedule, paired with a terminal UI (`tempo`) for managing tasks. Think "cron for Claude prompts."

## WHAT - Architecture

```
cmd/
├── tempo/      # TUI binary (Bubbletea)
└── tempod/     # Daemon binary
internal/
├── config/     # XDG data dir, CLAUDE_BIN env
├── daemon/     # Startup wiring + IPC handler registration
├── ipc/        # Unix socket, newline-delimited JSON RPC
├── scheduler/  # gocron wrapper + claude subprocess runner
├── store/      # Store interface, SQLite impl, migrations
└── tui/        # App model, views, shared styles, components
systemd/        # User-level systemd unit for tempod
```

**Stack:** Go 1.23, Bubbletea (TUI), gocron (scheduler), modernc/sqlite (pure Go, no CGO)

## HOW - Key Commands

```bash
make build           # bin/tempod + bin/tempo
make install         # build + install to ~/.local/bin + enable systemd user service
go run ./cmd/tempod  # run daemon directly
go run ./cmd/tempo   # run TUI (daemon must be running)
```

## Documentation Pointers

| Topic | File |
|-------|------|
| IPC protocol, daemon wiring, schedule types, store | `agent_docs/architecture.md` |
| Adding IPC methods, schedule types, systemd tips | `agent_docs/development.md` |

## Critical Notes

- The TUI connects to the daemon over a Unix socket — `tempo` will silently degrade if `tempod` is not running
- Data lives in `$XDG_DATA_HOME/tempo` (default `~/.local/share/tempo`); socket and DB paths derive from there
- Override the `claude` binary with `CLAUDE_BIN` env var
- Scheduled jobs time out after 10 minutes (`context.WithTimeout`)
- No test suite — validate changes by running both binaries manually
