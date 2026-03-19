package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/delvop-dev/delvop/internal/config"
	"github.com/delvop-dev/delvop/internal/hooks"
	"github.com/delvop-dev/delvop/internal/notify"
	"github.com/delvop-dev/delvop/internal/session"
)

type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewFocused
)

type Model struct {
	cfg      *config.Config
	manager  *session.Manager
	hooks    *hooks.Engine
	notifier *notify.Router

	viewMode     ViewMode
	selectedIdx  int
	width        int
	height       int
	statusMsg    string
	statusExpiry time.Time
	showHelp     bool

	inputMode    bool
	inputPurpose string
	textInput    textinput.Model

	focusedID    string
	scrollOffset int
}

func NewModel(cfg *config.Config, mgr *session.Manager, hookEngine *hooks.Engine, notif *notify.Router) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter agent name..."
	ti.CharLimit = 64

	return Model{
		cfg:       cfg,
		manager:   mgr,
		hooks:     hookEngine,
		notifier:  notif,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(m.cfg.General.PollIntervalMs),
		listenForHookEvents(m.hooks),
	)
}

func tickCmd(intervalMs int) tea.Cmd {
	return tea.Tick(time.Duration(intervalMs)*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func listenForHookEvents(engine *hooks.Engine) tea.Cmd {
	if engine == nil {
		return nil
	}
	ch := engine.Subscribe()
	return func() tea.Msg {
		evt := <-ch
		return HookEventMsg{Event: evt}
	}
}
