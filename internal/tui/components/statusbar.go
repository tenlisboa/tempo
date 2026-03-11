package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/tenlisboa/tempo/internal/tui/shared"
)

func StatusBar(width int, daemonOK bool, daemonRestarting bool, helpText string) string {
	var status string
	switch {
	case daemonRestarting:
		status = shared.StyleDaemonRestarting.Render("◌ daemon: restarting...")
	case daemonOK:
		status = shared.StyleDaemonOK.Render("● daemon: connected")
	default:
		status = shared.StyleDaemonErr.Render("○ daemon: disconnected")
	}

	help := shared.StyleHelp.Render(helpText)

	gap := width - lipgloss.Width(status) - lipgloss.Width(help)
	if gap < 0 {
		gap = 0
	}

	return fmt.Sprintf("%s%s%s",
		status,
		lipgloss.NewStyle().Width(gap).Render(""),
		help,
	)
}
