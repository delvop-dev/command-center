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

func TestTrackerUpdateWithNilStore(t *testing.T) {
	tracker := NewTracker(nil)
	// Should not panic
	tracker.Update("s1", 1.00, 100, 50)

	entry, ok := tracker.Get("s1")
	if !ok {
		t.Fatal("expected entry with nil store")
	}
	if entry.CostUSD != 1.00 {
		t.Errorf("expected cost 1.00, got %f", entry.CostUSD)
	}
	if entry.SessionID != "s1" {
		t.Errorf("expected session ID 's1', got %q", entry.SessionID)
	}
}

func TestTrackerTotalCostEmpty(t *testing.T) {
	tracker := NewTracker(nil)
	total := tracker.TotalCost()
	if total != 0.0 {
		t.Errorf("expected total 0.0, got %f", total)
	}
}

func TestTrackerUpdateOverwrite(t *testing.T) {
	tracker := NewTracker(nil)
	tracker.Update("s1", 1.00, 100, 50)
	tracker.Update("s1", 5.00, 500, 250)

	entry, ok := tracker.Get("s1")
	if !ok {
		t.Fatal("expected entry")
	}
	if entry.CostUSD != 5.00 {
		t.Errorf("expected cost 5.00, got %f", entry.CostUSD)
	}
	if entry.TokensIn != 500 {
		t.Errorf("expected tokensIn 500, got %d", entry.TokensIn)
	}
	if entry.TokensOut != 250 {
		t.Errorf("expected tokensOut 250, got %d", entry.TokensOut)
	}

	total := tracker.TotalCost()
	if total != 5.00 {
		t.Errorf("expected total 5.00, got %f", total)
	}
}

func TestNewStoreCreatesTable(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "new.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Verify table exists by inserting
	err = store.Upsert("test", 1.0, 100, 50)
	if err != nil {
		t.Fatalf("failed to upsert into new store: %v", err)
	}
}

func TestNewStoreInvalidPath(t *testing.T) {
	// Use an invalid DSN to trigger sql.Open success but Exec failure
	// SQLite will fail on a directory path
	_, err := NewStore("/dev/null/impossible.db")
	if err == nil {
		t.Error("expected error for invalid db path")
	}
}

func TestStoreClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "close.db")
	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
	// Upsert after close should fail
	err = store.Upsert("test", 1.0, 100, 50)
	if err == nil {
		t.Error("expected error after Close()")
	}
}
