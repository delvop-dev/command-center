package provider

import (
	"fmt"
	"os/exec"
	"sync"
)

// AgentState represents the current state of a coding agent.
type AgentState int

const (
	StateUnknown AgentState = iota
	StateIdle
	StateThinking
	StateEditing
	StateRunningTool
	StateWaitingPermission
	StateWorking
	StateWaitingInput
	StateCompacting
	StateError
	StatePreparing
)

// StateWaitingForPermission is an alias for StateWaitingPermission.
const StateWaitingForPermission = StateWaitingPermission

func (s AgentState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateThinking:
		return "thinking"
	case StateEditing:
		return "editing"
	case StateRunningTool:
		return "running_tool"
	case StateWaitingPermission:
		return "waiting_permission"
	case StateWorking:
		return "working"
	case StateWaitingInput:
		return "waiting_input"
	case StateCompacting:
		return "compacting"
	case StateError:
		return "error"
	case StatePreparing:
		return "preparing"
	default:
		return "unknown"
	}
}

// PermissionRequest represents a permission prompt from an agent.
type PermissionRequest struct {
	Tool        string
	Description string
	RawContent  string
}

// AgentProvider defines the interface for interacting with different coding agents.
type AgentProvider interface {
	// Name returns the provider identifier.
	Name() string

	// ParseState analyzes pane content and returns the current agent state.
	ParseState(paneContent string) AgentState

	// ParsePermission extracts permission request details from pane content.
	ParsePermission(paneContent string) *PermissionRequest

	// LaunchCmd returns the command to launch the agent with the given model and prompt.
	LaunchCmd(model, prompt string) string

	// CompactCmd returns the command/key sequence to trigger compaction.
	CompactCmd() string

	// ApproveKey returns the key to press to approve a permission request.
	ApproveKey() string

	// DenyKey returns the key to press to deny a permission request.
	DenyKey() string

	// ParseCost extracts cost and token usage from pane content.
	ParseCost(paneContent string) (costUSD float64, tokensIn int, tokensOut int)

	// BinaryName returns the CLI binary name (e.g. "claude", "codex", "gemini").
	BinaryName() string

	// InstallHint returns instructions for installing the agent CLI.
	InstallHint() string
}

// CheckInstalled verifies the provider's binary is available.
// It checks via a shell so that the user's profile PATH is used,
// matching how tmux will actually launch the command.
func CheckInstalled(p AgentProvider) error {
	// First try Go's LookPath (fast path).
	if _, err := exec.LookPath(p.BinaryName()); err == nil {
		return nil
	}
	// Fall back to shell-based lookup, which sources the user's profile
	// and sees the same PATH that tmux will use.
	cmd := exec.Command("sh", "-lc", "command -v "+p.BinaryName())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s not found in PATH. Install: %s", p.BinaryName(), p.InstallHint())
	}
	return nil
}

var (
	registry   = make(map[string]AgentProvider)
	registryMu sync.RWMutex
)

// Register adds a provider to the global registry.
func Register(p AgentProvider) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[p.Name()] = p
}

// Get retrieves a provider by name.
func Get(name string) (AgentProvider, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", name)
	}
	return p, nil
}

// List returns the names of all registered providers.
func List() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
