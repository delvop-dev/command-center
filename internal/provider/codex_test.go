package provider

import "testing"

func TestCodexName(t *testing.T) {
	p, err := Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "codex" {
		t.Errorf("expected name 'codex', got %q", p.Name())
	}
}

func TestCodexParseState(t *testing.T) {
	p, _ := Get("codex")
	state := p.ParseState("any content")
	if state != StateUnknown {
		t.Errorf("expected StateUnknown, got %v", state)
	}
}

func TestCodexParsePermission(t *testing.T) {
	p, _ := Get("codex")
	req := p.ParsePermission("any content")
	if req != nil {
		t.Error("expected nil permission")
	}
}

func TestCodexLaunchCmd(t *testing.T) {
	p, _ := Get("codex")

	tests := []struct {
		model  string
		prompt string
		want   string
	}{
		{"", "", "codex"},
		{"", "do stuff", "codex do stuff"},
		{"gpt4", "do stuff", "codex do stuff"},
		{"", "", "codex"},
	}

	for _, tt := range tests {
		got := p.LaunchCmd(tt.model, tt.prompt)
		if got != tt.want {
			t.Errorf("LaunchCmd(%q, %q) = %q, want %q", tt.model, tt.prompt, got, tt.want)
		}
	}
}

func TestCodexCompactCmd(t *testing.T) {
	p, _ := Get("codex")
	if p.CompactCmd() != "" {
		t.Errorf("expected empty compact cmd, got %q", p.CompactCmd())
	}
}

func TestCodexApproveKey(t *testing.T) {
	p, _ := Get("codex")
	if p.ApproveKey() != "y" {
		t.Errorf("expected 'y', got %q", p.ApproveKey())
	}
}

func TestCodexDenyKey(t *testing.T) {
	p, _ := Get("codex")
	if p.DenyKey() != "n" {
		t.Errorf("expected 'n', got %q", p.DenyKey())
	}
}

func TestCodexParseCost(t *testing.T) {
	p, _ := Get("codex")
	cost, tokIn, tokOut := p.ParseCost("any content")
	if cost != 0 || tokIn != 0 || tokOut != 0 {
		t.Errorf("expected zero cost, got %f, %d, %d", cost, tokIn, tokOut)
	}
}
