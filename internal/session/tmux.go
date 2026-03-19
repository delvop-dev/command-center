package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// TmuxBridge handles all interactions with tmux.
type TmuxBridge struct {
	prefix string
}

// NewTmuxBridge creates a new TmuxBridge with the given session name prefix.
func NewTmuxBridge(prefix string) *TmuxBridge {
	return &TmuxBridge{prefix: prefix}
}

// SessionName returns the full tmux session name for a given ID.
func (t *TmuxBridge) SessionName(id string) string {
	return t.prefix + id
}

// CreateSession creates a new tmux session running the given command.
// The command is executed through a shell so that arguments are parsed correctly.
func (t *TmuxBridge) CreateSession(id, workDir, command string) error {
	name := t.SessionName(id)
	shell := "bash"
	if s := strings.TrimSpace(shellFromEnv()); s != "" {
		shell = s
	}
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-c", workDir,
		shell, "-lc", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux new-session failed: %w (output: %s)", err, string(output))
	}
	return nil
}

// KillSession terminates a tmux session by ID.
func (t *TmuxBridge) KillSession(id string) error {
	name := t.SessionName(id)
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

// CapturePaneContent captures the visible content of a tmux session's pane.
// If lines > 0, it captures that many lines of scrollback history.
func (t *TmuxBridge) CapturePaneContent(id string, lines int) (string, error) {
	name := t.SessionName(id)
	args := []string{"capture-pane", "-t", name, "-p"}
	if lines > 0 {
		args = append(args, "-S", fmt.Sprintf("-%d", lines))
	}
	cmd := exec.Command("tmux", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("capture pane %s: %w", name, err)
	}
	return string(out), nil
}

// SendKeys sends keystrokes to a tmux session, followed by Enter.
func (t *TmuxBridge) SendKeys(id, keys string) error {
	name := t.SessionName(id)
	cmd := exec.Command("tmux", "send-keys", "-t", name, keys, "Enter")
	return cmd.Run()
}

// SendRawKey sends a raw key to a tmux session without appending Enter.
func (t *TmuxBridge) SendRawKey(id, key string) error {
	name := t.SessionName(id)
	cmd := exec.Command("tmux", "send-keys", "-t", name, key)
	return cmd.Run()
}

// ListSessions returns a list of tmux session names that match the prefix.
func (t *TmuxBridge) ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		// tmux returns error if no sessions exist
		if strings.Contains(err.Error(), "no server running") ||
			strings.Contains(err.Error(), "exit status") {
			return nil, nil
		}
		return nil, err
	}

	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, t.prefix) {
			sessions = append(sessions, strings.TrimPrefix(line, t.prefix))
		}
	}
	return sessions, nil
}

// AttachCmd returns the command string to attach to a tmux session.
func (t *TmuxBridge) AttachCmd(id string) string {
	name := t.SessionName(id)
	return fmt.Sprintf("tmux attach-session -t %s", name)
}

// shellFromEnv returns the user's preferred shell from $SHELL.
func shellFromEnv() string {
	return os.Getenv("SHELL")
}
