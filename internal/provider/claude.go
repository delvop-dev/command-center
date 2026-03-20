package provider

import (
	"fmt"
	"strings"
)

type claudeProvider struct{}

func init() {
	Register(&claudeProvider{})
}

func (c *claudeProvider) Name() string {
	return "claude"
}

// isChrome returns true if a line is purely decorative UI chrome.
func isChrome(line string) bool {
	for _, r := range line {
		switch {
		case r >= 0x2500 && r <= 0x257F: // Box Drawing
		case r >= 0x2580 && r <= 0x259F: // Block Elements
		case r == ' ', r == '-', r == '=':
		default:
			return false
		}
	}
	return true
}

// isHelperText returns true for Claude Code UI hint lines.
func isHelperText(line string) bool {
	lower := strings.ToLower(line)
	return strings.HasPrefix(lower, "esc to cancel") ||
		strings.HasPrefix(lower, "? for shortcuts") ||
		strings.HasPrefix(lower, "ctrl+") ||
		strings.HasPrefix(lower, "tab to amend") ||
		strings.HasPrefix(lower, "enter to confirm")
}

func (c *claudeProvider) ParseState(paneContent string) AgentState {
	if paneContent == "" {
		return StateUnknown
	}

	// Phase 1: Broad content scan for permission dialogs.
	if strings.Contains(paneContent, "Do you want to proceed?") {
		return StateWaitingPermission
	}
	if strings.Contains(paneContent, "I trust this folder") {
		return StateWaitingPermission
	}

	// Phase 2: Two-pass line scan.
	// Pass A: Check ONLY the last 3 meaningful lines for prompt/input state.
	// Pass B: Scan deeper (up to 40 lines) for working indicators.
	// This prevents old ❯ prompts in scrollback from masking working state.

	lines := strings.Split(paneContent, "\n")

	// --- Pass A: prompt detection (last 3 meaningful lines only) ---
	promptState := StateUnknown
	checked := 0
	for i := len(lines) - 1; i >= 0 && checked < 3; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || isChrome(line) || isHelperText(line) {
			continue
		}
		checked++

		if strings.Contains(line, "Allow") && strings.Contains(line, "?") {
			return StateWaitingPermission
		}
		if strings.HasPrefix(line, "❯") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "❯"))
			if rest == "" || (len(rest) > 0 && (rest[0] < '0' || rest[0] > '9')) {
				promptState = StateWaitingInput
			}
		}
		if line == ">" || strings.HasSuffix(line, ">") {
			promptState = StateWaitingInput
		}
	}

	// If pass A found a prompt at the bottom, that's the current state.
	// The agent has finished working and is waiting for input.
	if promptState != StateUnknown {
		return promptState
	}

	// --- Pass B: working state detection (scan deeper) ---
	checked = 0
	for i := len(lines) - 1; i >= 0 && checked < 40; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || isChrome(line) || isHelperText(line) {
			continue
		}
		checked++

		// Compacting
		if strings.Contains(line, "Compacting") || strings.Contains(line, "compacting") {
			return StateCompacting
		}

		// Thinking / active processing
		if strings.Contains(line, "Thinking") || strings.Contains(line, "thinking") {
			return StateThinking
		}
		if strings.Contains(line, "Worked for") {
			return StateThinking
		}

		// Task progress
		if strings.Contains(line, "tasks (") {
			return StateWorking
		}
		if strings.HasPrefix(line, "◼") {
			return StateWorking
		}

		// Tool execution
		if strings.HasPrefix(line, "Reading") || strings.HasPrefix(line, "Searching") {
			return StateRunningTool
		}
		if strings.HasPrefix(line, "Running") || strings.HasPrefix(line, "Executing") {
			return StateRunningTool
		}

		// Editing
		if strings.HasPrefix(line, "Editing") || strings.HasPrefix(line, "Writing") {
			return StateEditing
		}

		// Skill loading
		if strings.Contains(line, "Skill(") || strings.Contains(line, "loaded skill") {
			return StateThinking
		}

		// Active output indicators
		if strings.HasPrefix(line, "⏺") {
			return StateWorking
		}
		if strings.HasPrefix(line, "⎿") {
			return StateWorking
		}

		// Idle (welcome screen)
		if strings.Contains(line, "Claude Code") {
			// If prompt was found in pass A, it takes priority over idle
			if promptState != StateUnknown {
				return promptState
			}
			return StateIdle
		}
	}

	// If we found a prompt in pass A but no working indicators in pass B
	if promptState != StateUnknown {
		return promptState
	}

	return StateUnknown
}

func (c *claudeProvider) ParsePermission(paneContent string) *PermissionRequest {
	lines := strings.Split(paneContent, "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		// Pattern 1: "Allow <tool>?"
		if strings.Contains(line, "Allow") && strings.Contains(line, "?") {
			return &PermissionRequest{
				Tool:        extractTool(line),
				Description: line,
				RawContent:  paneContent,
			}
		}

		// Pattern 2: "Do you want to proceed?" with context above
		if strings.Contains(line, "Do you want to proceed?") {
			tool := "permission"
			desc := ""

			for j := i - 1; j >= 0 && j >= i-15; j-- {
				prev := strings.TrimSpace(lines[j])
				if prev == "" || isChrome(prev) {
					continue
				}

				lower := strings.ToLower(prev)

				if tool == "permission" {
					switch {
					case strings.Contains(lower, "bash command") || strings.Contains(lower, "bash"):
						tool = "Bash"
					case strings.Contains(lower, "read file") || strings.Contains(lower, "read"):
						tool = "Read"
					case strings.Contains(lower, "write file") || strings.Contains(lower, "write"):
						tool = "Write"
					case strings.Contains(lower, "edit file") || strings.Contains(lower, "edit"):
						tool = "Edit"
					case strings.Contains(lower, "glob") || strings.Contains(lower, "grep"):
						tool = "Search"
					}
				}

				if desc == "" && strings.HasPrefix(lines[j], "  ") {
					desc = prev
				}

				if tool != "permission" && desc != "" {
					break
				}
			}
			if desc == "" {
				desc = "Permission required"
			}
			return &PermissionRequest{
				Tool:        tool,
				Description: desc,
				RawContent:  paneContent,
			}
		}

		// Pattern 3: Trust prompt
		if strings.Contains(line, "I trust this folder") {
			return &PermissionRequest{
				Tool:        "Trust",
				Description: "Trust this workspace folder",
				RawContent:  paneContent,
			}
		}
	}
	return nil
}

func extractTool(line string) string {
	parts := strings.Fields(line)
	for i, p := range parts {
		if p == "Allow" && i+1 < len(parts) {
			tool := strings.TrimSuffix(parts[i+1], "?")
			return tool
		}
	}
	return "unknown"
}

func (c *claudeProvider) LaunchCmd(model, prompt string) string {
	// Never use -p flag — it runs in non-interactive mode and exits.
	// Prompt is sent separately via tmux send-keys after launch.
	cmd := "claude"
	if model != "" {
		cmd = fmt.Sprintf("claude --model %s", model)
	}
	return cmd
}

func (c *claudeProvider) CompactCmd() string {
	return "/compact"
}

func (c *claudeProvider) ApproveKey() string {
	return "Enter"
}

func (c *claudeProvider) DenyKey() string {
	return "Escape"
}

func (c *claudeProvider) ParseCost(paneContent string) (float64, int, int) {
	return 0, 0, 0
}

func (c *claudeProvider) BinaryName() string {
	return "claude"
}

func (c *claudeProvider) InstallHint() string {
	return "npm install -g @anthropic-ai/claude-code"
}
