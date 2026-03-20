package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/delvop-dev/delvop/internal/notify"
	"github.com/delvop-dev/delvop/internal/provider"
	"github.com/delvop-dev/delvop/internal/templates"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		m.manager.PollAll()
		for _, s := range m.manager.NeedsAttention() {
			m.notifier.Notify(s.ID, notify.EventInputNeeded,
				fmt.Sprintf("%s needs permission", s.Name))
		}
		return m, tickCmd(m.cfg.General.PollIntervalMs)

	case HookEventMsg:
		return m, listenForHookEvents(m.hooks)

	case ErrorMsg:
		m.statusMsg = fmt.Sprintf("Error: %v", msg.Err)
		m.statusExpiry = time.Now().Add(5 * time.Second)
		return m, nil

	case StatusMsg:
		m.statusMsg = msg.Message
		m.statusExpiry = time.Now().Add(3 * time.Second)
		return m, nil

	case SessionCreatedMsg:
		m.statusMsg = "Agent created"
		m.statusExpiry = time.Now().Add(3 * time.Second)
		return m, nil

	case tea.KeyMsg:
		if m.inputMode {
			return m.handleInputKey(msg)
		}
		switch m.viewMode {
		case ViewDashboard:
			return m.handleDashboardKey(msg)
		case ViewFocused:
			return m.handleFocusedKey(msg)
		}
	}
	return m, nil
}

func (m Model) handleDashboardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	sessions := m.manager.All()

	switch {
	case key.Matches(msg, Keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, Keys.Help):
		m.showHelp = !m.showHelp
		return m, nil
	case key.Matches(msg, Keys.Up):
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	case key.Matches(msg, Keys.Down):
		if m.selectedIdx < len(sessions)-1 {
			m.selectedIdx++
		}
		return m, nil
	case key.Matches(msg, Keys.Enter):
		if len(sessions) > 0 && m.selectedIdx < len(sessions) {
			s := sessions[m.selectedIdx]
			tmuxName := m.manager.TmuxSessionName(s.ID)
			if tmuxName == "" {
				return m, nil
			}
			// Verify session exists before attaching
			if err := exec.Command("tmux", "has-session", "-t", tmuxName).Run(); err != nil {
				m.statusMsg = fmt.Sprintf("Session %s not running — try restarting", s.Name)
				m.statusExpiry = time.Now().Add(3 * time.Second)
				return m, nil
			}
			m.focusedID = s.ID
			// Bind ctrl+\ to detach (single keypress, no prefix needed)
			_ = exec.Command("tmux", "bind-key", "-n", "C-\\", "detach-client").Run()
			// Hide tmux status bar — keep the terminal clean
			_ = exec.Command("tmux", "set-option", "-t", tmuxName, "status", "off").Run()
			c := exec.Command("tmux", "attach-session", "-t", tmuxName)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				if err != nil {
					return StatusMsg{Message: "Detached — back in delvop"}
				}
				return StatusMsg{Message: "Detached — back in delvop"}
			})
		}
		return m, nil
	case key.Matches(msg, Keys.New):
		m.inputMode = true
		m.inputPurpose = "new_agent"
		m.textInput.Placeholder = "Agent name (e.g. frontend, api-server)..."
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, m.textInput.Cursor.BlinkCmd()
	case key.Matches(msg, Keys.Approve):
		if len(sessions) > 0 && m.selectedIdx < len(sessions) {
			s := sessions[m.selectedIdx]
			if s.State == provider.StateWaitingForPermission {
				m.manager.Approve(s.ID)
				m.statusMsg = fmt.Sprintf("Approved %s", s.Name)
				m.statusExpiry = time.Now().Add(2 * time.Second)
			}
		}
		return m, nil
	case key.Matches(msg, Keys.Deny):
		if len(sessions) > 0 && m.selectedIdx < len(sessions) {
			s := sessions[m.selectedIdx]
			if s.State == provider.StateWaitingForPermission {
				m.manager.Deny(s.ID)
				m.statusMsg = fmt.Sprintf("Denied %s", s.Name)
				m.statusExpiry = time.Now().Add(2 * time.Second)
			}
		}
		return m, nil
	case key.Matches(msg, Keys.Message):
		if len(sessions) > 0 && m.selectedIdx < len(sessions) {
			m.inputMode = true
			m.inputPurpose = "message"
			m.textInput.Placeholder = "Message to send..."
			m.textInput.SetValue("")
			m.textInput.Focus()
			return m, m.textInput.Cursor.BlinkCmd()
		}
		return m, nil
	case key.Matches(msg, Keys.Kill):
		if len(sessions) > 0 && m.selectedIdx < len(sessions) {
			s := sessions[m.selectedIdx]
			m.manager.Remove(s.ID)
			if m.selectedIdx >= len(m.manager.All()) && m.selectedIdx > 0 {
				m.selectedIdx--
			}
			m.statusMsg = fmt.Sprintf("Killed %s", s.Name)
			m.statusExpiry = time.Now().Add(2 * time.Second)
		}
		return m, nil
	case key.Matches(msg, Keys.Template):
		tmpls, err := templates.LoadBuiltins()
		if err != nil || len(tmpls) == 0 {
			m.statusMsg = "No templates available"
			m.statusExpiry = time.Now().Add(2 * time.Second)
			return m, nil
		}
		// Show template names as options
		var names []string
		for _, t := range tmpls {
			names = append(names, fmt.Sprintf("%s (%s)", t.Name, t.Description))
		}
		m.inputMode = true
		m.inputPurpose = "template"
		m.textInput.Placeholder = fmt.Sprintf("Template: %s", strings.Join(names, ", "))
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, m.textInput.Cursor.BlinkCmd()
	case key.Matches(msg, Keys.Compact):
		if len(sessions) > 0 && m.selectedIdx < len(sessions) {
			s := sessions[m.selectedIdx]
			if err := m.manager.Compact(s.ID); err != nil {
				m.statusMsg = fmt.Sprintf("Compact error: %v", err)
			} else {
				m.statusMsg = fmt.Sprintf("Compacting %s", s.Name)
			}
			m.statusExpiry = time.Now().Add(2 * time.Second)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleFocusedKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Escape):
		m.viewMode = ViewDashboard
		m.notifier.SetFocused("")
		return m, nil
	case key.Matches(msg, Keys.Enter):
		tmuxName := m.manager.TmuxSessionName(m.focusedID)
		if tmuxName != "" {
			c := exec.Command("tmux", "attach-session", "-t", tmuxName)
			return m, tea.ExecProcess(c, func(err error) tea.Msg {
				if err != nil {
					return ErrorMsg{Err: err}
				}
				return StatusMsg{Message: "Detached — back in delvop"}
			})
		}
		return m, nil
	case key.Matches(msg, Keys.Up):
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil
	case key.Matches(msg, Keys.Down):
		m.scrollOffset++
		return m, nil
	case key.Matches(msg, Keys.Approve):
		m.manager.Approve(m.focusedID)
		return m, nil
	case key.Matches(msg, Keys.Deny):
		m.manager.Deny(m.focusedID)
		return m, nil
	case key.Matches(msg, Keys.Message):
		m.inputMode = true
		m.inputPurpose = "message"
		m.textInput.Placeholder = "Message to send..."
		m.textInput.SetValue("")
		m.textInput.Focus()
		return m, m.textInput.Cursor.BlinkCmd()
	case key.Matches(msg, Keys.Kill):
		m.manager.Remove(m.focusedID)
		m.viewMode = ViewDashboard
		return m, nil
	case key.Matches(msg, Keys.Tab):
		sessions := m.manager.All()
		for i, s := range sessions {
			if s.ID == m.focusedID {
				next := (i + 1) % len(sessions)
				m.focusedID = sessions[next].ID
				m.scrollOffset = 0
				break
			}
		}
		return m, nil
	case key.Matches(msg, Keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		value := m.textInput.Value()
		m.inputMode = false
		m.textInput.Blur()
		if value == "" {
			return m, nil
		}
		switch m.inputPurpose {
		case "new_agent":
			return m, m.createAgent(value)
		case "template":
			return m, m.launchTemplate(value)
		case "message":
			sessions := m.manager.All()
			var targetID string
			if m.viewMode == ViewFocused {
				targetID = m.focusedID
			} else if m.selectedIdx < len(sessions) {
				targetID = sessions[m.selectedIdx].ID
			}
			if targetID != "" {
				m.manager.SendKeys(targetID, value)
				m.statusMsg = "Sent message to agent"
				m.statusExpiry = time.Now().Add(2 * time.Second)
			}
		}
		return m, nil
	case "esc":
		m.inputMode = false
		m.textInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) launchTemplate(name string) tea.Cmd {
	return func() tea.Msg {
		tmpls, err := templates.LoadBuiltins()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		// Find matching template by name prefix
		var matched *templates.Template
		lower := strings.ToLower(strings.TrimSpace(name))
		for _, t := range tmpls {
			if strings.ToLower(t.Name) == lower || strings.HasPrefix(strings.ToLower(t.Name), lower) {
				matched = t
				break
			}
		}
		if matched == nil {
			return ErrorMsg{Err: fmt.Errorf("template %q not found", name)}
		}
		// Launch all sessions from the template
		var launched int
		for _, st := range matched.Sessions {
			p, err := provider.Get(st.Provider)
			if err != nil {
				p, _ = provider.Get(m.cfg.General.DefaultProvider)
			}
			if p == nil {
				continue
			}
			if err := provider.CheckInstalled(p); err != nil {
				return ErrorMsg{Err: err}
			}
			model := st.Model
			if model == "" {
				model = m.cfg.General.DefaultModel
			}
			workDir, _ := os.Getwd()
			sess, err := m.manager.Add(st.Name, p, model, workDir, "")
			if err != nil {
				continue
			}
			if err := m.manager.LaunchWithPrompt(sess, st.InitialPrompt); err != nil {
				continue
			}
			launched++
		}
		return StatusMsg{Message: fmt.Sprintf("Launched template: %s", matched.Name)}
	}
}

func (m Model) createAgent(name string) tea.Cmd {
	return func() tea.Msg {
		p, err := provider.Get(m.cfg.General.DefaultProvider)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		if err := provider.CheckInstalled(p); err != nil {
			return ErrorMsg{Err: err}
		}
		workDir, err := os.Getwd()
		if err != nil {
			workDir = "."
		}
		sess, err := m.manager.Add(name, p, m.cfg.General.DefaultModel, workDir, "")
		if err != nil {
			return ErrorMsg{Err: err}
		}
		if err := m.manager.Launch(sess); err != nil {
			return ErrorMsg{Err: fmt.Errorf("launch failed: %w", err)}
		}
		return SessionCreatedMsg{SessionID: sess.ID}
	}
}
