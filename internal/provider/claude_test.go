package provider

import (
	"testing"
)

func TestClaudeParseState(t *testing.T) {
	p, _ := Get("claude")

	tests := []struct {
		name    string
		content string
		want    AgentState
	}{
		{
			name:    "permission prompt",
			content: "Some output\nAllow Read file.go?",
			want:    StateWaitingPermission,
		},
		{
			name:    "proceed prompt",
			content: "Some output\nDo you want to proceed?",
			want:    StateWaitingPermission,
		},
		{
			name:    "thinking",
			content: "Processing\nThinking...",
			want:    StateThinking,
		},
		{
			name:    "waiting input",
			content: "Ready\n>",
			want:    StateWaitingInput,
		},
		{
			name:    "compacting",
			content: "Compacting conversation...",
			want:    StateCompacting,
		},
		{
			name:    "running tool",
			content: "Running bash command",
			want:    StateRunningTool,
		},
		{
			name:    "editing",
			content: "Editing main.go",
			want:    StateEditing,
		},
		{
			name:    "empty content",
			content: "",
			want:    StateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ParseState(tt.content)
			if got != tt.want {
				t.Errorf("ParseState(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestClaudeParsePermission(t *testing.T) {
	p, _ := Get("claude")

	req := p.ParsePermission("output\nAllow Read?")
	if req == nil {
		t.Fatal("expected permission request")
	}
	if req.Tool != "Read" {
		t.Errorf("expected tool 'Read', got %q", req.Tool)
	}

	req = p.ParsePermission("no permission here")
	if req != nil {
		t.Error("expected nil for no permission")
	}
}

func TestClaudeLaunchCmd(t *testing.T) {
	p, _ := Get("claude")

	cmd := p.LaunchCmd("opus", "fix the bug")
	if cmd != `claude --model opus -p "fix the bug"` {
		t.Errorf("unexpected launch cmd: %s", cmd)
	}

	cmd = p.LaunchCmd("", "")
	if cmd != "claude" {
		t.Errorf("unexpected launch cmd: %s", cmd)
	}
}

func TestClaudeCompactCmd(t *testing.T) {
	p, _ := Get("claude")
	if p.CompactCmd() != "/compact" {
		t.Errorf("unexpected compact cmd: %s", p.CompactCmd())
	}
}
