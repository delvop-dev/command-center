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
		// Permission states
		{
			name:    "simple allow prompt",
			content: "Some output\nAllow Read file.go?",
			want:    StateWaitingPermission,
		},
		{
			name:    "proceed prompt",
			content: "Some output\nDo you want to proceed?",
			want:    StateWaitingPermission,
		},
		{
			name: "full permission dialog with numbered choices",
			content: `────────────────────────────────────────────────────────────────────────────────
 Bash command

   ls ~/Desktop/folio 2>/dev/null; echo "EXIT: $?"
   Check if target directory exists

 Do you want to proceed?
 ❯ 1. Yes
   2. Yes, allow reading from Desktop/ from this project
   3. No

 Esc to cancel · Tab to amend · ctrl+e to explain`,
			want: StateWaitingPermission,
		},

		// Input prompt states
		{
			name:    "bare ❯ prompt with separators",
			content: "some output\n────────────\n❯ \n────────────\n  ctrl+t to hide tasks",
			want:    StateWaitingInput,
		},
		{
			name:    "❯ with user typed text",
			content: "────────────\n❯ fix the bug\n────────────",
			want:    StateWaitingInput,
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
			name:    "prompt trailing >",
			content: "Ready\nprompt>",
			want:    StateWaitingInput,
		},

		// Thinking states
		{
			name:    "thinking uppercase",
			content: "Processing\nThinking...",
			want:    StateThinking,
		},
		{
			name:    "thinking lowercase",
			content: "Some output\nthinking...",
			want:    StateThinking,
		},
		{
			name:    "worked for timer",
			content: "✻ Worked for 37s",
			want:    StateThinking,
		},
		{
			name:    "skill loading",
			content: "⏺ Skill(superpowers:brainstorming)\n  ⎿  Successfully loaded skill",
			want:    StateThinking,
		},

		// Working states
		{
			name:    "tasks in progress",
			content: "  8 tasks (1 done, 1 in progress, 6 open)\n  ◼ Build frontend",
			want:    StateWorking,
		},
		{
			name:    "in-progress task marker only",
			content: "  ◼ Build frontend\n  ◻ Deploy",
			want:    StateWorking,
		},

		// Tool execution
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
			name:    "reading file",
			content: "Reading 1 file… (ctrl+o to expand)",
			want:    StateRunningTool,
		},
		{
			name:    "searching",
			content: "Searching for matches...",
			want:    StateRunningTool,
		},

		// Editing
		{
			name:    "editing file",
			content: "Editing main.go",
			want:    StateEditing,
		},
		{
			name:    "writing file",
			content: "Writing config.toml",
			want:    StateEditing,
		},

		// Compacting
		{
			name:    "compacting uppercase",
			content: "Compacting conversation...",
			want:    StateCompacting,
		},
		{
			name:    "compacting lowercase",
			content: "compacting context",
			want:    StateCompacting,
		},

		// Idle
		{
			name:    "idle with Claude Code",
			content: "Claude Code v2.1.79",
			want:    StateIdle,
		},

		// Unknown
		{
			name:    "empty content",
			content: "",
			want:    StateUnknown,
		},
		{
			name:    "only whitespace",
			content: "\n\n   \n",
			want:    StateUnknown,
		},
		{
			name:    "only decorative lines",
			content: "────────────\n────────────",
			want:    StateUnknown,
		},

		// Real-world composite: prompt visible after task completion
		{
			name: "worked for + prompt",
			content: `✻ Worked for 37s

  8 tasks (1 done, 1 in progress, 6 open)
────────────
❯
────────────
  ctrl+t to hide tasks`,
			want: StateWaitingInput,
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

func TestClaudeParsePermissionProceedDialog(t *testing.T) {
	p, _ := Get("claude")

	content := `────────────────────────────────────────────────────────────────────────────────
 Bash command

   ls ~/Desktop/folio 2>/dev/null; echo "EXIT: $?"
   Check if target directory exists

 Do you want to proceed?
 ❯ 1. Yes
   2. Yes, allow reading from Desktop/
   3. No`

	req := p.ParsePermission(content)
	if req == nil {
		t.Fatal("expected permission request for proceed dialog")
	}
	if req.Tool != "Bash" {
		t.Errorf("expected tool 'Bash', got %q", req.Tool)
	}
	if req.Description == "" || req.Description == "Do you want to proceed?" {
		t.Errorf("expected meaningful description, got %q", req.Description)
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
		{"opus", "fix the bug", "claude --model opus"},
		{"", "", "claude"},
		{"sonnet", "", "claude --model sonnet"},
		{"", "do stuff", "claude"},
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
	if p.ApproveKey() != "Enter" {
		t.Errorf("unexpected approve key: %s", p.ApproveKey())
	}
}

func TestClaudeDenyKey(t *testing.T) {
	p, _ := Get("claude")
	if p.DenyKey() != "Escape" {
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
