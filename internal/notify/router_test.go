package notify

import (
	"testing"

	"github.com/delvop-dev/delvop/internal/config"
)

func TestDebounce(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	r := New(cfg)

	r.Notify("test-session", EventInputNeeded, "Agent needs input")
	r.Notify("test-session", EventInputNeeded, "Agent needs input again")

	count := r.RecentCount("test-session")
	if count > 1 {
		t.Errorf("expected debounce to collapse, got %d", count)
	}
}

func TestFocusSuppression(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	cfg.Notify.FocusSuppress = true
	r := New(cfg)
	r.SetFocused("test-session")

	r.Notify("test-session", EventInputNeeded, "Agent needs input")
	count := r.RecentCount("test-session")
	if count != 0 {
		t.Errorf("expected focus suppression, got %d", count)
	}
}

func TestDifferentSessionNotSuppressed(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	cfg.Notify.FocusSuppress = true
	r := New(cfg)
	r.SetFocused("focused-session")

	r.Notify("other-session", EventInputNeeded, "Other agent needs input")
	count := r.RecentCount("other-session")
	if count != 1 {
		t.Errorf("expected 1 notification for unfocused session, got %d", count)
	}
}

func TestDifferentEventTypesNotDebounced(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	r := New(cfg)

	r.Notify("s1", EventInputNeeded, "needs input")
	r.Notify("s1", EventTaskDone, "task done")

	count := r.RecentCount("s1")
	if count != 2 {
		t.Errorf("expected 2 different event types, got %d", count)
	}
}
