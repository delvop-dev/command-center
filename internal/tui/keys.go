package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Escape   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	New      key.Binding
	Template key.Binding
	Approve  key.Binding
	Deny     key.Binding
	Message  key.Binding
	Kill     key.Binding
	Compact  key.Binding
	Quit     key.Binding
	Help     key.Binding
}

var Keys = KeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "focus/zoom")),
	Escape:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
	New:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new agent")),
	Template: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "from template")),
	Approve:  key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "approve")),
	Deny:     key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "deny")),
	Message:  key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "message")),
	Kill:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "kill agent")),
	Compact:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compact")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}
