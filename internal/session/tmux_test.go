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

func TestAttachCmd(t *testing.T) {
	bridge := NewTmuxBridge("dv-")
	if got := bridge.AttachCmd("task1"); got != "tmux attach-session -t dv-task1" {
		t.Errorf("expected 'tmux attach-session -t dv-task1', got %q", got)
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
	err := bridge.CreateSession("integration", "sleep 30", "/tmp")
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
	content, err := bridge.CapturePaneContent("integration")
	if err != nil {
		t.Fatalf("failed to capture pane: %v", err)
	}
	// Content should be a string (may be empty for sleep command)
	_ = content

	// Kill session
	err = bridge.KillSession("integration")
	if err != nil {
		t.Fatalf("failed to kill session: %v", err)
	}
}
