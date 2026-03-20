package session

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/delvop-dev/delvop/internal/config"
	"github.com/delvop-dev/delvop/internal/provider"
)

type Manager struct {
	sessions map[string]*Session
	order    []string
	tmux     *TmuxBridge
	cfg      *config.Config
	mu       sync.RWMutex
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		tmux:     NewTmuxBridge(cfg.Tmux.Prefix),
		cfg:      cfg,
	}
}

func (m *Manager) Add(name string, p provider.AgentProvider, model, workDir, branch string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := shortID()
	sess := &Session{
		ID:           id,
		Name:         name,
		TmuxSession:  m.tmux.SessionName(id),
		Provider:     p,
		ProviderName: strings.ToLower(strings.Split(p.Name(), " ")[0]),
		Model:        model,
		WorkDir:      workDir,
		Branch:       branch,
		State:        provider.StateIdle,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.sessions[id] = sess
	m.order = append(m.order, id)
	return sess, nil
}

func (m *Manager) Launch(sess *Session) error {
	cmd := sess.Provider.LaunchCmd(sess.Model, "")
	return m.tmux.CreateSession(sess.ID, sess.WorkDir, cmd)
}

func (m *Manager) LaunchWithPrompt(sess *Session, prompt string) error {
	if prompt != "" {
		sess.HasWorked = true // Agent will start working immediately
	}
	cmd := sess.Provider.LaunchCmd(sess.Model, prompt)
	return m.tmux.CreateSession(sess.ID, sess.WorkDir, cmd)
}

func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[id]
	return s, ok
}

func (m *Manager) All() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Session, 0, len(m.order))
	for _, id := range m.order {
		if s, ok := m.sessions[id]; ok {
			result = append(result, s)
		}
	}
	return result
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		_ = m.tmux.KillSession(s.ID)
		delete(m.sessions, id)
		for i, oid := range m.order {
			if oid == id {
				m.order = append(m.order[:i], m.order[i+1:]...)
				break
			}
		}
	}
}

func (m *Manager) PollState(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return
	}
	content, err := m.tmux.CapturePaneContent(s.ID, 100)
	if err != nil {
		return
	}
	s.PaneContent = content
	newState := s.Provider.ParseState(content)

	// Don't show "NEEDS FOCUS" for a brand-new agent that hasn't done any work yet.
	// The bare ❯ prompt on launch is just idle, not waiting for user focus.
	if newState == provider.StateWaitingInput && !s.HasWorked {
		newState = provider.StateIdle
	}
	// Track once the agent has done real work (thinking, editing, tool use, permission)
	if newState == provider.StateThinking || newState == provider.StateWorking ||
		newState == provider.StateEditing || newState == provider.StateRunningTool ||
		newState == provider.StateWaitingForPermission {
		s.HasWorked = true
	}

	if newState != s.State {
		oldState := s.State
		s.State = newState
		s.UpdatedAt = time.Now()
		s.Events = append(s.Events, Event{
			Time:    time.Now(),
			Type:    "state_change",
			Message: fmt.Sprintf("%s → %s", oldState, newState),
		})
	}
	if newState == provider.StateWaitingForPermission {
		s.Permission = s.Provider.ParsePermission(content)
	} else {
		s.Permission = nil
	}
	cost, tokIn, tokOut := s.Provider.ParseCost(content)
	if cost > 0 {
		s.CostUSD = cost
		s.TokensIn = tokIn
		s.TokensOut = tokOut
	}
}

func (m *Manager) PollAll() {
	m.mu.RLock()
	ids := make([]string, len(m.order))
	copy(ids, m.order)
	m.mu.RUnlock()

	for _, id := range ids {
		m.PollState(id)
	}
}

func (m *Manager) SendKeys(id, keys string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	return m.tmux.SendKeys(s.ID, keys)
}

func (m *Manager) Approve(id string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	return m.tmux.SendRawKey(s.ID, s.Provider.ApproveKey())
}

func (m *Manager) Deny(id string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	return m.tmux.SendRawKey(s.ID, s.Provider.DenyKey())
}

func (m *Manager) Compact(id string) error {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("session %s not found", id)
	}
	cmd := s.Provider.CompactCmd()
	if cmd == "" {
		return fmt.Errorf("provider %s does not support compacting", s.Provider.Name())
	}
	return m.tmux.SendKeys(s.ID, cmd)
}

func (m *Manager) AttachCmd(id string) string {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return ""
	}
	return m.tmux.AttachCmd(s.ID)
}

// TmuxSessionName returns the full tmux session name for a given session ID.
func (m *Manager) TmuxSessionName(id string) string {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	if !ok {
		return ""
	}
	return m.tmux.SessionName(s.ID)
}

func (m *Manager) NeedsAttention() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Session
	for _, id := range m.order {
		if s, ok := m.sessions[id]; ok && (s.State == provider.StateWaitingForPermission || s.State == provider.StateWaitingInput) {
			result = append(result, s)
		}
	}
	return result
}

func (m *Manager) KPI() KPIData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	kpi := KPIData{TotalCount: len(m.sessions)}
	var earliest time.Time
	for _, s := range m.sessions {
		if s.State == provider.StateWorking || s.State == provider.StateWaitingForPermission || s.State == provider.StateIdle {
			kpi.ActiveCount++
		}
		kpi.TotalCost += s.CostUSD
		kpi.TotalTokens += s.TokensIn + s.TokensOut
		if earliest.IsZero() || s.CreatedAt.Before(earliest) {
			earliest = s.CreatedAt
		}
	}
	if !earliest.IsZero() {
		kpi.Uptime = time.Since(earliest)
	}
	return kpi
}

func (m *Manager) Tmux() *TmuxBridge {
	return m.tmux
}

// Cleanup kills all tmux sessions managed by this manager.
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		_ = m.tmux.KillSession(s.ID)
	}
}

func shortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
