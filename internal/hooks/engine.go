package hooks

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"sync"
)

type HookEvent struct {
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

type Engine struct {
	socketPath  string
	listener    net.Listener
	subscribers []chan HookEvent
	mu          sync.RWMutex
	done        chan struct{}
}

func New(socketPath string) *Engine {
	return &Engine{
		socketPath: socketPath,
		done:       make(chan struct{}),
	}
}

func (e *Engine) Start() error {
	os.Remove(e.socketPath)
	ln, err := net.Listen("unix", e.socketPath)
	if err != nil {
		return err
	}
	e.listener = ln
	go e.acceptLoop()
	return nil
}

func (e *Engine) Stop() {
	close(e.done)
	if e.listener != nil {
		e.listener.Close()
	}
	os.Remove(e.socketPath)
}

func (e *Engine) Subscribe() <-chan HookEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	ch := make(chan HookEvent, 64)
	e.subscribers = append(e.subscribers, ch)
	return ch
}

func (e *Engine) SocketPath() string {
	return e.socketPath
}

func (e *Engine) acceptLoop() {
	for {
		conn, err := e.listener.Accept()
		if err != nil {
			select {
			case <-e.done:
				return
			default:
				log.Printf("hook accept error: %v", err)
				continue
			}
		}
		go e.handleConn(conn)
	}
}

func (e *Engine) handleConn(conn net.Conn) {
	defer conn.Close()
	var evt HookEvent
	if err := json.NewDecoder(conn).Decode(&evt); err != nil {
		return
	}
	e.broadcast(evt)
}

func (e *Engine) broadcast(evt HookEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, ch := range e.subscribers {
		select {
		case ch <- evt:
		default:
		}
	}
}
