# tempo

Cron for Claude — run prompts on a schedule from a terminal UI.

## Installation

### Linux

**With auto-start at login (recommended)**

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_linux_amd64.tar.gz | tar -xz
sudo mv tempod tempo /usr/local/bin/
mkdir -p ~/.config/systemd/user
cp tempod.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now tempod
```

**Without auto-start**

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_linux_amd64.tar.gz | tar -xz
sudo mv tempod tempo /usr/local/bin/
tempod &
```

For ARM64, replace `linux_amd64` with `linux_arm64` in the download URL.

### macOS

**With auto-start at login (recommended)**

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_darwin_arm64.tar.gz | tar -xz
mv tempod tempo /usr/local/bin/
sed "s|/Users/Shared/placeholder|$(which tempod)|" com.tempod.plist \
  > ~/Library/LaunchAgents/com.tempod.plist
launchctl load -w ~/Library/LaunchAgents/com.tempod.plist
```

**Without auto-start**

```bash
curl -L https://github.com/tenlisboa/tempo/releases/latest/download/tempo_darwin_arm64.tar.gz | tar -xz
mv tempod tempo /usr/local/bin/
tempod &
```

For Intel Macs, replace `darwin_arm64` with `darwin_amd64` in the download URL.

### Windows

**With auto-start at login (recommended)**

1. Download and extract [tempo_windows_amd64.zip](https://github.com/tenlisboa/tempo/releases/latest/download/tempo_windows_amd64.zip)
2. Move `tempo.exe` to a directory in your `PATH`
3. Run in PowerShell from the extracted folder:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
.\install-service.ps1
```

This registers `tempod` as a Task Scheduler task that starts at login and restarts automatically if it crashes. No third-party tools required.

**Without auto-start**

1. Download and extract [tempo_windows_amd64.zip](https://github.com/tenlisboa/tempo/releases/latest/download/tempo_windows_amd64.zip)
2. Move `tempod.exe` and `tempo.exe` to a directory in your `PATH`
3. Start the daemon manually: `tempod.exe`

## Usage

Once the daemon is running, open the TUI:

```bash
tempo
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
