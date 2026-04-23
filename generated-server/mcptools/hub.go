package mcptools

// WebSocket hub that bridges MCP tool calls to connected SPA instances.
//
// One hub per process. SPA tabs open a WebSocket to the hub, send a
// "hello" frame with their session id + target metadata, and then
// block waiting for request frames. Tool handlers pick a session,
// send a JSON request with a unique id, and wait for a matching
// response frame (or timeout).
//
// All tool-call payloads are opaque JSON — the hub only routes them.

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Session struct {
	ID          string            `json:"id"`
	Target      map[string]any    `json:"target,omitempty"` // {owner, repo, branch, path}
	ConnectedAt time.Time         `json:"connectedAt"`
	UserAgent   string            `json:"userAgent,omitempty"`
	conn        *websocket.Conn
	writeMu     sync.Mutex
	pending     sync.Map // reqID → chan json.RawMessage
}

type inbound struct {
	Type      string          `json:"type"`
	SessionID string          `json:"sessionId,omitempty"`
	Target    map[string]any  `json:"target,omitempty"`
	RequestID string          `json:"requestId,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type outbound struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId"`
	Method    string          `json:"method"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

var (
	sessions   sync.Map // sessionID → *Session
	upgrader   = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	reqCounter atomic.Uint64
)

// StartWebSocketHub spawns an HTTP listener on addr that accepts
// WebSocket upgrades on /ws. Non-blocking; logs errors to stdout.
func StartWebSocketHub(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleWS)
	go func() {
		log.Printf("SPA websocket hub listening on %s (ws://localhost%s/ws)", addr, addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("websocket hub: %v", err)
		}
	}()
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	defer conn.Close()

	// First frame must be {type:"hello", sessionId, target?}.
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	var hello inbound
	if err := conn.ReadJSON(&hello); err != nil {
		log.Printf("ws hello read: %v", err)
		return
	}
	conn.SetReadDeadline(time.Time{})
	if hello.Type != "hello" || hello.SessionID == "" {
		log.Printf("ws rejected (type=%q sessionId=%q)", hello.Type, hello.SessionID)
		return
	}
	sess := &Session{
		ID:          hello.SessionID,
		Target:      hello.Target,
		ConnectedAt: time.Now().UTC(),
		UserAgent:   r.Header.Get("User-Agent"),
		conn:        conn,
	}
	// If a previous session with the same id is still registered, drop it.
	if prev, ok := sessions.LoadAndDelete(sess.ID); ok {
		if old, _ := prev.(*Session); old != nil && old.conn != nil {
			old.conn.Close()
		}
	}
	sessions.Store(sess.ID, sess)
	log.Printf("ws session connected: %s (target=%v)", sess.ID, sess.Target)
	defer func() {
		sessions.Delete(sess.ID)
		log.Printf("ws session disconnected: %s", sess.ID)
	}()

	// Read loop: dispatch responses to pending calls.
	for {
		var msg inbound
		if err := conn.ReadJSON(&msg); err != nil {
			if !errors.Is(err, websocket.ErrCloseSent) {
				log.Printf("ws read (%s): %v", sess.ID, err)
			}
			return
		}
		switch msg.Type {
		case "response":
			if ch, ok := sess.pending.LoadAndDelete(msg.RequestID); ok {
				ch.(chan inbound) <- msg
			}
		case "target":
			// SPA can update its target (user changed owner/repo, etc).
			sess.Target = msg.Target
		default:
			log.Printf("ws unknown message type %q from %s", msg.Type, sess.ID)
		}
	}
}

// CallSession sends an opaque JSON request to the named session and
// waits up to timeout for a matching response. Returns the raw
// payload bytes (caller JSON-decodes).
func CallSession(sessionID, method string, payload any, timeout time.Duration) (json.RawMessage, error) {
	v, ok := sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("no such session: %s", sessionID)
	}
	sess := v.(*Session)

	reqID := strconv.FormatUint(reqCounter.Add(1), 10)
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	ch := make(chan inbound, 1)
	sess.pending.Store(reqID, ch)

	sess.writeMu.Lock()
	err = sess.conn.WriteJSON(outbound{
		Type:      "request",
		RequestID: reqID,
		Method:    method,
		Payload:   raw,
	})
	sess.writeMu.Unlock()
	if err != nil {
		sess.pending.Delete(reqID)
		return nil, fmt.Errorf("ws write: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != "" {
			return nil, fmt.Errorf("session error: %s", resp.Error)
		}
		return resp.Payload, nil
	case <-time.After(timeout):
		sess.pending.Delete(reqID)
		return nil, fmt.Errorf("timeout after %s waiting for session %s", timeout, sessionID)
	}
}

// PickSession returns the single connected session id if exactly one
// is connected, otherwise an error that names all connected ids so
// the caller (or the LLM) can pick explicitly.
func PickSession() (string, error) {
	var ids []string
	sessions.Range(func(k, _ any) bool {
		ids = append(ids, k.(string))
		return true
	})
	if len(ids) == 1 {
		return ids[0], nil
	}
	if len(ids) == 0 {
		return "", fmt.Errorf("no SPA sessions connected — open the web UI and click 'Connect local MCP'")
	}
	return "", fmt.Errorf("multiple sessions connected (%v); specify session_id", ids)
}

// ListSessions returns a snapshot of connected sessions.
func ListSessions() []Session {
	var out []Session
	sessions.Range(func(_, v any) bool {
		s := v.(*Session)
		out = append(out, Session{
			ID:          s.ID,
			Target:      s.Target,
			ConnectedAt: s.ConnectedAt,
			UserAgent:   s.UserAgent,
		})
		return true
	})
	return out
}
