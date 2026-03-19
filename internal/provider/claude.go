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

func (c *claudeProvider) ParseState(paneContent string) AgentState {
	lines := strings.Split(paneContent, "\n")
	// Check from the bottom up for the most recent state indicator.
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Permission prompt
		if strings.Contains(line, "Allow") && strings.Contains(line, "?") {
			return StateWaitingPermission
		}
		if strings.Contains(line, "Do you want to proceed?") {
			return StateWaitingPermission
		}

		// Waiting for user input
		if strings.HasSuffix(line, ">") || strings.HasPrefix(line, ">") {
			return StateWaitingInput
		}

		// Compacting (check before thinking since compacting messages may contain "...")
		if strings.Contains(line, "Compacting") || strings.Contains(line, "compacting") {
			return StateCompacting
		}

		// Spinner / thinking indicators
		if strings.Contains(line, "Thinking") || strings.Contains(line, "...") {
			return StateThinking
		}

		// Tool use
		if strings.Contains(line, "Running") || strings.Contains(line, "Executing") {
			return StateRunningTool
		}

		// Editing
		if strings.Contains(line, "Editing") || strings.Contains(line, "Writing") {
			return StateEditing
		}

		// Idle prompt
		if strings.Contains(line, "$") || strings.Contains(line, "Claude") {
			return StateIdle
		}

		break
	}

	return StateUnknown
}

func (c *claudeProvider) ParsePermission(paneContent string) *PermissionRequest {
	lines := strings.Split(paneContent, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "Allow") && strings.Contains(line, "?") {
			return &PermissionRequest{
				Tool:        extractTool(line),
				Description: line,
				RawContent:  paneContent,
			}
		}
	}
	return nil
}

func extractTool(line string) string {
	// Try to extract the tool name from "Allow <tool>?" style prompts
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
	cmd := "claude"
	if model != "" {
		cmd = fmt.Sprintf("claude --model %s", model)
	}
	if prompt != "" {
		cmd = fmt.Sprintf("%s -p %q", cmd, prompt)
	}
	return cmd
}

func (c *claudeProvider) CompactCmd() string {
	return "/compact"
}

func (c *claudeProvider) ApproveKey() string {
	return "y"
}

func (c *claudeProvider) DenyKey() string {
	return "n"
}

func (c *claudeProvider) ParseCost(paneContent string) (float64, int, int) {
	return 0, 0, 0
}
