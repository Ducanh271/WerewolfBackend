package ws

import (
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

func TestRejoin_DroppedConnectionKeepsThePlayerAndTheSecretCodeComesBack(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	host := e.join(code, "An")
	e.join(code, "Binh")
	token := host.Snapshot().Welcome.Str("token")

	host.Close() // the phone loses signal
	select {
	case <-host.Done():
	case <-ctx(t).Done():
		t.Fatal("connection did not end")
	}

	// The player is still at the table: the nickname stays taken.
	third := e.dial(nil)
	e.send(third, map[string]any{"type": "join", "room": code, "nickname": "An"})
	if got := e.waitError(third, 1); got != "nickname_taken" {
		t.Fatalf("joining as the dropped player: code = %q, want nickname_taken", got)
	}

	back := e.dial(nil)
	e.send(back, map[string]any{"type": "rejoin", "room": code, "token": token})
	if !back.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Welcome != nil && s.View != nil && len(s.Events) > 0 }) {
		t.Fatalf("rejoin: no welcome, view and events: %v", back.Snapshot().Errors)
	}
	s := back.Snapshot()
	if s.Welcome.Str("player") != "p1" || s.View.Str("view", "me", "id") != "p1" {
		t.Fatalf("rejoined as %q, want p1", s.Welcome.Str("player"))
	}
	if s.Welcome.Str("token") != "" || s.Welcome.Str("pin") != "" {
		t.Fatalf("rejoin by token must not hand out new secrets: %v", s.Welcome)
	}
	if s.Frames[0].Type() != "welcome" {
		t.Fatalf("first message = %q, want welcome", s.Frames[0].Type())
	}
}

func TestRejoin_OtherPhoneWithNicknameAndPINGetsANewTokenAndTheOldPhoneIsOut(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	old := e.join(code, "An")
	e.join(code, "Binh")
	w := old.Snapshot().Welcome

	fresh := e.dial(nil)
	e.send(fresh, map[string]any{"type": "rejoin", "room": code, "nickname": "An", "pin": w.Str("pin")})
	fresh.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Welcome != nil })
	nw := fresh.Snapshot().Welcome
	if nw.Str("token") == "" || nw.Str("token") == w.Str("token") {
		t.Fatalf("new token = %q, old %q", nw.Str("token"), w.Str("token"))
	}

	// The old phone can no longer act, and is closed.
	e.send(old, cmd("nominate", map[string]any{"target": "p2"}))
	if got := e.waitError(old, 1); got != "disconnected" {
		t.Fatalf("old phone: code = %q, want disconnected", got)
	}
	select {
	case <-old.Done():
	case <-ctx(t).Done():
		t.Fatal("the replaced phone stayed connected")
	}

	// The old token stopped working.
	stale := e.dial(nil)
	e.send(stale, map[string]any{"type": "rejoin", "room": code, "token": w.Str("token")})
	if got := e.waitError(stale, 1); got != "bad_credentials" {
		t.Fatalf("old token: code = %q, want bad_credentials", got)
	}
}

func TestRejoin_FiveWrongPINsLockWithRetryAfter(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	e.join(code, "An")
	c := e.dial(nil)
	for i := 1; i <= 5; i++ {
		e.send(c, map[string]any{"type": "rejoin", "room": code, "nickname": "An", "pin": "0000"})
		e.waitError(c, i)
		e.now.advance(200 * 1000 * 1000) // stay under the rate limit
	}
	e.send(c, map[string]any{"type": "rejoin", "room": code, "nickname": "An", "pin": "0000"})
	if got := e.waitError(c, 6); got != "pin_locked" {
		t.Fatalf("code = %q, want pin_locked", got)
	}
	if ra, _ := c.Snapshot().Errors[5]["retryAfter"].(float64); ra <= 0 {
		t.Fatalf("pin_locked has no retryAfter: %v", c.Snapshot().Errors[5])
	}
}

func TestRejoin_LeaveRemovesThePlayerForGood(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	e.join(code, "An")
	b := e.join(code, "Binh")
	e.send(b, map[string]any{"type": "leave"})
	select {
	case <-b.Done():
	case <-ctx(t).Done():
		t.Fatal("leave did not close the connection")
	}
	again := e.join(code, "Binh") // the nickname is free again
	if again.Snapshot().Welcome.Str("player") == b.Snapshot().Welcome.Str("player") {
		t.Fatal("a player who left got the same id back")
	}
}
