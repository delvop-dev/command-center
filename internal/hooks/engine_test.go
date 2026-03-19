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
