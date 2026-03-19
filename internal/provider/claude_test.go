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
			name:    "waiting input suffix >",
			content: "Ready\n>",
			want:    StateWaitingInput,
		},
		{
			name:    "waiting input prefix >",
			content: "Ready\n> ",
			want:    StateWaitingInput,
		},
		{
			name:    "waiting input with trailing >",
			content: "Ready\nprompt>",
			want:    StateWaitingInput,
		},
		{
			name:    "compacting",
			content: "Compacting conversation...",
			want:    StateCompacting,
		},
		{
			name:    "compacting lowercase",
			content: "compacting context",
			want:    StateCompacting,
		},
		{
			name:    "running tool",
			content: "Running bash command",
			want:    StateRunningTool,
		},
		{
			name:    "executing tool",
			content: "Executing script",
			want:    StateRunningTool,
		},
		{
			name:    "editing",
			content: "Editing main.go",
			want:    StateEditing,
		},
		{
			name:    "writing file",
			content: "Writing config.toml",
			want:    StateEditing,
		},
		{
			name:    "idle with dollar",
			content: "$ ",
			want:    StateIdle,
		},
		{
			name:    "idle with Claude",
			content: "Claude Code",
			want:    StateIdle,
		},
		{
			name:    "empty content",
			content: "",
			want:    StateUnknown,
		},
		{
			name:    "only whitespace lines",
			content: "\n\n   \n",
			want:    StateUnknown,
		},
		{
			name:    "ellipsis indicates thinking",
			content: "Some output\nProcessing...",
			want:    StateThinking,
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
	if req.RawContent != "output\nAllow Read?" {
		t.Errorf("expected raw content preserved")
	}
	if req.Description == "" {
		t.Error("expected non-empty description")
	}

	// Multiple Allow lines - should get the last one
	req = p.ParsePermission("Allow OldTool?\nAllow Write?")
	if req == nil {
		t.Fatal("expected permission request")
	}
	if req.Tool != "Write" {
		t.Errorf("expected tool 'Write', got %q", req.Tool)
	}

	req = p.ParsePermission("no permission here")
	if req != nil {
		t.Error("expected nil for no permission")
	}
}

func TestClaudeParsePermissionExtractTool(t *testing.T) {
	p, _ := Get("claude")

	tests := []struct {
		content  string
		wantTool string
	}{
		{"Allow Read?", "Read"},
		{"Allow Write?", "Write"},
		{"Allow Bash?", "Bash"},
		{"Some Allow text?", "text"},
	}

	for _, tt := range tests {
		req := p.ParsePermission(tt.content)
		if req == nil {
			t.Fatalf("expected permission for %q", tt.content)
		}
		if req.Tool != tt.wantTool {
			t.Errorf("ParsePermission(%q): expected tool %q, got %q", tt.content, tt.wantTool, req.Tool)
		}
	}
}

func TestClaudeLaunchCmd(t *testing.T) {
	p, _ := Get("claude")

	tests := []struct {
		model  string
		prompt string
		want   string
	}{
		{"opus", "fix the bug", `claude --model opus -p "fix the bug"`},
		{"", "", "claude"},
		{"sonnet", "", "claude --model sonnet"},
		{"", "do stuff", `claude -p "do stuff"`},
	}

	for _, tt := range tests {
		got := p.LaunchCmd(tt.model, tt.prompt)
		if got != tt.want {
			t.Errorf("LaunchCmd(%q, %q) = %q, want %q", tt.model, tt.prompt, got, tt.want)
		}
	}
}

func TestClaudeCompactCmd(t *testing.T) {
	p, _ := Get("claude")
	if p.CompactCmd() != "/compact" {
		t.Errorf("unexpected compact cmd: %s", p.CompactCmd())
	}
}

func TestClaudeApproveKey(t *testing.T) {
	p, _ := Get("claude")
	if p.ApproveKey() != "y" {
		t.Errorf("unexpected approve key: %s", p.ApproveKey())
	}
}

func TestClaudeDenyKey(t *testing.T) {
	p, _ := Get("claude")
	if p.DenyKey() != "n" {
		t.Errorf("unexpected deny key: %s", p.DenyKey())
	}
}

func TestClaudeParseCost(t *testing.T) {
	p, _ := Get("claude")
	cost, tokIn, tokOut := p.ParseCost("some output")
	if cost != 0 || tokIn != 0 || tokOut != 0 {
		t.Errorf("expected zero cost, got %f, %d, %d", cost, tokIn, tokOut)
	}
}
