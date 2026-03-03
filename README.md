# tempo

Cron for Claude — run prompts on a schedule from a terminal UI.

## Installation

### Linux

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_linux_amd64.tar.gz | tar -xz
sudo mv tempod tempo /usr/local/bin/
```

For ARM64:

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_linux_arm64.tar.gz | tar -xz
sudo mv tempod tempo /usr/local/bin/
```

After installing, start the daemon:

```bash
tempod &
```

Or install as a systemd user service (see [manual install](#manual-install-from-source)).

### macOS

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_darwin_arm64.tar.gz | tar -xz
mv tempod tempo /usr/local/bin/
```

For Intel Macs:

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_darwin_amd64.tar.gz | tar -xz
mv tempod tempo /usr/local/bin/
```

After installing, start the daemon:

```bash
tempod &
```

Or install as a launchd user agent (see [manual install](#manual-install-from-source)).

### Windows

1. Download [tempo_windows_amd64.zip](https://github.com/tenlisboa/tempo/releases/latest/download/tempo_windows_amd64.zip)
2. Extract the archive
3. Move `tempod.exe` and `tempo.exe` to a directory in your `PATH`
4. Start the daemon: `tempod.exe`

## Usage

The daemon must be running before using the TUI:

```bash
tempod &   # start daemon
tempo      # open TUI
```

The TUI lets you create, edit, enable, and disable scheduled Claude prompts.

## Manual install from source

Requires Go 1.23+.

```bash
git clone https://github.com/tenlisboa/tempo
cd tempo
make install
```

`make install` builds both binaries, copies them to `~/.local/bin`, and registers the daemon as a system service (systemd on Linux, launchd on macOS).

To uninstall:

```bash
make uninstall
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CLAUDE_BIN` | `claude` | Path to the Claude CLI binary |
| `XDG_DATA_HOME` | `~/.local/share` | Data directory (socket + DB live in `$XDG_DATA_HOME/tempo`) |

## Requirements

- [Claude Code](https://claude.ai/code) CLI installed and authenticated
