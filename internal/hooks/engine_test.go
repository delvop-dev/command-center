package hooks

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEngineStartStop(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	e := New(sockPath)
	if err := e.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if _, err := os.Stat(sockPath); err != nil {
		t.Fatalf("socket not created: %v", err)
	}
	e.Stop()
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("socket not cleaned up")
	}
}

func TestEngineReceiveEvent(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	e := New(sockPath)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()

	events := e.Subscribe()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	evt := HookEvent{SessionID: "test-session", Type: "state_change", Data: map[string]interface{}{"state": "working"}}
	json.NewEncoder(conn).Encode(evt)
	conn.Close()

	select {
	case received := <-events:
		if received.SessionID != "test-session" {
			t.Errorf("expected 'test-session', got %q", received.SessionID)
		}
		if received.Type != "state_change" {
			t.Errorf("expected 'state_change', got %q", received.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for event")
	}
}

func TestEngineSocketPath(t *testing.T) {
	e := New("/tmp/test-delvop.sock")
	if e.SocketPath() != "/tmp/test-delvop.sock" {
		t.Errorf("wrong socket path: %q", e.SocketPath())
	}
}

func TestEngineMultipleSubscribers(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	e := New(sockPath)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()

	ch1 := e.Subscribe()
	ch2 := e.Subscribe()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	evt := HookEvent{SessionID: "s1", Type: "test", Data: map[string]interface{}{}}
	json.NewEncoder(conn).Encode(evt)
	conn.Close()

	// Both subscribers should receive the event
	for i, ch := range []<-chan HookEvent{ch1, ch2} {
		select {
		case received := <-ch:
			if received.SessionID != "s1" {
				t.Errorf("subscriber %d: expected 's1', got %q", i, received.SessionID)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("subscriber %d: timed out", i)
		}
	}
}

func TestEngineBroadcastDropsWhenFull(t *testing.T) {
	e := New("/tmp/unused.sock")
	ch := e.Subscribe()

	// Fill up the channel buffer (64)
	for i := 0; i < 64; i++ {
		e.broadcast(HookEvent{SessionID: "fill", Type: "test"})
	}

	// This should not block; it should drop the event
	e.broadcast(HookEvent{SessionID: "dropped", Type: "test"})

	// Drain the channel
	count := 0
	for {
		select {
		case <-ch:
			count++
		default:
			goto done
		}
	}
done:
	if count != 64 {
		t.Errorf("expected 64 events, got %d", count)
	}
}

func TestEngineInvalidJSON(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	e := New(sockPath)
	if err := e.Start(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()

	ch := e.Subscribe()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	// Send invalid JSON
	conn.Write([]byte("this is not json\n"))
	conn.Close()

	// Should not receive any event
	select {
	case <-ch:
		t.Error("should not receive event for invalid JSON")
	case <-time.After(200 * time.Millisecond):
		// Expected: no event
	}
}

func TestEngineStopWithoutStart(t *testing.T) {
	e := New("/tmp/test-stop.sock")
	// Should not panic
	e.Stop()
}

func TestEngineStartRemovesExistingSocket(t *testing.T) {
	// Use /tmp directly to avoid macOS socket path length limits
	sockPath := "/tmp/delvop-test-existing.sock"
	os.Remove(sockPath) // clean up from previous runs

	// Create a real unix socket at the path first
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("failed to create existing socket: %v", err)
	}
	ln.Close()

	e := New(sockPath)
	if err := e.Start(); err != nil {
		t.Fatalf("Start failed with existing socket: %v", err)
	}
	defer e.Stop()
}
