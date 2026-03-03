# Development

## Build

```bash
make build           # builds bin/tempod and bin/tempo
make install         # builds, copies to ~/.local/bin, enables systemd user service
make uninstall       # disables service, removes binaries
make clean           # removes bin/
```

Run without installing:
```bash
go run ./cmd/tempod  # start daemon (blocks, listens on socket)
go run ./cmd/tempo   # launch TUI (requires daemon running)
```

## Testing

No automated test suite currently. Test manually by running both binaries and exercising the TUI.

## Adding a new IPC method

1. Add types to `internal/ipc/types.go`
2. Add a handler function in `internal/daemon/handlers.go`
3. Register it via `srv.Handle("method.name", fn)` inside `h.register(srv)`

## Adding a new schedule type

1. Add a `ScheduleType` constant to `internal/store/store.go`
2. Add a `case` to `jobDef()` in `internal/scheduler/scheduler.go`

## Systemd service

`systemd/tempod.service` is a user-level unit. It expects the binary at `~/.local/bin/tempod`. `make install` handles copying and enabling it.

Check service status:
```bash
systemctl --user status tempod
journalctl --user -u tempod -f
```
