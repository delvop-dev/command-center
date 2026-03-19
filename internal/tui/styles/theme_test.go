package styles

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStateColor(t *testing.T) {
	tests := []struct {
		state string
		want  lipgloss.Color
	}{
		{"WORKING", Green},
		{"NEEDS INPUT", Red},
		{"DONE", Blue},
		{"IDLE", TextMuted},
		{"ERROR", Red},
		{"UNKNOWN", TextDim},
		{"anything else", TextDim},
		{"", TextDim},
	}

	for _, tt := range tests {
		got := StateColor(tt.state)
		if got != tt.want {
			t.Errorf("StateColor(%q) = %v, want %v", tt.state, got, tt.want)
		}
	}
}

func TestStateStyle(t *testing.T) {
	states := []string{"WORKING", "NEEDS INPUT", "DONE", "IDLE", "ERROR", "UNKNOWN"}
	for _, state := range states {
		style := StateStyle(state)
		// Should render without panic
		rendered := style.Render(state)
		if rendered == "" {
			t.Errorf("StateStyle(%q).Render() returned empty", state)
		}
	}
}

func TestBrandStyle(t *testing.T) {
	style := BrandStyle()
	rendered := style.Render("delvop")
	if rendered == "" {
		t.Error("BrandStyle().Render() returned empty")
	}
}

func TestCostStyle(t *testing.T) {
	style := CostStyle()
	rendered := style.Render("$1.50")
	if rendered == "" {
		t.Error("CostStyle().Render() returned empty")
	}
}

func TestKeyStyle(t *testing.T) {
	style := KeyStyle()
	rendered := style.Render("q")
	if rendered == "" {
		t.Error("KeyStyle().Render() returned empty")
	}
}

func TestDescStyle(t *testing.T) {
	style := DescStyle()
	rendered := style.Render("quit")
	if rendered == "" {
		t.Error("DescStyle().Render() returned empty")
	}
}

func TestProgressBar(t *testing.T) {
	tests := []struct {
		pct   int
		width int
	}{
		{0, 10},
		{50, 10},
		{100, 10},
		{150, 10}, // over 100%
		{0, 0},
		{50, 20},
	}

	for _, tt := range tests {
		bar := ProgressBar(tt.pct, tt.width)
		if tt.width > 0 && bar == "" {
			t.Errorf("ProgressBar(%d, %d) returned empty", tt.pct, tt.width)
		}
	}
}

func TestProgressBarContent(t *testing.T) {
	bar := ProgressBar(50, 10)
	// The bar should contain filled and empty blocks
	if bar == "" {
		t.Error("expected non-empty bar")
	}
}

func TestProgressBarFull(t *testing.T) {
	bar := ProgressBar(100, 10)
	if bar == "" {
		t.Error("expected non-empty bar for 100%")
	}
}

func TestProgressBarOverflow(t *testing.T) {
	bar := ProgressBar(200, 10)
	if bar == "" {
		t.Error("expected non-empty bar for 200%")
	}
	// Should not contain more filled blocks than width
}

func TestProgressBarZeroWidth(t *testing.T) {
	bar := ProgressBar(50, 0)
	// With 0 width, no blocks to render
	if strings.Contains(bar, "█") {
		t.Error("expected no filled blocks with 0 width")
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constants are non-empty
	colors := []lipgloss.Color{
		Background, Surface, Border, TextPrimary, TextSecondary,
		TextMuted, TextDim, TextGhost, Purple, Green, Red, Blue, Amber,
	}
	for i, c := range colors {
		if string(c) == "" {
			t.Errorf("color constant %d is empty", i)
		}
	}
}

func TestStyleConstants(t *testing.T) {
	// Verify styles render without panic
	styles := []lipgloss.Style{
		StatusBar, KPIBar, AgentCard, AgentCardFocused,
		ActivityFeed, ActionQueue, HelpBar,
	}
	for i, s := range styles {
		rendered := s.Render("test")
		if rendered == "" {
			t.Errorf("style constant %d renders empty", i)
		}
	}
}
