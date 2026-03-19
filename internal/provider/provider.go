package provider

import (
	"fmt"
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
