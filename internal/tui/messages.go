package tui

import "github.com/delvop-dev/delvop/internal/hooks"

type TickMsg struct{}

type HookEventMsg struct {
	Event hooks.HookEvent
}

type SessionCreatedMsg struct {
	SessionID string
}

type SessionKilledMsg struct {
	SessionID string
}

type ErrorMsg struct {
	Err error
}

type StatusMsg struct {
	Message string
}
