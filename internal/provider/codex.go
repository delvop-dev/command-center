package provider

type codexProvider struct{}

func init() {
	Register(&codexProvider{})
}

func (c *codexProvider) Name() string {
	return "codex"
}

func (c *codexProvider) ParseState(paneContent string) AgentState {
	return StateUnknown
}

func (c *codexProvider) ParsePermission(paneContent string) *PermissionRequest {
	return nil
}

func (c *codexProvider) LaunchCmd(model, prompt string) string {
	cmd := "codex"
	if prompt != "" {
		cmd = cmd + " " + prompt
	}
	return cmd
}

func (c *codexProvider) CompactCmd() string {
	return ""
}

func (c *codexProvider) ApproveKey() string {
	return "y"
}

func (c *codexProvider) DenyKey() string {
	return "n"
}

func (c *codexProvider) ParseCost(paneContent string) (float64, int, int) {
	return 0, 0, 0
}
