package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/delvop-dev/delvop/internal/provider"
	"github.com/delvop-dev/delvop/internal/session"
	"github.com/delvop-dev/delvop/internal/tui/styles"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	var view string
	switch m.viewMode {
	case ViewFocused:
		view = m.viewFocused()
	default:
		view = m.viewDashboard()
	}

	if m.inputMode {
		view = view + "\n" + m.renderInput()
	}

	return view
}

func (m Model) viewDashboard() string {
	sessions := m.manager.All()
	kpi := m.manager.KPI()

	if len(sessions) == 0 {
		return m.renderEmptyState()
	}

	var sections []string

	// Status bar
	sections = append(sections, m.renderStatusBar(len(sessions)))

	// KPI bar
	sections = append(sections, m.renderKPIBar(kpi))

	// Agent grid
	sections = append(sections, m.renderAgentGrid(sessions))

	// Bottom panels: activity feed + action queue
	bottomHeight := m.height - lipgloss.Height(strings.Join(sections, "\n")) - 2
	if bottomHeight < 3 {
		bottomHeight = 3
	}
	feedWidth := m.width / 2
	queueWidth := m.width - feedWidth

	feed := m.renderActivityFeed(sessions, feedWidth, bottomHeight)
	queue := m.renderActionQueue(sessions, queueWidth, bottomHeight)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top, feed, queue)
	sections = append(sections, bottom)

	// Help bar
	sections = append(sections, m.renderHelpBar())

	return strings.Join(sections, "\n")
}

func (m Model) viewFocused() string {
	s, ok := m.manager.Get(m.focusedID)
	if !ok {
		return m.viewDashboard() // graceful fallback
	}

	var sections []string

	// Breadcrumb
	breadcrumb := styles.BrandStyle().Render("delvop") +
		lipgloss.NewStyle().Foreground(styles.TextDim).Render(" > ") +
		lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(s.Name)
	sections = append(sections, styles.StatusBar.Width(m.width).Render(breadcrumb))

	// Agent profile
	stateLabel := stateDisplayLabel(s.State)
	profile := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(s.Name),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Model: %s  Provider: %s", s.Model, s.ProviderName)),
		styles.StateStyle(stateLabel).Render(stateLabel),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Cost: $%.4f  Tokens: %d/%d", s.CostUSD, s.TokensIn, s.TokensOut)),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Running: %s  WorkDir: %s", formatDuration(time.Since(s.CreatedAt)), s.WorkDir)),
	)
	sections = append(sections, lipgloss.NewStyle().Padding(1, 2).Render(profile))

	// Permission detail
	if s.State == provider.StateWaitingForPermission && s.Permission != nil {
		permBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Red).
			Padding(0, 1).
			Width(m.width - 4).
			Render(
				lipgloss.JoinVertical(lipgloss.Left,
					lipgloss.NewStyle().Foreground(styles.Red).Bold(true).Render("Permission Required"),
					lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(fmt.Sprintf("Tool: %s", s.Permission.Tool)),
					lipgloss.NewStyle().Foreground(styles.TextMuted).Render(s.Permission.Description),
					"",
					styles.KeyStyle().Render("y")+styles.DescStyle().Render(" approve  ")+
						styles.KeyStyle().Render("N")+styles.DescStyle().Render(" deny"),
				),
			)
		sections = append(sections, permBox)
	}

	// Live agent output
	if s.PaneContent != "" {
		outputTitle := lipgloss.NewStyle().Foreground(styles.TextSecondary).Bold(true).Render("Agent Output")
		paneLines := strings.Split(s.PaneContent, "\n")
		// Show the last N lines that fit
		maxOutputLines := m.height - 16
		if maxOutputLines < 5 {
			maxOutputLines = 5
		}
		start := 0
		if len(paneLines) > maxOutputLines {
			start = len(paneLines) - maxOutputLines
		}
		if m.scrollOffset > 0 {
			start -= m.scrollOffset
			if start < 0 {
				start = 0
			}
		}
		end := start + maxOutputLines
		if end > len(paneLines) {
			end = len(paneLines)
		}
		visibleLines := paneLines[start:end]
		outputContent := strings.Join(visibleLines, "\n")
		outputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.TextDim).
			Padding(0, 1).
			Width(m.width - 4).
			Render(lipgloss.JoinVertical(lipgloss.Left, outputTitle, outputContent))
		sections = append(sections, outputBox)
	}

	// Work log (recent events)
	logTitle := lipgloss.NewStyle().Foreground(styles.TextSecondary).Bold(true).Render("Recent Activity")
	var logLines []string
	logLines = append(logLines, logTitle)
	events := s.Events
	evtStart := 0
	if len(events) > 5 {
		evtStart = len(events) - 5
	}
	for i := evtStart; i < len(events); i++ {
		e := events[i]
		ts := lipgloss.NewStyle().Foreground(styles.TextDim).Render(e.Time.Format("15:04:05"))
		msg := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(e.Message)
		logLines = append(logLines, fmt.Sprintf("  %s  %s", ts, msg))
	}
	if len(events) == 0 {
		logLines = append(logLines, lipgloss.NewStyle().Foreground(styles.TextDim).Render("  No activity yet"))
	}
	sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(logLines, "\n")))

	// File changes
	if len(s.FileChanges) > 0 {
		fileTitle := lipgloss.NewStyle().Foreground(styles.TextSecondary).Bold(true).Render("File Changes")
		var fileLines []string
		fileLines = append(fileLines, fileTitle)
		shown := s.FileChanges
		if len(shown) > 10 {
			shown = shown[len(shown)-10:]
		}
		for _, fc := range shown {
			op := lipgloss.NewStyle().Foreground(styles.Green).Render(fc.Operation)
			path := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fc.Path)
			fileLines = append(fileLines, fmt.Sprintf("  %s %s", op, path))
		}
		sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(fileLines, "\n")))
	}

	// Help bar
	help := styles.KeyStyle().Render("esc") + styles.DescStyle().Render(" back  ") +
		styles.KeyStyle().Render("enter") + styles.DescStyle().Render(" attach  ") +
		styles.KeyStyle().Render("j/k") + styles.DescStyle().Render(" scroll  ") +
		styles.KeyStyle().Render("tab") + styles.DescStyle().Render(" next  ") +
		styles.KeyStyle().Render("m") + styles.DescStyle().Render(" message  ") +
		styles.KeyStyle().Render("x") + styles.DescStyle().Render(" kill")
	sections = append(sections, styles.HelpBar.Width(m.width).Render(help))

	return strings.Join(sections, "\n")
}

func (m Model) renderStatusBar(agentCount int) string {
	brand := styles.BrandStyle().Render("delvop")
	count := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf(" %d agents", agentCount))

	attention := m.manager.NeedsAttention()
	var attnStr string
	if len(attention) > 0 {
		attnStr = lipgloss.NewStyle().Foreground(styles.Red).Bold(true).
			Render(fmt.Sprintf(" [!] %d need attention", len(attention)))
	}

	var statusStr string
	if m.statusMsg != "" && time.Now().Before(m.statusExpiry) {
		statusStr = lipgloss.NewStyle().Foreground(styles.Amber).Render("  " + m.statusMsg)
	}

	left := brand + count + attnStr + statusStr
	return styles.StatusBar.Width(m.width).Render(left)
}

func (m Model) renderKPIBar(kpi session.KPIData) string {
	parts := []string{
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Active: %d/%d", kpi.ActiveCount, kpi.TotalCount)),
		styles.CostStyle().Render(fmt.Sprintf("Cost: $%.4f", kpi.TotalCost)),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Tokens: %d", kpi.TotalTokens)),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Uptime: %s", formatDuration(kpi.Uptime))),
	}
	return styles.KPIBar.Width(m.width).Render(strings.Join(parts, "  |  "))
}

func (m Model) renderAgentGrid(sessions []*session.Session) string {
	cardWidth := 28
	cardsPerRow := m.width / cardWidth
	if cardsPerRow < 1 {
		cardsPerRow = 1
	}

	var rows []string
	for i := 0; i < len(sessions); i += cardsPerRow {
		end := i + cardsPerRow
		if end > len(sessions) {
			end = len(sessions)
		}
		var cards []string
		for j := i; j < end; j++ {
			selected := j == m.selectedIdx
			cards = append(cards, m.renderAgentCard(sessions[j], selected, cardWidth))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderAgentCard(s *session.Session, selected bool, width int) string {
	stateLabel := stateDisplayLabel(s.State)

	name := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(truncate(s.Name, width-4))
	state := styles.StateStyle(stateLabel).Render(stateLabel)
	model := lipgloss.NewStyle().Foreground(styles.TextDim).Render(s.Model)
	cost := styles.CostStyle().Render(fmt.Sprintf("$%.4f", s.CostUSD))

	var activity string
	if len(s.Events) > 0 {
		last := s.Events[len(s.Events)-1]
		activity = lipgloss.NewStyle().Foreground(styles.TextDim).Render(truncate(last.Message, width-4))
	}

	var files string
	if len(s.FileChanges) > 0 {
		files = lipgloss.NewStyle().Foreground(styles.TextDim).Render(fmt.Sprintf("%d files", len(s.FileChanges)))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, name, state, model, cost)
	if activity != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, activity)
	}
	if files != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, files)
	}

	style := styles.AgentCard.Width(width)
	if selected {
		style = styles.AgentCardFocused.Width(width)
	}

	return style.Render(content)
}

func (m Model) renderEmptyState() string {
	ascii := lipgloss.NewStyle().Foreground(styles.Purple).Render(`
     _      _
    | |    | |
  __| | ___| |_   _____  _ __
 / _  |/ _ \ \ \ / / _ \| '_ \
| (_| |  __/ |\ V / (_) | |_) |
 \__,_|\___|_| \_/ \___/| .__/
                         | |
                         |_|
`)

	subtitle := lipgloss.NewStyle().Foreground(styles.TextMuted).Render("Multi-agent coding orchestrator")
	hint := styles.KeyStyle().Render("n") + styles.DescStyle().Render(" new agent  ") +
		styles.KeyStyle().Render("t") + styles.DescStyle().Render(" template  ") +
		styles.KeyStyle().Render("?") + styles.DescStyle().Render(" help  ") +
		styles.KeyStyle().Render("q") + styles.DescStyle().Render(" quit")

	splash := lipgloss.JoinVertical(lipgloss.Center, ascii, "", subtitle, "", hint)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, splash)
}

func (m Model) renderActivityFeed(sessions []*session.Session, width, height int) string {
	title := lipgloss.NewStyle().Foreground(styles.TextSecondary).Bold(true).Render("Activity")

	var lines []string
	lines = append(lines, title)

	// Collect all events across sessions, show most recent
	type taggedEvent struct {
		name  string
		event session.Event
	}
	var all []taggedEvent
	for _, s := range sessions {
		for _, e := range s.Events {
			all = append(all, taggedEvent{name: s.Name, event: e})
		}
	}

	// Sort by time (already chronological per session, merge and take tail)
	// Simple approach: take last N
	maxLines := height - 2
	if maxLines < 1 {
		maxLines = 1
	}
	start := 0
	if len(all) > maxLines {
		start = len(all) - maxLines
	}
	for i := start; i < len(all); i++ {
		e := all[i]
		ts := lipgloss.NewStyle().Foreground(styles.TextDim).Render(e.event.Time.Format("15:04"))
		name := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(truncate(e.name, 10))
		msg := lipgloss.NewStyle().Foreground(styles.TextDim).Render(truncate(e.event.Message, width-20))
		lines = append(lines, fmt.Sprintf(" %s %s %s", ts, name, msg))
	}

	if len(all) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.TextDim).Render(" No activity"))
	}

	content := strings.Join(lines, "\n")
	return styles.ActivityFeed.Width(width).Height(height).Render(content)
}

func (m Model) renderActionQueue(sessions []*session.Session, width, height int) string {
	title := lipgloss.NewStyle().Foreground(styles.TextSecondary).Bold(true).Render("Action Queue")

	var lines []string
	lines = append(lines, title)

	for _, s := range sessions {
		if s.State == provider.StateWaitingForPermission && s.Permission != nil {
			name := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(truncate(s.Name, 12))
			tool := lipgloss.NewStyle().Foreground(styles.Red).Render(s.Permission.Tool)
			desc := lipgloss.NewStyle().Foreground(styles.TextDim).Render(truncate(s.Permission.Description, width-20))
			lines = append(lines, fmt.Sprintf(" %s  %s", name, tool))
			lines = append(lines, fmt.Sprintf("   %s", desc))
		}
	}

	if len(lines) == 1 {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.TextDim).Render(" No pending actions"))
	}

	content := strings.Join(lines, "\n")
	return styles.ActionQueue.Width(width).Height(height).Render(content)
}

func (m Model) renderInput() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Padding(0, 1).
		Width(m.width - 4).
		Render(m.textInput.View())
}

func (m Model) renderHelpBar() string {
	help := styles.KeyStyle().Render("n") + styles.DescStyle().Render(" new  ") +
		styles.KeyStyle().Render("enter") + styles.DescStyle().Render(" focus  ") +
		styles.KeyStyle().Render("y") + styles.DescStyle().Render("/") +
		styles.KeyStyle().Render("N") + styles.DescStyle().Render(" approve/deny  ") +
		styles.KeyStyle().Render("m") + styles.DescStyle().Render(" msg  ") +
		styles.KeyStyle().Render("x") + styles.DescStyle().Render(" kill  ") +
		styles.KeyStyle().Render("c") + styles.DescStyle().Render(" compact  ") +
		styles.KeyStyle().Render("?") + styles.DescStyle().Render(" help  ") +
		styles.KeyStyle().Render("q") + styles.DescStyle().Render(" quit")
	return styles.HelpBar.Width(m.width).Render(help)
}

func (m Model) renderHelpOverlay() string {
	title := lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render("delvop Key Bindings")

	bindings := []struct {
		key  string
		desc string
	}{
		{"j/k or arrows", "Navigate agents"},
		{"enter", "Focus agent / attach terminal"},
		{"esc", "Back to dashboard"},
		{"n", "New agent"},
		{"t", "New from template"},
		{"y", "Approve permission"},
		{"N (shift+n)", "Deny permission"},
		{"m", "Send message to agent"},
		{"x", "Kill agent"},
		{"c", "Compact agent context"},
		{"tab", "Next agent (focused view)"},
		{"?", "Toggle help"},
		{"q / ctrl+c", "Quit"},
	}

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")
	for _, b := range bindings {
		k := styles.KeyStyle().Render(fmt.Sprintf("%-16s", b.key))
		d := styles.DescStyle().Render(b.desc)
		lines = append(lines, "  "+k+d)
	}
	lines = append(lines, "")
	lines = append(lines, styles.DescStyle().Render("  Press ? to close"))

	content := strings.Join(lines, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// --- helpers ---

func stateDisplayLabel(state provider.AgentState) string {
	switch state {
	case provider.StateIdle:
		return "IDLE"
	case provider.StateWorking, provider.StateThinking, provider.StateEditing, provider.StateRunningTool:
		return "WORKING"
	case provider.StateWaitingForPermission, provider.StateWaitingInput:
		return "NEEDS INPUT"
	case provider.StateCompacting:
		return "WORKING"
	case provider.StateError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func formatAgo(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}
