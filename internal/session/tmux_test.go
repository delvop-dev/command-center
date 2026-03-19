package session

import (
	"os/exec"
	"testing"
)

func TestSessionName(t *testing.T) {
	bridge := NewTmuxBridge("dv-")
	if got := bridge.SessionName("task1"); got != "dv-task1" {
		t.Errorf("expected 'dv-task1', got %q", got)
	}
}

func TestSessionNameEmpty(t *testing.T) {
	bridge := NewTmuxBridge("")
	if got := bridge.SessionName("task1"); got != "task1" {
		t.Errorf("expected 'task1', got %q", got)
	}
}

func TestAttachCmd(t *testing.T) {
	bridge := NewTmuxBridge("dv-")
	if got := bridge.AttachCmd("task1"); got != "tmux attach-session -t dv-task1" {
		t.Errorf("expected 'tmux attach-session -t dv-task1', got %q", got)
	}
}

func TestNewTmuxBridge(t *testing.T) {
	bridge := NewTmuxBridge("test-")
	if bridge == nil {
		t.Fatal("expected non-nil bridge")
	}
	if bridge.prefix != "test-" {
		t.Errorf("expected prefix 'test-', got %q", bridge.prefix)
	}
}

func tmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestTmuxIntegration(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not available")
	}

	bridge := NewTmuxBridge("dv-test-")

	// Create a session
	err := bridge.CreateSession("integration", "/tmp", "sleep 30")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	defer bridge.KillSession("integration")

	// List sessions
	sessions, err := bridge.ListSessions()
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}

	found := false
	for _, s := range sessions {
		if s == "integration" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find session 'integration' in %v", sessions)
	}

	// Capture pane content
	content, err := bridge.CapturePaneContent("integration", 0)
	if err != nil {
		t.Fatalf("failed to capture pane: %v", err)
	}
	_ = content

	// Capture with lines
	content2, err := bridge.CapturePaneContent("integration", 30)
	if err != nil {
		t.Fatalf("failed to capture pane with lines: %v", err)
	}
	_ = content2

	// Send keys
	err = bridge.SendKeys("integration", "echo hello")
	if err != nil {
		t.Fatalf("failed to send keys: %v", err)
	}

	// Send raw key
	err = bridge.SendRawKey("integration", "y")
	if err != nil {
		t.Fatalf("failed to send raw key: %v", err)
	}

	// Kill session
	err = bridge.KillSession("integration")
	if err != nil {
		t.Fatalf("failed to kill session: %v", err)
	}
}

func TestListSessionsNoTmux(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not available")
	}
	bridge := NewTmuxBridge("dv-nonexistent-prefix-")
	sessions, err := bridge.ListSessions()
	if err != nil {
		// Error is OK if no tmux sessions exist
		return
	}
	// Should return empty list for non-matching prefix
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions with non-matching prefix, got %d", len(sessions))
	}
}
