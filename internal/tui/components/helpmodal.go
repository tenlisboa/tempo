package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tenlisboa/tempo/internal/tui/shared"
)

type HelpModal struct{}

func NewHelpModal() HelpModal {
	return HelpModal{}
}

func (h HelpModal) Update(msg tea.KeyMsg) (HelpModal, bool) {
	switch msg.String() {
	case "esc", "?", "q":
		return h, true
	}
	return h, false
}

func (h HelpModal) View(width, height int) string {
	type entry struct {
		kind     string
		examples string
		desc     string
	}
	entries := []entry{
		{"interval", "30m  ·  2h  ·  1h30m", "repeat every N duration"},
		{"daily", "09:00  ·  14:30", "run at HH:MM each day (24h)"},
		{"weekly", "mon:09:00  ·  fri:17:00", "run on a weekday at a time"},
		{"cron", "* * * * *  ·  0 9 * * 1-5", "standard 5-field cron expression"},
		{"once", "2026-01-01T09:00:00Z", "run exactly once (RFC3339)"},
	}

	kindStyle := lipgloss.NewStyle().Width(10).Foreground(shared.ColorPrimary).Bold(true)
	exampleStyle := lipgloss.NewStyle().Foreground(shared.ColorHighlight)
	indent := strings.Repeat(" ", 12)

	var lines []string
	for _, e := range entries {
		lines = append(lines, kindStyle.Render(e.kind)+"  "+exampleStyle.Render(e.examples))
		lines = append(lines, indent+shared.StyleSubtle.Render(e.desc))
		lines = append(lines, "")
	}

	footer := lipgloss.NewStyle().Foreground(shared.ColorMuted).Italic(true).Render("esc · ? to close")
	content := strings.Join(lines, "\n") + footer

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.ColorPrimary).
		Padding(1, 3).
		Render(shared.StyleModalTitle.Render("Schedule Reference") + "\n\n" + content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
