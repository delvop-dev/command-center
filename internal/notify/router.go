package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/delvop-dev/delvop/internal/config"
)

type EventType int

const (
	EventInputNeeded EventType = iota
	EventTaskDone
	EventError
)

type notification struct {
	sessionID string
	eventType EventType
	message   string
	time      time.Time
}

type Router struct {
	cfg     *config.Config
	focused string
	recent  []notification
	mu      sync.Mutex
}

func New(cfg *config.Config) *Router {
	return &Router{cfg: cfg}
}

func (r *Router) SetFocused(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.focused = sessionID
}

func (r *Router) Notify(sessionID string, eventType EventType, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cfg.Notify.FocusSuppress && sessionID == r.focused {
		return
	}

	debounce := time.Duration(r.cfg.Notify.DebounceMs) * time.Millisecond
	for _, n := range r.recent {
		if n.sessionID == sessionID && n.eventType == eventType && time.Since(n.time) < debounce {
			return
		}
	}

	n := notification{sessionID: sessionID, eventType: eventType, message: message, time: time.Now()}
	r.recent = append(r.recent, n)

	cutoff := time.Now().Add(-10 * time.Second)
	pruned := r.recent[:0]
	for _, n := range r.recent {
		if n.time.After(cutoff) {
			pruned = append(pruned, n)
		}
	}
	r.recent = pruned

	for _, ch := range r.cfg.Notify.Channels {
		switch ch {
		case "native":
			go r.sendNative(message)
		case "sound":
			go r.sendSound(eventType)
		}
	}
}

func (r *Router) RecentCount(sessionID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, n := range r.recent {
		if n.sessionID == sessionID {
			count++
		}
	}
	return count
}

func (r *Router) sendNative(message string) {
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf(`display notification "%s" with title "delvop" sound name "Basso"`, message)
		exec.Command("osascript", "-e", script).Run()
	} else {
		exec.Command("notify-send", "delvop", message).Run()
	}
}

func (r *Router) sendSound(eventType EventType) {
	if runtime.GOOS != "darwin" {
		return
	}
	sound := r.cfg.Notify.Sound.InputNeeded
	switch eventType {
	case EventTaskDone:
		sound = r.cfg.Notify.Sound.TaskDone
	case EventError:
		sound = r.cfg.Notify.Sound.Error
	}
	exec.Command("afplay", fmt.Sprintf("/System/Library/Sounds/%s.aiff", sound)).Run()
}
