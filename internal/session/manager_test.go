package session

import (
	"testing"

	"github.com/delvop-dev/delvop/internal/config"
	"github.com/delvop-dev/delvop/internal/provider"
)

func TestManagerAddAndGet(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg)

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

func TestManagerKPI(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("a1", p, "opus", "/tmp", "")
	s.State = provider.StateWorking
	s.CostUSD = 1.50

	s2, _ := mgr.Add("a2", p, "sonnet", "/tmp", "")
	s2.State = provider.StateWaitingForPermission
	s2.CostUSD = 0.75

	kpi := mgr.KPI()
	if kpi.ActiveCount != 2 {
		t.Errorf("expected 2 active, got %d", kpi.ActiveCount)
	}
	if kpi.TotalCost != 2.25 {
		t.Errorf("expected cost 2.25, got %f", kpi.TotalCost)
	}
}

func TestManagerRemove(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg)

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

func TestManagerNeedsAttention(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg)

	p, _ := provider.Get("claude")
	s, _ := mgr.Add("waiting", p, "opus", "/tmp", "")
	s.State = provider.StateWaitingForPermission

	s2, _ := mgr.Add("working", p, "opus", "/tmp", "")
	s2.State = provider.StateWorking

	attention := mgr.NeedsAttention()
	if len(attention) != 1 {
		t.Errorf("expected 1 needing attention, got %d", len(attention))
	}
}

func TestManagerSendKeys(t *testing.T) {
	cfg := config.Default()
	mgr := NewManager(cfg)

	err := mgr.SendKeys("nonexistent", "hello")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}
