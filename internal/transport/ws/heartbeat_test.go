package ws

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/coder/websocket"

	"masoi/internal/room"
	"masoi/internal/transport/ws/wsclient"
)

// pingClockFor is a FakeClock dedicated to one test's heartbeat, separate
// from env's room clock: sharing one clock between game-phase timers and the
// ping schedule would mean advancing it for a phase timeout could also cross
// the ping-timeout threshold in unrelated tests.
func pingClockFor(t *testing.T) *room.FakeClock {
	t.Helper()
	return room.NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
}

// silentJoin opens a raw connection and sends `join`, then never reads from
// it again — not even the welcome. A normal wsclient.Client always has a
// Read call in flight between messages, which answers pings on its own
// (that is the websocket library's job, not ours); the only way to simulate
// a phone that genuinely lost signal, so nothing it holds auto-answers a
// ping, is to never read at all. The TCP connection itself stays open, as a
// real network partition would leave it.
func (e *env) silentJoin(code, nickname string) *websocket.Conn {
	e.t.Helper()
	conn, _, err := websocket.Dial(ctx(e.t), e.wsURL(), nil)
	if err != nil {
		e.t.Fatalf("dial: %v", err)
	}
	e.t.Cleanup(func() { _ = conn.CloseNow() })
	b, err := json.Marshal(map[string]any{"type": "join", "room": code, "nickname": nickname})
	if err != nil {
		e.t.Fatal(err)
	}
	if err := conn.Write(context.Background(), websocket.MessageText, b); err != nil {
		e.t.Fatalf("send join: %v", err)
	}
	return conn
}

// TestHeartbeat_SilentConnectionIsDroppedAndSpeakerMoves is R1: a phone that
// lost signal still looks "connected" at the TCP level (nothing tells the
// server the socket is gone). The server's own ping must catch this; a
// dropped-connection error never comes on its own.
func TestHeartbeat_SilentConnectionIsDroppedAndSpeakerMoves(t *testing.T) {
	pingClk := pingClockFor(t)
	e := newEnv(t, func(c *Config) {
		c.PingInterval = time.Second
		c.PongTimeout = time.Second
		c.PingClock = pingClk
	})
	code := e.createRoom()
	e.silentJoin(code, "An")          // joins first: the host, and so the speaker
	time.Sleep(50 * time.Millisecond) // let the server process it before Binh joins
	other := e.join(code, "Binh")

	if other.Snapshot().View.Get("speaker") == true {
		t.Fatalf("Binh must not be the speaker yet: %v", other.Snapshot().View)
	}

	pingClk.Advance(time.Second)      // a ping goes out to both phones
	time.Sleep(50 * time.Millisecond) // Binh's phone answers for real; An's never will
	pingClk.Advance(time.Second)      // An's pong never came: dropped. Binh's did: still open

	if !other.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.View.Get("speaker") == true }) {
		t.Fatalf("the next connected player did not become the speaker: closed=%v errors=%v view=%v",
			other.Snapshot().Closed, other.Snapshot().Errors, other.Snapshot().View)
	}
}

// TestHeartbeat_RespondingConnectionStaysOpen is the companion case: as long
// as a connection answers pings, it must not be closed by the heartbeat. A
// short real pause follows each tick so the real (if near-instant) ping/pong
// round trip has landed before the fake clock is pushed to the next one.
func TestHeartbeat_RespondingConnectionStaysOpen(t *testing.T) {
	pingClk := pingClockFor(t)
	e := newEnv(t, func(c *Config) {
		c.PingInterval = 50 * time.Millisecond
		c.PongTimeout = 50 * time.Millisecond
		c.PingClock = pingClk
	})
	code := e.createRoom()
	c := e.join(code, "An")

	for i := 0; i < 5; i++ {
		pingClk.Advance(50 * time.Millisecond)
		time.Sleep(20 * time.Millisecond)
	}
	select {
	case <-c.Done():
		t.Fatal("a connection that keeps answering pings must not be closed")
	case <-time.After(100 * time.Millisecond):
	}
}
