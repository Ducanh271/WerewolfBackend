package ws

import (
	"fmt"
	"testing"
	"time"

	"masoi/internal/transport/ws/wsclient"
)

// TestJoinDeadline_CountsOnceFromConnectNotPerMessage is S6: the old code
// recreated the 10 s join deadline on every message, so a connection that
// sent anything every few seconds, without ever joining, stayed open
// forever.
func TestJoinDeadline_CountsOnceFromConnectNotPerMessage(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.JoinTimeout = 400 * time.Millisecond })
	c := e.dial(nil)
	stop := time.Now().Add(700 * time.Millisecond)
	for time.Now().Before(stop) {
		_ = c.Send(ctx(t), map[string]any{"type": "teleport"})
		time.Sleep(80 * time.Millisecond)
	}
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("a connection that never joins must be closed once the join deadline passes, even while it keeps sending messages")
	}
}

// TestJoinAttempts_ClosedAfterTooManyFailedJoinOrRejoinMessages is S6 (typos,
// stray clients) and S7 (a PIN guessed across many nicknames on one
// connection): a connection that keeps failing to join or rejoin is closed
// after a few tries, instead of being left open to keep trying.
func TestJoinAttempts_ClosedAfterTooManyFailedJoinOrRejoinMessages(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.MaxJoinAttempts = 3 })
	c := e.dial(nil)
	for i := 0; i < 3; i++ {
		e.send(c, map[string]any{"type": "rejoin", "room": "ZZZZ", "token": "bogus"})
	}
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("a connection must close after too many failed join/rejoin attempts")
	}
}

// TestJoinAttempts_PINGuessesAcrossManyNicknamesOnOneConnectionAreBounded is
// S7's main case: trying many nicknames' PINs from one connection must not
// be unlimited, even though each nickname's own 5-per-minute lock (S7
// existing behaviour) is unaffected by the others.
func TestJoinAttempts_PINGuessesAcrossManyNicknamesOnOneConnectionAreBounded(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.MaxJoinAttempts = 4 })
	code := e.createRoom()
	for i := 0; i < 4; i++ {
		e.join(code, fmt.Sprintf("P%d", i)).Close()
	}
	c := e.dial(nil)
	for i := 0; i < 4; i++ {
		e.send(c, map[string]any{"type": "rejoin", "room": code, "nickname": fmt.Sprintf("P%d", i), "pin": "0000"})
	}
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("guessing PINs across 4 different nicknames on one connection must still hit the attempt cap")
	}
}

// TestConnLimits_GlobalCapRejectsNewConnections is S6: a hard ceiling on the
// total number of open WebSockets.
func TestConnLimits_GlobalCapRejectsNewConnections(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.MaxConnections = 2 })
	_ = e.dial(nil)
	_ = e.dial(nil)
	if _, err := wsclient.Dial(ctx(t), e.wsURL(), nil); err == nil {
		t.Fatal("a third connection should have been rejected by the global cap")
	}
}

// TestConnLimits_PerIPRateRejectsAFloodOfNewConnections is S6/S5's real
// control: not how many connections one IP may hold open (that cap is kept
// high, see below), but how fast it may open new ones.
func TestConnLimits_PerIPRateRejectsAFloodOfNewConnections(t *testing.T) {
	e := newEnv(t, func(c *Config) {
		c.NewConnPerMinute, c.NewConnBurst = 5, 5
		c.MaxConnectionsPerIP, c.MaxConnections = 1000, 1000
	})
	for i := 0; i < 5; i++ {
		_ = e.dial(nil) // consumes the burst
	}
	if _, err := wsclient.Dial(ctx(t), e.wsURL(), nil); err == nil {
		t.Fatal("a 6th connection within the same window should be rate limited")
	}
	e.now.advance(time.Minute) // the burst refills
	if _, err := wsclient.Dial(ctx(t), e.wsURL(), nil); err != nil {
		t.Fatalf("after a minute the rate should have recovered: %v", err)
	}
}

// TestConnLimits_FullRoomPlusReconnectsFromOneIPAllSucceed is the scenario
// the per-IP concurrent cap must never block: a table of players (who often
// share one public IP — café Wi-Fi, carrier NAT) joining, and later
// reconnecting, from what the server sees as a single address.
func TestConnLimits_FullRoomPlusReconnectsFromOneIPAllSucceed(t *testing.T) {
	e := newEnv(t, nil) // production defaults: 100 concurrent, 30/min new
	code := e.createRoom()
	tokens := make([]string, 30)
	for i := 0; i < 30; i++ {
		c := e.join(code, fmt.Sprintf("P%02d", i))
		tokens[i] = c.Snapshot().Welcome.Str("token")
		c.Close() // the tab closes; the token is kept, as a phone would
	}
	e.now.advance(time.Minute) // "a few minutes" later: the burst has refilled
	for i := 0; i < 30; i++ {
		back := e.dial(nil)
		e.send(back, map[string]any{"type": "rejoin", "room": code, "token": tokens[i]})
		if !back.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Welcome != nil }) {
			t.Fatalf("reconnect #%d from the same IP was refused: %v", i, back.Snapshot().Errors)
		}
	}
}

// TestConnLimits_FloodOfNewConnectionsFromOneIPIsRejected checks the default
// configuration actually rejects a burst beyond it, from one IP.
func TestConnLimits_FloodOfNewConnectionsFromOneIPIsRejected(t *testing.T) {
	e := newEnv(t, nil) // defaults: 30 new connections per minute, burst 30
	for i := 0; i < 30; i++ {
		_ = e.dial(nil)
	}
	if _, err := wsclient.Dial(ctx(t), e.wsURL(), nil); err == nil {
		t.Fatal("connection #31 within the same burst window should be rejected")
	}
}
