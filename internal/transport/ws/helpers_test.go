package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"masoi/internal/engine"
	_ "masoi/internal/roles"
	"masoi/internal/room"
	"masoi/internal/transport/ws/wsclient"
)

// syncBuffer is a log destination safe for the server's goroutines.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// fakeNow is the rate limiter's clock.
type fakeNow struct {
	mu sync.Mutex
	t  time.Time
}

func (f *fakeNow) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.t
}

func (f *fakeNow) advance(d time.Duration) {
	f.mu.Lock()
	f.t = f.t.Add(d)
	f.mu.Unlock()
}

// env is a running server with a fake room clock and a fake limiter clock.
type env struct {
	t     *testing.T
	srv   *httptest.Server
	rooms *room.Manager
	clk   *room.FakeClock
	now   *fakeNow
	logs  *syncBuffer
}

func newEnv(t *testing.T, mod func(*Config)) *env {
	t.Helper()
	e := &env{
		t:    t,
		clk:  room.NewFakeClock(time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)),
		now:  &fakeNow{t: time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)},
		logs: &syncBuffer{},
	}
	e.rooms = room.NewManager(room.ManagerConfig{Clock: e.clk})
	cfg := Config{
		// PingClock is left at its default (room.RealClock{}): the heartbeat's
		// real 12s/10s defaults never fire within a test's real run time. Tests
		// of the heartbeat itself (heartbeat_test.go) set PingClock to e.clk
		// explicitly; sharing it here by default would mean any test that
		// advances e.clk for an unrelated reason (a game timer) could also
		// cross the ping-timeout threshold and close connections spuriously.
		Rooms: e.rooms, AllowAnyOrigin: true, Now: e.now.Now,
		Logger: slog.New(slog.NewTextHandler(e.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}
	if mod != nil {
		mod(&cfg)
	}
	e.srv = httptest.NewServer(New(cfg))
	t.Cleanup(func() {
		e.srv.Close()
		e.rooms.Close()
	})
	return e
}

func (e *env) wsURL() string { return "ws" + strings.TrimPrefix(e.srv.URL, "http") + "/ws" }

// createRoom does POST /rooms and returns the code.
func (e *env) createRoom() string {
	code, _ := e.createRoomWithHost()
	return code
}

// createRoomWithHost does POST /rooms and returns the code and the creator's
// secret hostKey (audit S8): only the join that presents it becomes host.
func (e *env) createRoomWithHost() (code, hostKey string) {
	e.t.Helper()
	resp, err := http.Post(e.srv.URL+"/rooms", "application/json", nil)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct{ Room, HostKey string }
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || resp.StatusCode != http.StatusCreated {
		e.t.Fatalf("POST /rooms: status %d, err %v", resp.StatusCode, err)
	}
	return body.Room, body.HostKey
}

func ctx(t *testing.T) context.Context {
	t.Helper()
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return c
}

func (e *env) dial(header http.Header) *wsclient.Client {
	e.t.Helper()
	c, err := wsclient.Dial(ctx(e.t), e.wsURL(), header)
	if err != nil {
		e.t.Fatalf("dial: %v", err)
	}
	e.t.Cleanup(c.Close)
	return c
}

// join connects a phone and joins the room; it waits for the welcome.
func (e *env) join(code, nickname string) *wsclient.Client {
	return e.joinAs(code, "", nickname)
}

// joinHost is join, presenting the room's hostKey (audit S8).
func (e *env) joinHost(code, hostKey, nickname string) *wsclient.Client {
	return e.joinAs(code, hostKey, nickname)
}

func (e *env) joinAs(code, hostKey, nickname string) *wsclient.Client {
	e.t.Helper()
	c := e.dial(nil)
	msg := map[string]any{"type": "join", "room": code, "nickname": nickname}
	if hostKey != "" {
		msg["hostKey"] = hostKey
	}
	e.send(c, msg)
	if !c.Wait(ctx(e.t), func(s wsclient.Snapshot) bool { return s.Welcome != nil }) {
		e.t.Fatalf("%s: no welcome; errors: %v", nickname, c.Snapshot().Errors)
	}
	return c
}

func (e *env) send(c *wsclient.Client, v any) {
	e.t.Helper()
	if err := c.Send(ctx(e.t), v); err != nil {
		e.t.Fatalf("send: %v", err)
	}
}

// waitError waits for the n-th error message and returns its code.
func (e *env) waitError(c *wsclient.Client, n int) string {
	e.t.Helper()
	if !c.Wait(ctx(e.t), func(s wsclient.Snapshot) bool { return len(s.Errors) >= n }) {
		e.t.Fatalf("no error #%d; frames: %v", n, c.Snapshot().Frames)
	}
	return c.Snapshot().Errors[n-1].Str("code")
}

func cmd(name string, args map[string]any) map[string]any {
	return map[string]any{"type": "command", "name": name, "args": args}
}

// wireCommand decodes a `command` message the way the server does, for the
// player-free part: it returns the builder that fills in the sender.
func wireCommand(t *testing.T, name string, args json.RawMessage) (func(engine.PlayerID) engine.Command, error) {
	t.Helper()
	return buildCommand(commandMsg{Type: typeCommand, Name: name, Args: args})
}

// wireCommandNames lists every command name the server accepts.
func wireCommandNames(t *testing.T) []string {
	t.Helper()
	names := make([]string, 0, len(commandTable))
	for n := range commandTable {
		names = append(names, n)
	}
	return names
}
