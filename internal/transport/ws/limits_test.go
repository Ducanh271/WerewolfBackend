package ws

import (
	"strings"
	"testing"
	"time"

	"masoi/internal/transport/ws/wsclient"
)

func TestLimits_BigMessageGetsAnErrorAndTheConnectionStaysUsable(t *testing.T) {
	e := newEnv(t, nil)
	c := e.join(e.createRoom(), "An")

	big := `{"type":"command","name":"nominate","args":{"target":"` + strings.Repeat("x", 5000) + `"}}`
	if err := c.SendRaw(ctx(t), []byte(big)); err != nil {
		t.Fatal(err)
	}
	if got := e.waitError(c, 1); got != "message_too_large" {
		t.Fatalf("code = %q, want message_too_large", got)
	}
	e.send(c, cmd("nominate", map[string]any{"target": "p2"}))
	if got := e.waitError(c, 2); got != "wrong_phase" {
		t.Fatalf("after the big message: code = %q, want wrong_phase (the connection works)", got)
	}
}

func TestLimits_HugeMessageClosesTheConnection(t *testing.T) {
	e := newEnv(t, nil)
	c := e.join(e.createRoom(), "An")
	_ = c.SendRaw(ctx(t), []byte(strings.Repeat("x", 100<<10)))
	select {
	case <-c.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("a 100 KB message did not close the connection")
	}
}

func TestLimits_TooManyMessagesGetAnError(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.Burst = 5; c.RatePerSecond = 1 })
	c := e.join(e.createRoom(), "An") // the join used one of the 5

	for i := 0; i < 8; i++ {
		e.send(c, map[string]any{"type": "teleport"})
	}
	c.Wait(ctx(t), func(s wsclient.Snapshot) bool { return len(s.Errors) >= 8 })
	var limited int
	for _, f := range c.Snapshot().Errors {
		if f.Str("code") == "rate_limited" {
			limited++
		}
	}
	if limited != 4 { // 4 tokens left after the join, so 4 of the 8 pass
		t.Fatalf("rate_limited errors = %d, want 4; errors: %v", limited, c.Snapshot().Errors)
	}

	// A second later one more message is allowed again.
	e.now.advance(time.Second)
	e.send(c, map[string]any{"type": "teleport"})
	if got := e.waitError(c, 9); got != "unknown_type" {
		t.Fatalf("after waiting: code = %q, want unknown_type", got)
	}
}

func TestLimits_ContinuousFloodClosesTheConnection(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.Burst = 2; c.RatePerSecond = 1 })
	c := e.join(e.createRoom(), "An")

	// Over the limit in every second for more than 3 seconds. Each second's
	// messages are answered before the clock moves on, so the server sees the
	// same timeline the test describes.
	sent := 0
	for sec := 0; sec < 5; sec++ {
		for i := 0; i < 4; i++ {
			_ = c.Send(ctx(t), map[string]any{"type": "teleport"})
			sent++
		}
		c.Wait(ctx(t), func(s wsclient.Snapshot) bool { return len(s.Errors) >= sent })
		e.now.advance(time.Second)
	}
	select {
	case <-c.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("a flood did not close the connection")
	}
}

func TestLimits_ABurstThatStopsDoesNotCloseTheConnection(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.Burst = 2; c.RatePerSecond = 1 })
	c := e.join(e.createRoom(), "An")
	for i := 0; i < 5; i++ {
		e.send(c, map[string]any{"type": "teleport"})
	}
	e.now.advance(10 * time.Second)
	e.send(c, map[string]any{"type": "teleport"})
	e.waitError(c, 6)
	select {
	case <-c.Done():
		t.Fatal("a short burst closed the connection")
	default:
	}
}

func TestLimiter_FloodMeansRefusalsWithoutABreakForTheWholeWindow(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	l := newLimiter(1, 1, 3*time.Second, t0)
	l.allow(t0) // the only token

	for _, at := range []time.Duration{0, 1, 2} {
		if ok, flood := l.allow(t0.Add(at * time.Second)); ok && at == 0 || flood {
			t.Fatalf("at %ds: ok=%v flood=%v, too early to be a flood", at, ok, flood)
		}
		l.allow(t0.Add(at * time.Second)) // refused again in the same second
	}
	if _, flood := l.allow(t0.Add(3 * time.Second)); flood {
		// a token came back at 3s, so this one passed: not refused, not a flood
		t.Fatal("flood reported for an allowed message")
	}
	if ok, flood := l.allow(t0.Add(3 * time.Second)); ok || !flood {
		t.Fatalf("after 3 seconds over the limit: ok=%v flood=%v, want refused flood", ok, flood)
	}

	// A refusal, a long quiet break, another refusal: the run starts over.
	l = newLimiter(1, 1, 3*time.Second, t0)
	l.allow(t0)
	l.allow(t0)
	l.allow(t0.Add(time.Second))
	l.allow(t0.Add(30 * time.Second)) // passes
	if ok, flood := l.allow(t0.Add(30 * time.Second)); ok || flood {
		t.Fatalf("after a break: ok=%v flood=%v, want refused but not a flood", ok, flood)
	}
}
