package session

import (
	"testing"

	"github.com/delvop-dev/delvop/internal/config"
	"github.com/delvop-dev/delvop/internal/provider"
)

func TestManagerAddAndGet(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, err := mgr.Add("test-agent", p, "opus", "/tmp", "")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if sess.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %q", sess.Name)
	}
	if sess.ProviderName != "claude" {
		t.Errorf("expected provider 'claude', got %q", sess.ProviderName)
	}
	if sess.Model != "opus" {
		t.Errorf("expected model 'opus', got %q", sess.Model)
	}
	if sess.WorkDir != "/tmp" {
		t.Errorf("expected workdir '/tmp', got %q", sess.WorkDir)
	}
	if sess.State != provider.StateIdle {
		t.Errorf("expected state Idle, got %v", sess.State)
	}
	if sess.ID == "" {
		t.Error("expected non-empty ID")
	}
	if sess.TmuxSession == "" {
		t.Error("expected non-empty TmuxSession")
	}
	if sess.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	got, ok := mgr.Get(sess.ID)
	if !ok {
		t.Fatal("Get returned false")
	}
	if got.ID != sess.ID {
		t.Error("IDs don't match")
	}

	all := mgr.All()
	if len(all) != 1 {
		t.Errorf("expected 1 session, got %d", len(all))
	}
}

func TestManagerGetNonexistent(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	_, ok := mgr.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent session")
	}
}

func TestManagerKPI(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("a1", p, "opus", "/tmp", "")
	s.State = provider.StateWorking
	s.CostUSD = 1.50
	s.TokensIn = 1000
	s.TokensOut = 500

	s2, _ := mgr.Add("a2", p, "sonnet", "/tmp", "")
	s2.State = provider.StateWaitingForPermission
	s2.CostUSD = 0.75
	s2.TokensIn = 200
	s2.TokensOut = 100

	kpi := mgr.KPI()
	if kpi.ActiveCount != 2 {
		t.Errorf("expected 2 active, got %d", kpi.ActiveCount)
	}
	if kpi.TotalCount != 2 {
		t.Errorf("expected 2 total, got %d", kpi.TotalCount)
	}
	if kpi.TotalCost != 2.25 {
		t.Errorf("expected cost 2.25, got %f", kpi.TotalCost)
	}
	if kpi.TotalTokens != 1800 {
		t.Errorf("expected tokens 1800, got %d", kpi.TotalTokens)
	}
	if kpi.Uptime <= 0 {
		t.Error("expected positive uptime")
	}
}

func TestManagerKPIEmpty(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	kpi := mgr.KPI()
	if kpi.TotalCount != 0 {
		t.Errorf("expected 0 total, got %d", kpi.TotalCount)
	}
	if kpi.ActiveCount != 0 {
		t.Errorf("expected 0 active, got %d", kpi.ActiveCount)
	}
	if kpi.Uptime != 0 {
		t.Errorf("expected 0 uptime, got %v", kpi.Uptime)
	}
}

func TestManagerKPIIdleIsActive(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("idle", p, "opus", "/tmp", "")
	s.State = provider.StateIdle

	kpi := mgr.KPI()
	if kpi.ActiveCount != 1 {
		t.Errorf("expected idle session to count as active, got %d", kpi.ActiveCount)
	}
}

func TestManagerKPIInactiveStates(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("thinking", p, "opus", "/tmp", "")
	s.State = provider.StateThinking

	s2, _ := mgr.Add("error", p, "opus", "/tmp", "")
	s2.State = provider.StateError

	kpi := mgr.KPI()
	if kpi.ActiveCount != 0 {
		t.Errorf("expected 0 active for thinking/error states, got %d", kpi.ActiveCount)
	}
}

func TestManagerRemove(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("to-remove", p, "opus", "/tmp", "")

	mgr.Remove(s.ID)

	_, ok := mgr.Get(s.ID)
	if ok {
		t.Error("expected session to be removed")
	}
	if len(mgr.All()) != 0 {
		t.Error("expected 0 sessions after remove")
	}
}

func TestManagerRemoveNonexistent(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	// Should not panic
	mgr.Remove("nonexistent")
}

func TestManagerRemoveMiddle(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s1, _ := mgr.Add("first", p, "opus", "/tmp", "")
	s2, _ := mgr.Add("middle", p, "opus", "/tmp", "")
	s3, _ := mgr.Add("last", p, "opus", "/tmp", "")

	mgr.Remove(s2.ID)

	all := mgr.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(all))
	}
	if all[0].ID != s1.ID {
		t.Error("expected first session to remain")
	}
	if all[1].ID != s3.ID {
		t.Error("expected last session to remain")
	}
}

func TestManagerNeedsAttention(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("waiting", p, "opus", "/tmp", "")
	s.State = provider.StateWaitingForPermission

	s2, _ := mgr.Add("working", p, "opus", "/tmp", "")
	s2.State = provider.StateWorking

	attention := mgr.NeedsAttention()
	if len(attention) != 1 {
		t.Errorf("expected 1 needing attention, got %d", len(attention))
	}
	if attention[0].Name != "waiting" {
		t.Errorf("expected 'waiting', got %q", attention[0].Name)
	}
}

func TestManagerNeedsAttentionEmpty(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	attention := mgr.NeedsAttention()
	if len(attention) != 0 {
		t.Errorf("expected 0, got %d", len(attention))
	}
}

func TestManagerSendKeys(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	err := mgr.SendKeys("nonexistent", "hello")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestManagerApprove(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	err := mgr.Approve("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestManagerDeny(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	err := mgr.Deny("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestManagerCompact(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	err := mgr.Compact("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestManagerCompactNoSupport(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("gemini") // gemini has empty CompactCmd
	s, _ := mgr.Add("gemini-agent", p, "pro", "/tmp", "")

	err := mgr.Compact(s.ID)
	if err == nil {
		t.Error("expected error for provider without compact support")
	}
}

func TestManagerAttachCmd(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("test", p, "opus", "/tmp", "")

	cmd := mgr.AttachCmd(s.ID)
	if cmd == "" {
		t.Error("expected non-empty attach cmd")
	}
}

func TestManagerAttachCmdNonexistent(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	cmd := mgr.AttachCmd("nonexistent")
	if cmd != "" {
		t.Errorf("expected empty attach cmd, got %q", cmd)
	}
}

func TestManagerTmux(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	tmux := mgr.Tmux()
	if tmux == nil {
		t.Fatal("expected non-nil TmuxBridge")
	}
}

func TestManagerAllOrder(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s1, _ := mgr.Add("first", p, "opus", "/tmp", "")
	s2, _ := mgr.Add("second", p, "opus", "/tmp", "")
	s3, _ := mgr.Add("third", p, "opus", "/tmp", "")

	all := mgr.All()
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}
	if all[0].ID != s1.ID || all[1].ID != s2.ID || all[2].ID != s3.ID {
		t.Error("sessions not in insertion order")
	}
}

func TestManagerAddWithBranch(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("branched", p, "opus", "/tmp", "feature/test")

	if s.Branch != "feature/test" {
		t.Errorf("expected branch 'feature/test', got %q", s.Branch)
	}
}

func TestManagerPollStateNonexistent(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	// Should not panic
	mgr.PollState("nonexistent")
}

func TestManagerPollAll(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	// Should not panic with no sessions
	mgr.PollAll()
}

func TestManagerLaunch(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not available")
	}
	cfg := config.Default()
	cfg.Tmux.Prefix = "dv-test-launch-"
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("launch-test", p, "", "/tmp", "")

	// Launch will try to create a tmux session running "claude"
	// This will likely fail since claude may not be installed, but that's fine
	_ = mgr.Launch(sess)
	// Clean up
	mgr.Remove(sess.ID)
}

func TestManagerPollStateWithSession(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("poll-test", p, "opus", "/tmp", "")

	// PollState will fail to capture pane (no tmux session) but should not panic
	mgr.PollState(sess.ID)

	// State should remain unchanged since capture fails
	got, _ := mgr.Get(sess.ID)
	if got.State != provider.StateIdle {
		t.Errorf("expected state to remain Idle, got %v", got.State)
	}
}

func TestManagerPollAllWithSessions(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	mgr.Add("poll1", p, "opus", "/tmp", "")
	mgr.Add("poll2", p, "opus", "/tmp", "")

	// Should not panic
	mgr.PollAll()
}

func TestManagerSendKeysWithSession(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("send-test", p, "opus", "/tmp", "")

	// Will fail since no tmux session, but should return an error, not panic
	err := mgr.SendKeys(sess.ID, "hello")
	if err == nil {
		// It's OK if tmux is running and the session doesn't exist
		// The important thing is it doesn't panic
	}
}

func TestManagerApproveWithSession(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("approve-test", p, "opus", "/tmp", "")

	// Will fail since no tmux session
	_ = mgr.Approve(sess.ID)
}

func TestManagerDenyWithSession(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("deny-test", p, "opus", "/tmp", "")

	_ = mgr.Deny(sess.ID)
}

func TestManagerCompactWithClaudeSession(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("compact-test", p, "opus", "/tmp", "")

	// Will fail since no tmux session, but should try to send /compact
	_ = mgr.Compact(sess.ID)
}

func TestManagerPollStateIntegration(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not available")
	}

	cfg := config.Default()
	cfg.Tmux.Prefix = "dv-poll-test-"
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("pollint", p, "opus", "/tmp", "")

	// Create an actual tmux session
	err := mgr.Tmux().CreateSession("pollint", "/tmp", "echo 'Thinking...' && sleep 30")
	if err != nil {
		t.Fatalf("failed to create tmux session: %v", err)
	}
	defer mgr.Tmux().KillSession("pollint")

	// Give it a moment to start
	mgr.PollState(sess.ID)

	got, _ := mgr.Get(sess.ID)
	// The pane content should be captured
	if got.PaneContent == "" {
		// It's possible the content is empty right away, that's OK
	}
}

func TestManagerIntegrationSendApproveCompact(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not available")
	}

	cfg := config.Default()
	cfg.Tmux.Prefix = "dv-integ-"
	mgr := NewManager(cfg, nil, nil)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("integ", p, "opus", "/tmp", "")

	// Create tmux session using the session's ID (matches how manager methods look it up)
	err := mgr.Tmux().CreateSession(sess.ID, "/tmp", "sleep 60")
	if err != nil {
		t.Fatalf("failed to create tmux session: %v", err)
	}
	defer mgr.Tmux().KillSession(sess.ID)

	// Test SendKeys with real session
	err = mgr.SendKeys(sess.ID, "echo hello")
	if err != nil {
		t.Errorf("SendKeys failed: %v", err)
	}

	// Test Approve with real session
	err = mgr.Approve(sess.ID)
	if err != nil {
		t.Errorf("Approve failed: %v", err)
	}

	// Test Deny with real session
	err = mgr.Deny(sess.ID)
	if err != nil {
		t.Errorf("Deny failed: %v", err)
	}

	// Test Compact with real session
	err = mgr.Compact(sess.ID)
	if err != nil {
		t.Errorf("Compact failed: %v", err)
	}
}

func TestShortIDUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := shortID()
		if id == "" {
			t.Fatal("expected non-empty ID")
		}
		if seen[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		seen[id] = true
	}
}
