package provider

import "testing"

func TestGeminiName(t *testing.T) {
	p, err := Get("gemini")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "gemini" {
		t.Errorf("expected name 'gemini', got %q", p.Name())
	}
}

func TestGeminiParseState(t *testing.T) {
	p, _ := Get("gemini")
	state := p.ParseState("any content")
	if state != StateUnknown {
		t.Errorf("expected StateUnknown, got %v", state)
	}
}

func TestGeminiParsePermission(t *testing.T) {
	p, _ := Get("gemini")
	req := p.ParsePermission("any content")
	if req != nil {
		t.Error("expected nil permission")
	}
}

func TestGeminiLaunchCmd(t *testing.T) {
	p, _ := Get("gemini")

	tests := []struct {
		model  string
		prompt string
		want   string
	}{
		{"", "", "gemini"},
		{"", "do stuff", "gemini do stuff"},
		{"pro", "do stuff", "gemini do stuff"},
		{"", "", "gemini"},
	}

	for _, tt := range tests {
		got := p.LaunchCmd(tt.model, tt.prompt)
		if got != tt.want {
			t.Errorf("LaunchCmd(%q, %q) = %q, want %q", tt.model, tt.prompt, got, tt.want)
		}
	}
}

func TestGeminiCompactCmd(t *testing.T) {
	p, _ := Get("gemini")
	if p.CompactCmd() != "" {
		t.Errorf("expected empty compact cmd, got %q", p.CompactCmd())
	}
}

func TestGeminiApproveKey(t *testing.T) {
	p, _ := Get("gemini")
	if p.ApproveKey() != "y" {
		t.Errorf("expected 'y', got %q", p.ApproveKey())
	}
}

func TestGeminiDenyKey(t *testing.T) {
	p, _ := Get("gemini")
	if p.DenyKey() != "n" {
		t.Errorf("expected 'n', got %q", p.DenyKey())
	}
}

func TestGeminiParseCost(t *testing.T) {
	p, _ := Get("gemini")
	cost, tokIn, tokOut := p.ParseCost("any content")
	if cost != 0 || tokIn != 0 || tokOut != 0 {
		t.Errorf("expected zero cost, got %f, %d, %d", cost, tokIn, tokOut)
	}
}
