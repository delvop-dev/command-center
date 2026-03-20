package session

import (
	"time"

	"github.com/delvop-dev/delvop/internal/provider"
)

// Session represents a running coding agent session.
type Session struct {
	ID           string
	Name         string
	TmuxSession  string
	TmuxPane     string
	Provider     provider.AgentProvider
	ProviderName string
	State        provider.AgentState
	Model        string
	Prompt       string
	WorkDir      string
	Branch       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    time.Time

	// Metrics
	TokensIn     int
	TokensOut    int
	CostUSD      float64
	LinesChanged int

	// Tracks whether agent has done real work (thinking/editing/tools)
	HasWorked bool

	// Recent activity
	LastOutput string
	PaneContent  string
	FileChanges  []FileChange
	LastActivity time.Time
	Events       []Event
	Permission   *provider.PermissionRequest
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
	Time      time.Time
	Type      string // "state_change", "permission", "error", "file_change", "cost_update"
	Message   string
	Data      interface{}
	Timestamp time.Time
}

// KPIData holds key performance indicators across all sessions.
type KPIData struct {
	TotalCount        int
	ActiveCount       int
	WaitingPermission int
	TotalCost         float64
	TotalTokens       int
	TotalCostUSD      float64
	TotalTokensIn     int
	TotalTokensOut    int
	TotalLinesChanged int
	Uptime            time.Duration
	ActiveSessions    int
}
