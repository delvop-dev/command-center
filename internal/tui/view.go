package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/delvop-dev/delvop/internal/provider"
	"github.com/delvop-dev/delvop/internal/security"
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
	case ViewGovernance:
		view = m.viewGovernance()
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
	rule := styles.HRule(m.width)

	// Status bar: "delvop · 4 agents [!] 1 need attention"
	sections = append(sections, m.renderStatusBar(len(sessions)))
	sections = append(sections, rule)

	// KPI bar: "Active: 3/4  Cost: $2.84  Tokens: 108.4k  Uptime: 1h24m"
	sections = append(sections, m.renderKPIBar(kpi))
	sections = append(sections, rule)

	// Agent grid
	sections = append(sections, m.renderAgentGrid(sessions))
	sections = append(sections, rule)

	// Bottom panels: Activity Feed | Action Queue
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

	// Help bar with separator
	sections = append(sections, rule)
	sections = append(sections, styles.HelpBar.Width(m.width).Render(
		m.renderHelpBarContent()))

	return strings.Join(sections, "\n")
}

func (m Model) viewFocused() string {
	s, ok := m.manager.Get(m.focusedID)
	if !ok {
		return m.viewDashboard()
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
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(
			fmt.Sprintf("%s/%s", s.ProviderName, s.Model)),
		stateDot(stateLabel)+" "+styles.StateStyle(stateLabel).Render(stateLabel),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(
			fmt.Sprintf("Cost: $%.2f  Tokens: %s  Running: %s",
				s.CostUSD, formatTokens(s.TokensIn+s.TokensOut), formatDuration(time.Since(s.CreatedAt)))),
		lipgloss.NewStyle().Foreground(styles.TextDim).Render("WorkDir: "+s.WorkDir),
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
					lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(s.Name)+" "+
						lipgloss.NewStyle().Foreground(styles.Red).Render(fmt.Sprintf("Allow %s?", s.Permission.Tool)),
					lipgloss.NewStyle().Foreground(styles.TextMuted).Render(s.Permission.Description),
					"",
					styles.KeyStyle().Render("y")+styles.DescStyle().Render(" approve  ")+
						styles.KeyStyle().Render("N")+styles.DescStyle().Render(" deny"),
				),
			)
		sections = append(sections, permBox)
	}

	// Calculate remaining height for agent output
	usedHeight := lipgloss.Height(strings.Join(sections, "\n"))
	// Reserve: 3 lines for activity + 1 for help bar + 1 buffer
	availableForOutput := m.height - usedHeight - 5
	if availableForOutput < 3 {
		availableForOutput = 3
	}

	// Live agent output — takes all remaining space
	if s.PaneContent != "" {
		outputTitle := lipgloss.NewStyle().Foreground(styles.TextSecondary).Bold(true).Render("Agent Output")
		paneLines := strings.Split(s.PaneContent, "\n")
		// Remove trailing empty lines
		for len(paneLines) > 0 && strings.TrimSpace(paneLines[len(paneLines)-1]) == "" {
			paneLines = paneLines[:len(paneLines)-1]
		}
		maxLines := availableForOutput - 3 // border(2) + title(1)
		if maxLines < 2 {
			maxLines = 2
		}
		start := 0
		if len(paneLines) > maxLines {
			start = len(paneLines) - maxLines
		}
		if m.scrollOffset > 0 {
			start -= m.scrollOffset
			if start < 0 {
				start = 0
			}
		}
		end := start + maxLines
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

	// Recent activity (last 3 events, compact)
	events := s.Events
	if len(events) > 0 {
		evtStart := 0
		if len(events) > 3 {
			evtStart = len(events) - 3
		}
		var logLines []string
		for i := evtStart; i < len(events); i++ {
			e := events[i]
			ts := lipgloss.NewStyle().Foreground(styles.TextDim).Render(e.Time.Format("15:04:05"))
			msg := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(e.Message)
			logLines = append(logLines, fmt.Sprintf(" %s  %s", ts, msg))
		}
		sections = append(sections, strings.Join(logLines, "\n"))
	}

	// Help bar
	help := styles.KeyStyle().Render("esc") + styles.DescStyle().Render(" back  ") +
		styles.KeyStyle().Render("↵") + styles.DescStyle().Render(" attach (ctrl+b d detach)  ") +
		styles.KeyStyle().Render("j/k") + styles.DescStyle().Render(" scroll  ") +
		styles.KeyStyle().Render("tab") + styles.DescStyle().Render(" next  ") +
		styles.KeyStyle().Render("m") + styles.DescStyle().Render(" message  ") +
		lipgloss.NewStyle().Foreground(styles.Red).Render("x") + styles.DescStyle().Render(" kill")
	sections = append(sections, styles.HelpBar.Width(m.width).Render(help))

	return strings.Join(sections, "\n")
}

func (m Model) viewGovernance() string {
	var sections []string

	// Title
	title := styles.BrandStyle().Render("Governance")
	sections = append(sections, styles.StatusBar.Width(m.width).Render(title))
	sections = append(sections, styles.HRule(m.width))

	// Security Rules section
	rulesTitle := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render("Security Rules")
	sections = append(sections, lipgloss.NewStyle().Padding(0, 1).Render(rulesTitle))

	rules := security.AllRules()
	for _, r := range rules {
		disabled := m.gov != nil && m.gov.IsRuleDisabled(r.ID)

		var dot string
		if disabled {
			dot = lipgloss.NewStyle().Foreground(styles.TextDim).Render("○")
		} else {
			dot = lipgloss.NewStyle().Foreground(styles.Green).Render("●")
		}

		var sevStyle lipgloss.Style
		if r.Severity == security.Critical {
			sevStyle = lipgloss.NewStyle().Foreground(styles.Red).Bold(true)
		} else {
			sevStyle = lipgloss.NewStyle().Foreground(styles.Amber)
		}

		id := lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(r.ID)
		sev := sevStyle.Render(string(r.Severity))
		msg := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(r.Message)
		line := fmt.Sprintf("  %s %s  %s  %s", dot, id, sev, msg)
		sections = append(sections, line)
	}

	activeCount := len(rules)
	if m.gov != nil {
		activeCount = m.gov.ActiveRuleCount()
	}
	summary := lipgloss.NewStyle().Foreground(styles.TextDim).Padding(0, 1).Render(
		fmt.Sprintf("%d/%d rules active", activeCount, len(rules)))
	sections = append(sections, summary)
	sections = append(sections, styles.HRule(m.width))

	// Project Rules section
	projTitle := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render("Project Rules")
	sections = append(sections, lipgloss.NewStyle().Padding(0, 1).Render(projTitle))

	if m.gov != nil {
		proj := m.gov.Project
		if proj.Language != "" {
			sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
				lipgloss.NewStyle().Foreground(styles.TextDim).Render("Language: ")+
					lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(proj.Language)))
		}
		sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.NewStyle().Foreground(styles.TextDim).Render("Test before commit: ")+
				m.boolLabel(proj.TestBeforeCommit)))
		sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.NewStyle().Foreground(styles.TextDim).Render("No commit to main: ")+
				m.boolLabel(proj.NoCommitToMain)))
		sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.NewStyle().Foreground(styles.TextDim).Render("Lint on save: ")+
				m.boolLabel(proj.LintOnSave)))
		if proj.MaxFileSizeKB > 0 {
			sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
				lipgloss.NewStyle().Foreground(styles.TextDim).Render("Max file size: ")+
					lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(fmt.Sprintf("%d KB", proj.MaxFileSizeKB))))
		}
	} else {
		sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.NewStyle().Foreground(styles.TextDim).Render("No governance file loaded")))
	}

	sections = append(sections, styles.HRule(m.width))

	// Shared Skills section
	skillsTitle := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render("Shared Skills")
	sections = append(sections, lipgloss.NewStyle().Padding(0, 1).Render(skillsTitle))

	if m.gov != nil && len(m.gov.Skills) > 0 {
		for _, s := range m.gov.Skills {
			name := lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render(s.Name)
			instr := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(s.Instruction)
			sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
				name+"  "+instr))
		}
	} else {
		sections = append(sections, lipgloss.NewStyle().Padding(0, 2).Render(
			lipgloss.NewStyle().Foreground(styles.TextDim).Render("No skills defined")))
	}

	sections = append(sections, styles.HRule(m.width))

	// Help bar
	help := styles.KeyStyle().Render("esc") + styles.DescStyle().Render(" back  ") +
		styles.KeyStyle().Render("g") + styles.DescStyle().Render(" close  ") +
		lipgloss.NewStyle().Foreground(styles.Red).Render("q") + styles.DescStyle().Render(" quit")
	sections = append(sections, styles.HelpBar.Width(m.width).Render(help))

	return strings.Join(sections, "\n")
}

func (m Model) boolLabel(v bool) string {
	if v {
		return lipgloss.NewStyle().Foreground(styles.Green).Render("yes")
	}
	return lipgloss.NewStyle().Foreground(styles.TextDim).Render("no")
}

// --- Status bar ---

func (m Model) renderStatusBar(agentCount int) string {
	brand := styles.BrandStyle().Render("delvop")
	sep := lipgloss.NewStyle().Foreground(styles.TextDim).Render(" · ")
	count := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("%d agents", agentCount))

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

	left := brand + sep + count + attnStr + statusStr
	return styles.StatusBar.Width(m.width).Render(left)
}

// --- KPI bar ---

func (m Model) renderKPIBar(kpi session.KPIData) string {
	active := lipgloss.NewStyle().Foreground(styles.TextDim).Render("Active: ") +
		lipgloss.NewStyle().Foreground(styles.Green).Render(fmt.Sprintf("%d/%d", kpi.ActiveCount, kpi.TotalCount))
	cost := lipgloss.NewStyle().Foreground(styles.TextDim).Render("Cost: ") +
		styles.CostStyle().Render(fmt.Sprintf("$%.2f", kpi.TotalCost))
	tokens := lipgloss.NewStyle().Foreground(styles.TextDim).Render("Tokens: ") +
		lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(formatTokens(kpi.TotalTokens))
	uptime := lipgloss.NewStyle().Foreground(styles.TextDim).Render("Uptime: ") +
		lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(formatDuration(kpi.Uptime))
	parts := []string{active, cost, tokens, uptime}
	return styles.KPIBar.Width(m.width).Render(strings.Join(parts, "    "))
}

// --- Agent grid ---

func (m Model) renderAgentGrid(sessions []*session.Session) string {
	// Calculate card width based on terminal width
	// Target: 4 cards across for wide terminals, fewer for narrow
	cardWidth := m.width / 4
	if cardWidth < 28 {
		cardWidth = 28
	}
	if cardWidth > m.width/2 {
		cardWidth = m.width / 2
	}
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
	innerW := width - 4 // account for border + padding

	// Row 1: name + state with dot (right-aligned state)
	name := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(truncate(s.Name, innerW-15))
	dot := stateDot(stateLabel)
	stateStr := dot + " " + styles.StateStyle(stateLabel).Render(stateLabel)
	if len(s.Alerts) > 0 {
		for _, a := range s.Alerts {
			if a.Severity == security.Critical {
				stateStr = lipgloss.NewStyle().Foreground(styles.Red).Bold(true).Render("! ") + stateStr
				break
			}
		}
	}
	nameWidth := lipgloss.Width(name)
	stateWidth := lipgloss.Width(stateStr)
	gap := innerW - nameWidth - stateWidth
	if gap < 1 {
		gap = 1
	}
	row1 := name + strings.Repeat(" ", gap) + stateStr

	// Row 2: provider/model  cost  tokens  files
	provModel := lipgloss.NewStyle().Foreground(styles.TextDim).Render(
		fmt.Sprintf("%s/%s", s.ProviderName, s.Model))
	cost := styles.CostStyle().Render(fmt.Sprintf("$%.2f", s.CostUSD))
	tokens := lipgloss.NewStyle().Foreground(styles.TextDim).Render(formatTokens(s.TokensIn + s.TokensOut))
	var filesStr string
	if len(s.FileChanges) > 0 {
		filesStr = lipgloss.NewStyle().Foreground(styles.TextDim).Render(fmt.Sprintf("%d files", len(s.FileChanges)))
	}

	row2Parts := []string{provModel, cost, tokens}
	if filesStr != "" {
		row2Parts = append(row2Parts, filesStr)
	}
	row2 := strings.Join(row2Parts, "  ")

	content := lipgloss.JoinVertical(lipgloss.Left, row1, row2)

	style := styles.AgentCard.Width(width)
	if selected {
		style = styles.AgentCardFocused.Width(width)
	}

	return style.Render(content)
}

// --- Activity Feed ---

func (m Model) renderActivityFeed(sessions []*session.Session, width, height int) string {
	title := lipgloss.NewStyle().Foreground(styles.TextDim).Render("Activity Feed")
	titleRule := lipgloss.NewStyle().Foreground(styles.Border).Render(strings.Repeat("─", width-3))

	var lines []string
	lines = append(lines, title)
	lines = append(lines, titleRule)

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

	maxLines := height - 3
	if maxLines < 1 {
		maxLines = 1
	}
	start := 0
	if len(all) > maxLines {
		start = len(all) - maxLines
	}
	for i := start; i < len(all); i++ {
		e := all[i]
		ts := lipgloss.NewStyle().Foreground(styles.TextDim).Render(e.event.Time.Format("15:04:05"))
		nameColor := agentColor(e.name)
		name := lipgloss.NewStyle().Foreground(nameColor).Render(truncate(e.name, 14))
		msg := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(truncate(e.event.Message, width-28))
		if e.event.Type == "security" {
			msg = lipgloss.NewStyle().Foreground(styles.Red).Render(truncate(e.event.Message, width-28))
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", ts, name, msg))
	}

	if len(all) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.TextDim).Render("No activity"))
	}

	content := strings.Join(lines, "\n")
	return styles.ActivityFeed.Width(width).Height(height).Render(content)
}

// --- Action Queue ---

func (m Model) renderActionQueue(sessions []*session.Session, width, height int) string {
	title := lipgloss.NewStyle().Foreground(styles.TextDim).Render("Action Queue")
	titleRule := lipgloss.NewStyle().Foreground(styles.Border).Render(strings.Repeat("─", width-3))

	var sections []string
	sections = append(sections, title)
	sections = append(sections, titleRule)

	for _, s := range sessions {
		if s.State == provider.StateWaitingForPermission && s.Permission != nil {
			// Permission box matching the design mockup
			name := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(s.Name)
			toolLabel := lipgloss.NewStyle().Foreground(styles.Red).Render(
				fmt.Sprintf("Allow %s?", s.Permission.Tool))
			desc := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(
				truncate(s.Permission.Description, width-8))
			hint := styles.KeyStyle().Render("y") + styles.DescStyle().Render(" approve  ") +
				styles.KeyStyle().Render("N") + styles.DescStyle().Render(" deny")

			boxContent := []string{name + " " + toolLabel, desc}

			var alertLines []string
			for _, alert := range s.Alerts {
				alertColor := styles.Amber
				icon := "WARNING"
				if alert.Severity == security.Critical {
					alertColor = styles.Red
					icon = "CRITICAL"
				}
				alertLines = append(alertLines,
					lipgloss.NewStyle().Foreground(alertColor).Bold(true).
						Render(fmt.Sprintf("%s: %s", icon, alert.RuleID)),
					lipgloss.NewStyle().Foreground(alertColor).
						Render(alert.Message))
			}
			boxContent = append(boxContent, alertLines...)
			boxContent = append(boxContent, hint)

			box := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.Red).
				Padding(0, 1).
				Width(width - 4).
				Render(lipgloss.JoinVertical(lipgloss.Left, boxContent...))
			sections = append(sections, box)
		} else if s.State == provider.StateWaitingInput {
			name := lipgloss.NewStyle().Foreground(styles.TextPrimary).Bold(true).Render(s.Name)
			label := lipgloss.NewStyle().Foreground(styles.Amber).Render("needs focus")
			hint := styles.KeyStyle().Render("↵") + styles.DescStyle().Render(" focus  ") +
				styles.KeyStyle().Render("m") + styles.DescStyle().Render(" message")
			sections = append(sections, fmt.Sprintf(" %s  %s  %s", name, label, hint))
		}
	}

	if len(sections) == 2 { // only title + blank line
		sections = append(sections, lipgloss.NewStyle().Foreground(styles.TextGhost).Render("No pending actions"))
	}

	content := strings.Join(sections, "\n")
	return styles.ActionQueue.Width(width).Height(height).Render(content)
}

// --- Shared components ---

func (m Model) renderInput() string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Purple).
		Padding(0, 1).
		Width(m.width - 4).
		Render(m.textInput.View())
}

func (m Model) renderHelpBarContent() string {
	gap := "   "
	items := []string{
		styles.KeyStyle().Render("j/k") + styles.DescStyle().Render(" navigate"),
		styles.KeyStyle().Render("↵") + styles.DescStyle().Render(" attach"),
		lipgloss.NewStyle().Foreground(styles.Green).Bold(true).Render("y") + styles.DescStyle().Render(" approve"),
		lipgloss.NewStyle().Foreground(styles.Red).Bold(true).Render("N") + styles.DescStyle().Render(" deny"),
		styles.KeyStyle().Render("n") + styles.DescStyle().Render(" new"),
		styles.CostStyle().Render("$") + styles.DescStyle().Render(" cost"),
		styles.KeyStyle().Render("g") + styles.DescStyle().Render(" governance"),
		styles.KeyStyle().Render("?") + styles.DescStyle().Render(" help"),
		lipgloss.NewStyle().Foreground(styles.Red).Render("q") + styles.DescStyle().Render(" quit"),
	}
	return strings.Join(items, gap)
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

func (m Model) renderHelpOverlay() string {
	title := lipgloss.NewStyle().Foreground(styles.Purple).Bold(true).Render("delvop Key Bindings")

	bindings := []struct {
		key  string
		desc string
	}{
		{"j/k or arrows", "Navigate agents"},
		{"enter / ↵", "Attach to agent terminal"},
		{"ctrl+\\", "Detach back to delvop"},
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
	case provider.StatePreparing:
		return "PREPARING"
	case provider.StateWorking, provider.StateThinking, provider.StateEditing, provider.StateRunningTool:
		return "WORKING"
	case provider.StateWaitingForPermission:
		return "NEEDS INPUT"
	case provider.StateWaitingInput:
		return "NEEDS FOCUS"
	case provider.StateCompacting:
		return "WORKING"
	case provider.StateError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// agentColor returns a consistent color for an agent name.
func agentColor(name string) lipgloss.Color {
	colors := []lipgloss.Color{
		styles.Purple, // purple
		styles.Green,  // green
		styles.Blue,   // blue
		styles.Amber,  // amber
	}
	hash := 0
	for _, c := range name {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return colors[hash%len(colors)]
}

func stateDot(stateLabel string) string {
	color := styles.StateColor(stateLabel)
	return lipgloss.NewStyle().Foreground(color).Render("●")
}

func formatTokens(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000)
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
