package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tenlisboa/tempo/internal/config"
	"github.com/tenlisboa/tempo/internal/ipc"
	"github.com/tenlisboa/tempo/internal/tui"
)

func main() {
	cfg := config.Load()

	pwd, _ := os.Getwd()
	client := ipc.NewClient(cfg.SocketPath)
	app := tui.NewApp(client, pwd)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
