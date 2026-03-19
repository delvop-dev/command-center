package provider

type geminiProvider struct{}

func init() {
	Register(&geminiProvider{})
}

func (g *geminiProvider) Name() string {
	return "gemini"
}

func (g *geminiProvider) ParseState(paneContent string) AgentState {
	return StateUnknown
}

func (g *geminiProvider) ParsePermission(paneContent string) *PermissionRequest {
	return nil
}

func (g *geminiProvider) LaunchCmd(model, prompt string) string {
	cmd := "gemini"
	if prompt != "" {
		cmd = cmd + " " + prompt
	}
	return cmd
}

func (g *geminiProvider) CompactCmd() string {
	return ""
}

func (g *geminiProvider) ApproveKey() string {
	return "y"
}

func (g *geminiProvider) DenyKey() string {
	return "n"
}

func (g *geminiProvider) ParseCost(paneContent string) (float64, int, int) {
	return 0, 0, 0
}
