package config

import (
	"os"
	"path/filepath"
)

const AppName = "tempo"
const Version = "0.1.0"

type Config struct {
	DataDir    string
	SocketPath string
	DBPath     string
	ClaudeBin  string
}

func Load() *Config {
	dataDir := filepath.Join(xdgDataHome(), AppName)
	return &Config{
		DataDir:    dataDir,
		SocketPath: filepath.Join(dataDir, "daemon.sock"),
		DBPath:     filepath.Join(dataDir, "tasks.db"),
		ClaudeBin:  claudeBin(),
	}
}

func xdgDataHome() string {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

func claudeBin() string {
	if b := os.Getenv("CLAUDE_BIN"); b != "" {
		return b
	}
	return "claude"
}
