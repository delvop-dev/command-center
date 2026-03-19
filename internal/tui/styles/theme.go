package styles

import "github.com/charmbracelet/lipgloss"

var (
	Background    = lipgloss.Color("#0f0f14")
	Surface       = lipgloss.Color("#13131a")
	Border        = lipgloss.Color("#1e1e2e")
	TextPrimary   = lipgloss.Color("#e2e4f0")
	TextSecondary = lipgloss.Color("#a0a4b8")
	TextMuted     = lipgloss.Color("#787c99")
	TextDim       = lipgloss.Color("#4a4a5e")
	TextGhost     = lipgloss.Color("#3a3a4e")
	Purple        = lipgloss.Color("#8b7cf6")
	Green         = lipgloss.Color("#3dd68c")
	Red           = lipgloss.Color("#f06449")
	Blue          = lipgloss.Color("#4a8bf5")
	Amber         = lipgloss.Color("#e0af68")
)

var (
	StatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a2e")).
			Foreground(TextSecondary).
			Padding(0, 1)

	KPIBar = lipgloss.NewStyle().
		Foreground(TextMuted).
		Padding(0, 1)

	AgentCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1).
			Width(20)

	AgentCardFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(0, 1).
				Width(20)

	ActivityFeed = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(Border).
			Padding(0, 1)

	ActionQueue = lipgloss.NewStyle().
			Padding(0, 1)

	HelpBar = lipgloss.NewStyle().
		Foreground(TextGhost).
		Padding(0, 1)
)

func StateColor(state string) lipgloss.Color {
	switch state {
	case "WORKING":
		return Green
	case "NEEDS INPUT":
		return Red
	case "DONE":
		return Blue
	case "IDLE":
		return TextMuted
	case "ERROR":
		return Red
	default:
		return TextDim
	}
}

func StateStyle(state string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(StateColor(state)).Bold(true)
}

func BrandStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Purple).Bold(true)
}

func CostStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Amber)
}

func KeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(Purple).Bold(true)
}

func DescStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(TextGhost)
}

func ProgressBar(pct int, width int) string {
	filled := width * pct / 100
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return lipgloss.NewStyle().Foreground(Purple).Render(bar)
}
