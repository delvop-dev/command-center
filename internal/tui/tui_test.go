package tui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/delvop-dev/delvop/internal/config"
	"github.com/delvop-dev/delvop/internal/hooks"
	"github.com/delvop-dev/delvop/internal/notify"
	"github.com/delvop-dev/delvop/internal/provider"
	"github.com/delvop-dev/delvop/internal/session"
)

func newTestModel() Model {
	cfg := config.Default()
	mgr := session.NewManager(cfg, nil, nil)
	hookEngine := hooks.New("/tmp/test-tui.sock")
	notif := notify.New(cfg)
	cfg.Notify.Channels = []string{} // disable actual notifications
	return NewModel(cfg, mgr, hookEngine, notif)
}

func newTestModelWithSession() (Model, *session.Session) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	mgr := session.NewManager(cfg, nil, nil)
	hookEngine := hooks.New("/tmp/test-tui.sock")
	notif := notify.New(cfg)

	p, _ := provider.Get("claude")
	sess, _ := mgr.Add("test-agent", p, "opus", "/tmp", "")

	m := NewModel(cfg, mgr, hookEngine, notif)
	m.width = 120
	m.height = 40
	return m, sess
}

func TestNewModel(t *testing.T) {
	m := newTestModel()
	if m.viewMode != ViewDashboard {
		t.Errorf("expected ViewDashboard, got %d", m.viewMode)
	}
	if m.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0, got %d", m.selectedIdx)
	}
	if m.showHelp {
		t.Error("expected showHelp false")
	}
	if m.inputMode {
		t.Error("expected inputMode false")
	}
	if m.width != 0 {
		t.Errorf("expected width 0, got %d", m.width)
	}
}

func TestInit(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected non-nil init cmd")
	}
}

func TestInitNilHookEngine(t *testing.T) {
	cfg := config.Default()
	mgr := session.NewManager(cfg, nil, nil)
	notif := notify.New(cfg)
	m := NewModel(cfg, mgr, nil, notif)
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected non-nil init cmd (tick at minimum)")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := newTestModel()
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model := updated.(Model)
	if model.width != 100 {
		t.Errorf("expected width 100, got %d", model.width)
	}
	if model.height != 50 {
		t.Errorf("expected height 50, got %d", model.height)
	}
	if cmd != nil {
		t.Error("expected nil cmd for window size")
	}
}

func TestUpdateErrorMsg(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(ErrorMsg{Err: fmt.Errorf("test error")})
	model := updated.(Model)
	if model.statusMsg != "Error: test error" {
		t.Errorf("expected error status, got %q", model.statusMsg)
	}
	if model.statusExpiry.Before(time.Now()) {
		t.Error("expected status expiry in the future")
	}
}

func TestUpdateStatusMsg(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(StatusMsg{Message: "hello"})
	model := updated.(Model)
	if model.statusMsg != "hello" {
		t.Errorf("expected 'hello', got %q", model.statusMsg)
	}
}

func TestUpdateSessionCreatedMsg(t *testing.T) {
	m := newTestModel()
	updated, _ := m.Update(SessionCreatedMsg{SessionID: "abc"})
	model := updated.(Model)
	if model.statusMsg != "Agent created" {
		t.Errorf("expected 'Agent created', got %q", model.statusMsg)
	}
}

func TestUpdateTickMsg(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(TickMsg{})
	if cmd == nil {
		t.Error("expected non-nil cmd after tick")
	}
}

func TestUpdateHookEventMsg(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(HookEventMsg{})
	// Should return a command to listen for more events
	if cmd == nil {
		t.Error("expected non-nil cmd after hook event")
	}
}

func TestDashboardKeyQuit(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestDashboardKeyHelp(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := updated.(Model)
	if !model.showHelp {
		t.Error("expected showHelp true after ?")
	}

	// Toggle back
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = updated.(Model)
	if model.showHelp {
		t.Error("expected showHelp false after second ?")
	}
}

func TestDashboardKeyUpDown(t *testing.T) {
	m, _ := newTestModelWithSession()

	// Add a second session
	p, _ := provider.Get("claude")
	m.manager.Add("second", p, "opus", "/tmp", "")

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model := updated.(Model)
	if model.selectedIdx != 1 {
		t.Errorf("expected selectedIdx 1, got %d", model.selectedIdx)
	}

	// Move up
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)
	if model.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0, got %d", model.selectedIdx)
	}

	// Move up at 0 should stay at 0
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)
	if model.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0, got %d", model.selectedIdx)
	}
}

func TestDashboardKeyEnterAttach(t *testing.T) {
	m, _ := newTestModelWithSession()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	// Without a real tmux session, attach shows a status message instead
	// Just verify it doesn't panic and stays on dashboard
	if model.viewMode != ViewDashboard {
		t.Error("expected to stay on dashboard when session not running")
	}
}

func TestDashboardKeyNew(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model := updated.(Model)
	if !model.inputMode {
		t.Error("expected inputMode true after 'n'")
	}
	if model.inputPurpose != "new_agent" {
		t.Errorf("expected purpose 'new_agent', got %q", model.inputPurpose)
	}
}

func TestDashboardKeyKill(t *testing.T) {
	m, _ := newTestModelWithSession()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(Model)
	if len(model.manager.All()) != 0 {
		t.Error("expected 0 sessions after kill")
	}
	if model.statusMsg == "" {
		t.Error("expected status message after kill")
	}
}

func TestDashboardKeyMessage(t *testing.T) {
	m, _ := newTestModelWithSession()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model := updated.(Model)
	if !model.inputMode {
		t.Error("expected inputMode true after 'm'")
	}
	if model.inputPurpose != "message" {
		t.Errorf("expected purpose 'message', got %q", model.inputPurpose)
	}
}

func TestDashboardKeyMessageNoSessions(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model := updated.(Model)
	if model.inputMode {
		t.Error("should not enter input mode with no sessions")
	}
}

func TestDashboardKeyApproveNoPermission(t *testing.T) {
	m, _ := newTestModelWithSession()

	// Session is Idle, not waiting permission
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model := updated.(Model)
	if model.statusMsg != "" {
		t.Errorf("should not set status for non-permission session, got %q", model.statusMsg)
	}
}

func TestDashboardKeyApproveWithPermission(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.State = provider.StateWaitingForPermission

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model := updated.(Model)
	if model.statusMsg == "" {
		t.Error("expected status message after approve")
	}
}

func TestDashboardKeyDenyWithPermission(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.State = provider.StateWaitingForPermission

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	model := updated.(Model)
	if model.statusMsg == "" {
		t.Error("expected status message after deny")
	}
}

func TestDashboardKeyCompact(t *testing.T) {
	m, _ := newTestModelWithSession()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model := updated.(Model)
	// Will fail since no tmux session, but should set status
	if model.statusMsg == "" {
		t.Error("expected status message after compact attempt")
	}
}

func TestFocusedKeyEscape(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)
	if model.viewMode != ViewDashboard {
		t.Error("expected ViewDashboard after Esc")
	}
}

func TestFocusedKeyUpDown(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	m.scrollOffset = 5

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(Model)
	if model.scrollOffset != 4 {
		t.Errorf("expected scrollOffset 4, got %d", model.scrollOffset)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.scrollOffset != 5 {
		t.Errorf("expected scrollOffset 5, got %d", model.scrollOffset)
	}

	// Scroll up at 0 should stay at 0
	model.scrollOffset = 0
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)
	if model.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0, got %d", model.scrollOffset)
	}
}

func TestFocusedKeyTab(t *testing.T) {
	m, sess := newTestModelWithSession()

	p, _ := provider.Get("claude")
	sess2, _ := m.manager.Add("second", p, "opus", "/tmp", "")

	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := updated.(Model)
	if model.focusedID != sess2.ID {
		t.Errorf("expected focusedID %q after tab, got %q", sess2.ID, model.focusedID)
	}

	// Tab again should wrap around
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)
	if model.focusedID != sess.ID {
		t.Errorf("expected focusedID %q after tab wrap, got %q", sess.ID, model.focusedID)
	}
}

func TestFocusedKeyKill(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(Model)
	if model.viewMode != ViewDashboard {
		t.Error("expected ViewDashboard after killing focused session")
	}
	if len(model.manager.All()) != 0 {
		t.Error("expected 0 sessions after kill")
	}
}

func TestFocusedKeyMessage(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model := updated.(Model)
	if !model.inputMode {
		t.Error("expected inputMode true after 'm' in focused view")
	}
}

func TestFocusedKeyApprove(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	// Should not panic even though tmux isn't running
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	_ = updated.(Model)
}

func TestFocusedKeyDeny(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	_ = updated.(Model)
}

func TestFocusedKeyQuit(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit cmd")
	}
}

func TestInputKeyEnterNewAgent(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40
	m.inputMode = true
	m.inputPurpose = "new_agent"
	m.textInput.SetValue("my-agent")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	if model.inputMode {
		t.Error("expected inputMode false after enter")
	}
	if cmd == nil {
		t.Error("expected cmd for createAgent")
	}
}

func TestInputKeyEnterEmpty(t *testing.T) {
	m := newTestModel()
	m.inputMode = true
	m.inputPurpose = "new_agent"
	m.textInput.SetValue("")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	if model.inputMode {
		t.Error("expected inputMode false after enter with empty value")
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty input")
	}
}

func TestInputKeyEsc(t *testing.T) {
	m := newTestModel()
	m.inputMode = true
	m.inputPurpose = "new_agent"
	m.textInput.SetValue("something")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)
	if model.inputMode {
		t.Error("expected inputMode false after esc")
	}
}

func TestInputKeyOtherChars(t *testing.T) {
	m := newTestModel()
	m.inputMode = true
	m.inputPurpose = "new_agent"
	m.textInput.Focus()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model := updated.(Model)
	if !model.inputMode {
		t.Error("should remain in input mode for regular chars")
	}
}

func TestInputKeyEnterMessage(t *testing.T) {
	m, _ := newTestModelWithSession()
	m.inputMode = true
	m.inputPurpose = "message"
	m.textInput.SetValue("hello agent")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	if model.inputMode {
		t.Error("expected inputMode false after enter")
	}
	// The SendKeys call will fail (no tmux) but the status should be set
	if model.statusMsg != "Sent message to agent" {
		t.Errorf("expected sent message status, got %q", model.statusMsg)
	}
}

func TestInputKeyEnterMessageFocused(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	m.inputMode = true
	m.inputPurpose = "message"
	m.textInput.SetValue("hello")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	if model.inputMode {
		t.Error("expected inputMode false")
	}
}

// View tests

func TestViewLoading(t *testing.T) {
	m := newTestModel()
	// width is 0
	view := m.View()
	if view != "Loading..." {
		t.Errorf("expected 'Loading...', got %q", view)
	}
}

func TestViewEmptyState(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
	if view == "Loading..." {
		t.Error("should not show loading with width set")
	}
}

func TestViewDashboardWithSessions(t *testing.T) {
	m, _ := newTestModelWithSession()
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestViewHelpOverlay(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40
	m.showHelp = true
	view := m.View()
	if view == "" {
		t.Error("expected non-empty help view")
	}
}

func TestViewFocused(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	view := m.View()
	if view == "" {
		t.Error("expected non-empty focused view")
	}
}

func TestViewFocusedInvalidID(t *testing.T) {
	m, _ := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = "nonexistent"
	// Should fallback to dashboard
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view on fallback")
	}
}

func TestViewFocusedWithPermission(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.State = provider.StateWaitingForPermission
	sess.Permission = &provider.PermissionRequest{
		Tool:        "Bash",
		Description: "Run command",
	}
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with permission")
	}
}

func TestViewFocusedWithEvents(t *testing.T) {
	m, sess := newTestModelWithSession()
	for i := 0; i < 25; i++ {
		sess.Events = append(sess.Events, session.Event{
			Time:    time.Now(),
			Type:    "state_change",
			Message: fmt.Sprintf("event %d", i),
		})
	}
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with events")
	}
}

func TestViewFocusedWithFileChanges(t *testing.T) {
	m, sess := newTestModelWithSession()
	for i := 0; i < 15; i++ {
		sess.FileChanges = append(sess.FileChanges, session.FileChange{
			Path:      fmt.Sprintf("file%d.go", i),
			Operation: "modify",
			Timestamp: time.Now(),
		})
	}
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with file changes")
	}
}

func TestViewInputMode(t *testing.T) {
	m, _ := newTestModelWithSession()
	m.inputMode = true
	m.textInput.Focus()
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view in input mode")
	}
}

func TestViewDashboardWithMultipleSessions(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.CostUSD = 1.5
	sess.TokensIn = 1000
	sess.TokensOut = 500

	p, _ := provider.Get("claude")
	sess2, _ := m.manager.Add("second", p, "opus", "/tmp", "")
	sess2.State = provider.StateWaitingForPermission
	sess2.Permission = &provider.PermissionRequest{
		Tool:        "Write",
		Description: "Write file",
	}
	sess2.Events = append(sess2.Events, session.Event{
		Time:    time.Now(),
		Type:    "state_change",
		Message: "working",
	})

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with multiple sessions")
	}
}

func TestViewDashboardWithStatusMsg(t *testing.T) {
	m, _ := newTestModelWithSession()
	m.statusMsg = "Test status"
	m.statusExpiry = time.Now().Add(5 * time.Second)
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with status")
	}
}

func TestViewDashboardWithAttention(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.State = provider.StateWaitingForPermission
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with attention")
	}
}

// Helper function tests

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"hi", 2, "hi"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		{"", 5, ""},
		{"a", 1, "a"},
		{"ab", 1, "a"},
		{"abcdef", 4, "a..."},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m0s"},
		{5*time.Minute + 30*time.Second, "5m30s"},
		{2 * time.Hour, "2h0m"},
		{2*time.Hour + 15*time.Minute, "2h15m"},
		{0, "0s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatAgo(t *testing.T) {
	tests := []struct {
		offset time.Duration
		want   string
	}{
		{10 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{2 * time.Hour, "2h ago"},
	}

	for _, tt := range tests {
		tm := time.Now().Add(-tt.offset)
		got := formatAgo(tm)
		if got != tt.want {
			t.Errorf("formatAgo(-%v) = %q, want %q", tt.offset, got, tt.want)
		}
	}
}

func TestStateDisplayLabel(t *testing.T) {
	tests := []struct {
		state provider.AgentState
		want  string
	}{
		{provider.StateIdle, "IDLE"},
		{provider.StateWorking, "WORKING"},
		{provider.StateThinking, "WORKING"},
		{provider.StateEditing, "WORKING"},
		{provider.StateRunningTool, "WORKING"},
		{provider.StateCompacting, "WORKING"},
		{provider.StateWaitingForPermission, "NEEDS INPUT"},
		{provider.StateWaitingInput, "NEEDS FOCUS"},
		{provider.StateError, "ERROR"},
		{provider.StateUnknown, "UNKNOWN"},
		{provider.AgentState(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := stateDisplayLabel(tt.state)
		if got != tt.want {
			t.Errorf("stateDisplayLabel(%v) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestViewModeConstants(t *testing.T) {
	if ViewDashboard != 0 {
		t.Errorf("expected ViewDashboard=0, got %d", ViewDashboard)
	}
	if ViewFocused != 1 {
		t.Errorf("expected ViewFocused=1, got %d", ViewFocused)
	}
}

func TestTickCmd(t *testing.T) {
	cmd := tickCmd(100)
	if cmd == nil {
		t.Error("expected non-nil tick cmd")
	}
}

func TestListenForHookEventsNil(t *testing.T) {
	cmd := listenForHookEvents(nil)
	if cmd != nil {
		t.Error("expected nil cmd for nil engine")
	}
}

func TestDashboardKeyDownAtEnd(t *testing.T) {
	m, _ := newTestModelWithSession()

	// Only one session, selectedIdx already at 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model := updated.(Model)
	if model.selectedIdx != 0 {
		t.Errorf("expected selectedIdx 0 (can't go past end), got %d", model.selectedIdx)
	}
}

func TestDashboardKeyApproveNoSessions(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	// Should not panic
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	_ = updated.(Model)
}

func TestDashboardKeyDenyNoSessions(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	_ = updated.(Model)
}

func TestDashboardKeyKillNoSessions(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	_ = updated.(Model)
}

func TestDashboardKeyCompactNoSessions(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	_ = updated.(Model)
}

func TestDashboardKeyEnterNoSessions(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	if model.viewMode != ViewDashboard {
		t.Error("should stay on dashboard with no sessions")
	}
}

func TestFocusedKeyEnter(t *testing.T) {
	m, sess := newTestModelWithSession()
	m.viewMode = ViewFocused
	m.focusedID = sess.ID

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// The attach cmd will attempt to exec, which returns a tea.Cmd
	if cmd == nil {
		t.Error("expected non-nil cmd for attach")
	}
}

func TestFocusedKeyEnterNoAttachCmd(t *testing.T) {
	m := newTestModel()
	m.width = 100
	m.height = 40
	m.viewMode = ViewFocused
	m.focusedID = "nonexistent"

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for nonexistent attach")
	}
}

func TestViewAgentCardWithEvents(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.Events = append(sess.Events, session.Event{
		Time:    time.Now(),
		Type:    "state_change",
		Message: "thinking -> working",
	})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestViewAgentCardWithFileChanges(t *testing.T) {
	m, sess := newTestModelWithSession()
	sess.FileChanges = append(sess.FileChanges, session.FileChange{
		Path:      "main.go",
		Operation: "modify",
	})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestViewNarrowWidth(t *testing.T) {
	m, _ := newTestModelWithSession()
	m.width = 20
	m.height = 40
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view at narrow width")
	}
}

func TestDashboardKeyKillAdjustsIndex(t *testing.T) {
	cfg := config.Default()
	cfg.Notify.Channels = []string{}
	mgr := session.NewManager(cfg, nil, nil)
	hookEngine := hooks.New("/tmp/test.sock")
	notif := notify.New(cfg)

	p, _ := provider.Get("claude")
	mgr.Add("first", p, "opus", "/tmp", "")
	s2, _ := mgr.Add("second", p, "opus", "/tmp", "")

	m := NewModel(cfg, mgr, hookEngine, notif)
	m.width = 120
	m.height = 40
	m.selectedIdx = 1 // Select "second"

	// Kill "second" (selected)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(Model)
	_ = s2
	if model.selectedIdx != 0 {
		t.Errorf("expected selectedIdx adjusted to 0, got %d", model.selectedIdx)
	}
}

func TestFocusedViewScrollWithEvents(t *testing.T) {
	m, sess := newTestModelWithSession()
	for i := 0; i < 30; i++ {
		sess.Events = append(sess.Events, session.Event{
			Time:    time.Now(),
			Type:    "state_change",
			Message: fmt.Sprintf("event %d", i),
		})
	}
	m.viewMode = ViewFocused
	m.focusedID = sess.ID
	m.scrollOffset = 5

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with scroll")
	}
}
