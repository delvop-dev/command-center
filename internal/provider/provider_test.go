package provider

import (
	"sort"
	"testing"
)

func TestAgentStateString(t *testing.T) {
	tests := []struct {
		state AgentState
		want  string
	}{
		{StateUnknown, "unknown"},
		{StateIdle, "idle"},
		{StateThinking, "thinking"},
		{StateEditing, "editing"},
		{StateRunningTool, "running_tool"},
		{StateWaitingPermission, "waiting_permission"},
		{StateWaitingInput, "waiting_input"},
		{StateCompacting, "compacting"},
		{StateError, "error"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("AgentState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestRegistryContainsProviders(t *testing.T) {
	names := List()
	sort.Strings(names)

	expected := []string{"claude", "codex", "gemini"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d providers, got %d: %v", len(expected), len(names), names)
	}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("expected provider %q at index %d, got %q", name, i, names[i])
		}
	}
}

func TestGetProvider(t *testing.T) {
	p, err := Get("claude")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "claude" {
		t.Errorf("expected name 'claude', got %q", p.Name())
	}

	_, err = Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent provider")
	}
}
