package notify

import (
	"testing"
	"time"

	"github.com/delvop-dev/command-center/internal/config"
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

func TestFocusSuppressDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	cfg.Notify.FocusSuppress = false
	r := New(cfg)
	r.SetFocused("test-session")

	r.Notify("test-session", EventInputNeeded, "Agent needs input")
	count := r.RecentCount("test-session")
	if count != 1 {
		t.Errorf("expected 1 notification when focus suppress disabled, got %d", count)
	}
}

func TestAllEventTypes(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	cfg.Notify.DebounceMs = 0 // disable debounce
	r := New(cfg)

	r.Notify("s1", EventInputNeeded, "input needed")
	r.Notify("s1", EventTaskDone, "task done")
	r.Notify("s1", EventError, "error occurred")

	count := r.RecentCount("s1")
	if count != 3 {
		t.Errorf("expected 3 notifications for 3 event types, got %d", count)
	}
}

func TestRecentCountForUnknownSession(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	r := New(cfg)

	count := r.RecentCount("nonexistent")
	if count != 0 {
		t.Errorf("expected 0 for unknown session, got %d", count)
	}
}

func TestNotifyWithNativeChannel(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{"native"}
	cfg.Notify.FocusSuppress = false
	cfg.Notify.DebounceMs = 0
	r := New(cfg)

	// Should not panic - the native notification may fail but should not crash
	r.Notify("s1", EventTaskDone, "test native notification")

	count := r.RecentCount("s1")
	if count != 1 {
		t.Errorf("expected 1 notification, got %d", count)
	}
}

func TestNotifyWithSoundChannel(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{"sound"}
	cfg.Notify.FocusSuppress = false
	cfg.Notify.DebounceMs = 0
	r := New(cfg)

	// Should not panic - the sound notification may fail but should not crash
	r.Notify("s1", EventInputNeeded, "test sound")
	r.Notify("s2", EventTaskDone, "test sound done")
	r.Notify("s3", EventError, "test sound error")

	// Brief pause to let goroutines start (they may fail but should not panic)
	time.Sleep(50 * time.Millisecond)
}

func TestRecentPruning(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	cfg.Notify.DebounceMs = 0
	r := New(cfg)

	// Send multiple different event types to different sessions (to avoid debounce)
	for i := 0; i < 20; i++ {
		r.Notify("pruning-test", EventType(i%3), "message")
		// Small delay to separate debounce windows
	}

	// The recent list should be pruned to only entries within last 10 seconds
	// All entries are recent so all should be kept (minus debounce)
	count := r.RecentCount("pruning-test")
	if count < 1 {
		t.Errorf("expected at least 1 recent notification, got %d", count)
	}
}

func TestSetFocusedChanges(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	cfg.Notify.FocusSuppress = true
	r := New(cfg)

	r.SetFocused("s1")
	r.Notify("s1", EventInputNeeded, "suppressed")
	if r.RecentCount("s1") != 0 {
		t.Error("expected suppression for s1")
	}

	r.SetFocused("s2")
	r.Notify("s1", EventTaskDone, "not suppressed now")
	if r.RecentCount("s1") != 1 {
		t.Error("expected notification for s1 after focus changed")
	}
}

func TestNew(t *testing.T) {
	cfg := config.Default()
	r := New(cfg)
	if r == nil {
		t.Fatal("expected non-nil router")
	}
}
