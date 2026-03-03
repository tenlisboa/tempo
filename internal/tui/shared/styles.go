package shared

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary   = lipgloss.Color("#7C3AED")
	ColorHighlight = lipgloss.Color("#A78BFA")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorDanger    = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorFg        = lipgloss.Color("#F9FAFB")
	ColorBorder    = lipgloss.Color("#374151")
	ColorSelected  = lipgloss.Color("#1F2937")

	StyleTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	StyleSubtle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleBadgeEnabled = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StyleBadgeDisabled = lipgloss.NewStyle().
				Foreground(ColorMuted)

	StyleBadgeSuccess = lipgloss.NewStyle().
				Foreground(ColorSuccess).
				Bold(true)

	StyleBadgeError = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	StyleSelected = lipgloss.NewStyle().
			Background(ColorSelected).
			Foreground(ColorFg)

	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	StyleDivider = lipgloss.NewStyle().
			Foreground(ColorBorder)

	StyleCount = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleModalTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	StyleDaemonOK = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StyleDaemonErr = lipgloss.NewStyle().
			Foreground(ColorDanger)

	StyleFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary)

	StyleBlurred = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder)
)
