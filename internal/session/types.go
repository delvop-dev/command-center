package session

import (
	"time"

	"github.com/delvop-dev/delvop/internal/provider"
)

// Session represents a running coding agent session.
type Session struct {
	ID        string
	Name      string
	TmuxPane  string
	Provider  provider.AgentProvider
	State     provider.AgentState
	Model     string
	Prompt    string
	StartedAt time.Time
	WorkDir   string
	GitBranch string

	// Metrics
	TokensIn     int
	TokensOut    int
	CostUSD      float64
	LinesChanged int

	// Recent activity
	LastOutput   string
	FileChanges  []FileChange
	LastActivity time.Time
}

// FileChange represents a file modification detected in a session.
type FileChange struct {
	Path      string
	Operation string // "add", "modify", "delete"
	Timestamp time.Time
}

// Event represents something notable that happened in a session.
type Event struct {
	SessionID string
	Type      string // "state_change", "permission", "error", "file_change", "cost_update"
	Message   string
	Data      interface{}
	Timestamp time.Time
}

// KPIData holds key performance indicators across all sessions.
type KPIData struct {
	ActiveSessions   int
	WaitingPermission int
	TotalCostUSD     float64
	TotalTokensIn    int
	TotalTokensOut   int
	TotalLinesChanged int
}
