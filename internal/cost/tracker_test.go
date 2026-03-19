package cost

import (
	"path/filepath"
	"testing"
)

func TestTrackerUpdate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	tracker := NewTracker(store)
	tracker.Update("session-1", 1.50, 10000, 5000)
	tracker.Update("session-1", 2.00, 15000, 7000)

	entry, ok := tracker.Get("session-1")
	if !ok {
		t.Fatal("expected entry")
	}
	if entry.CostUSD != 2.00 {
		t.Errorf("expected cost 2.00, got %f", entry.CostUSD)
	}
	if entry.TokensIn != 15000 {
		t.Errorf("expected tokensIn 15000, got %d", entry.TokensIn)
	}

	total := tracker.TotalCost()
	if total != 2.00 {
		t.Errorf("expected total 2.00, got %f", total)
	}
}

func TestTrackerMultipleSessions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	tracker := NewTracker(store)
	tracker.Update("s1", 1.00, 1000, 500)
	tracker.Update("s2", 2.50, 2000, 1000)

	total := tracker.TotalCost()
	if total != 3.50 {
		t.Errorf("expected total 3.50, got %f", total)
	}
}

func TestTrackerGetNonexistent(t *testing.T) {
	tracker := NewTracker(nil)
	_, ok := tracker.Get("nope")
	if ok {
		t.Error("expected false for nonexistent entry")
	}
}

func TestStoreUpsert(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	err = store.Upsert("s1", 1.50, 1000, 500)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Upsert("s1", 2.00, 2000, 1000)
	if err != nil {
		t.Fatal(err)
	}
}
