package cost

import "sync"

type Entry struct {
	SessionID string
	CostUSD   float64
	TokensIn  int64
	TokensOut int64
}

type Tracker struct {
	entries map[string]*Entry
	store   *Store
	mu      sync.RWMutex
}

func NewTracker(store *Store) *Tracker {
	return &Tracker{
		entries: make(map[string]*Entry),
		store:   store,
	}
}

func (t *Tracker) Update(sessionID string, costUSD float64, tokensIn, tokensOut int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries[sessionID] = &Entry{
		SessionID: sessionID,
		CostUSD:   costUSD,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
	}
	if t.store != nil {
		t.store.Upsert(sessionID, costUSD, tokensIn, tokensOut)
	}
}

func (t *Tracker) Get(sessionID string) (*Entry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.entries[sessionID]
	return e, ok
}

func (t *Tracker) TotalCost() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	total := 0.0
	for _, e := range t.entries {
		total += e.CostUSD
	}
	return total
}
