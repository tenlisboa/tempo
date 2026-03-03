package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tenlisboa/tempo/internal/tui/shared"
)

type ConfirmMsg struct {
	Confirmed bool
	Tag       string
}

type Confirm struct {
	Message string
	Tag     string
}

func NewConfirm(message, tag string) Confirm {
	return Confirm{Message: message, Tag: tag}
}

func (c Confirm) Update(msg tea.KeyMsg) (Confirm, *ConfirmMsg) {
	switch msg.String() {
	case "y", "Y":
		return c, &ConfirmMsg{Confirmed: true, Tag: c.Tag}
	case "n", "N", "esc":
		return c, &ConfirmMsg{Confirmed: false, Tag: c.Tag}
	}
	return c, nil
}

func (c Confirm) View() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(shared.ColorDanger).
		Padding(1, 3).
		Render(fmt.Sprintf("%s\n\n%s",
			c.Message,
			shared.StyleSubtle.Render("y = yes   n / esc = no"),
		))
	return lipgloss.NewStyle().Padding(1).Render(box)
}
