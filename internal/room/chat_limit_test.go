package room

import (
	"errors"
	"testing"
	"time"

	"masoi/internal/engine"
)

// chatGame is five players in the Chief election, when everybody may write.
func chatGame(t *testing.T) *game {
	t.Helper()
	g := newGame(t, 5)
	g.deal()
	g.step(engine.DefaultSettings().RoleView)
	if s := g.snap(); s.Phase != engine.PhaseChiefElection {
		t.Fatalf("phase = %s, want ChiefElection", s.Phase)
	}
	return g
}

func (g *game) chat(p engine.PlayerID, text string) error {
	return g.cl[p].Submit(as(engine.SendChat{By: p, Text: text}))
}

func chatCount(log []engine.Event) int {
	n := 0
	for _, e := range log {
		if _, ok := e.(engine.ChatSent); ok {
			n++
		}
	}
	return n
}

// Rules §4 "Nhắn tin": one message per 2 seconds, burst 5. Each row is the
// time passed before a message and whether it goes through.
func TestChatLimit_OnePerTwoSecondsBurstFive(t *testing.T) {
	type try struct {
		after time.Duration
		want  error
	}
	burst := []try{{0, nil}, {0, nil}, {0, nil}, {0, nil}, {0, nil}}
	tests := []struct {
		name  string
		tries []try
	}{
		{"five at once pass, the sixth waits", append(burst[:5:5], try{0, ErrChatTooFast})},
		{"two seconds later one more passes", append(burst[:5:5], try{2 * time.Second, nil}, try{0, ErrChatTooFast})},
		{"just under two seconds is too soon", append(burst[:5:5], try{2*time.Second - time.Nanosecond, ErrChatTooFast})},
		{"refused tries do not push the wait back", append(burst[:5:5],
			try{time.Second, ErrChatTooFast}, try{time.Second, nil})},
		{"a long silence gives five again, never more", append(burst[:5:5],
			try{time.Minute / 2, nil}, try{0, nil}, try{0, nil}, try{0, nil}, try{0, nil}, try{0, ErrChatTooFast})},
		{"one every two seconds never hits the limit", []try{
			{0, nil}, {2 * time.Second, nil}, {2 * time.Second, nil}, {2 * time.Second, nil}, {2 * time.Second, nil},
			{2 * time.Second, nil}, {2 * time.Second, nil}, {2 * time.Second, nil}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := chatGame(t)
			accepted := 0
			for i, try := range tc.tries {
				g.step(try.after)
				err := g.chat("p2", "chào")
				if !errors.Is(err, try.want) {
					t.Fatalf("message %d: err = %v, want %v", i+1, err, try.want)
				}
				if err == nil {
					accepted++
				}
			}
			if got := chatCount(g.eventLog()); got != accepted {
				t.Fatalf("%d messages in the log, want %d: a refused one must leave no event", got, accepted)
			}
		})
	}
}

func TestChatLimit_EachPlayerHasTheirOwn(t *testing.T) {
	g := chatGame(t)
	for range ChatBurst {
		g.must(g.chat("p2", "chào"))
	}
	if err := g.chat("p2", "chào"); !errors.Is(err, ErrChatTooFast) {
		t.Fatalf("p2: err = %v, want ErrChatTooFast", err)
	}
	g.must(g.chat("p3", "chào"))
}

func TestChatLimit_ComingBackDoesNotRefillIt(t *testing.T) {
	tests := []struct {
		name string
		back func(g *game, sink Sink) (*Client, error)
	}{
		{"same phone, by token", func(g *game, sink Sink) (*Client, error) {
			return g.r.RejoinToken(g.cred["p2"].Token, sink)
		}},
		{"another phone, by nickname and PIN", func(g *game, sink Sink) (*Client, error) {
			cl, _, err := g.r.RejoinPIN(g.nick["p2"], g.cred["p2"].PIN, sink)
			return cl, err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := chatGame(t)
			for range ChatBurst {
				g.must(g.chat("p2", "chào"))
			}
			g.cl["p2"].Disconnect()
			cl, err := tc.back(g, &recSink{})
			g.must(err)
			g.cl["p2"] = cl
			if err := g.chat("p2", "chào"); !errors.Is(err, ErrChatTooFast) {
				t.Fatalf("after coming back: err = %v, want ErrChatTooFast", err)
			}
		})
	}
}

// A message the engine refuses spends nothing, and an empty one cannot use up
// a player's limit.
func TestChatLimit_RefusedMessagesSpendNothing(t *testing.T) {
	g := chatGame(t)
	for range 3 * ChatBurst {
		if err := g.chat("p2", "   "); !errors.Is(err, engine.ErrChatEmpty) {
			t.Fatalf("err = %v, want ErrChatEmpty", err)
		}
	}
	for range ChatBurst {
		g.must(g.chat("p2", "chào"))
	}
}

// The engine's refusal comes first: at night a player over the limit gets the
// same answer as everybody else (rules §3).
func TestChatLimit_AtNight_TheAnswerDoesNotDependOnPastMessages(t *testing.T) {
	g := chatGame(t)
	for range ChatBurst + 1 {
		_ = g.chat("p2", "chào")
	}
	g.step(engine.DefaultSettings().ChiefElection)
	if s := g.snap(); s.Phase != engine.PhaseNight {
		t.Fatalf("phase = %s, want Night", s.Phase)
	}
	for _, p := range []engine.PlayerID{"p2", "p3"} {
		if err := g.chat(p, "chào"); !errors.Is(err, engine.ErrChatClosed) {
			t.Fatalf("%s: err = %v, want ErrChatClosed", p, err)
		}
	}
}

func TestChatLimit_FullGame_NoMoreMessagesUntilANewGame(t *testing.T) {
	g := chatGame(t)
	_ = g.r.do(func() error { g.r.chat.sent = ChatMaxPerGame - 1; return nil })

	g.must(g.chat("p2", "tin cuối"))
	for _, p := range []engine.PlayerID{"p2", "p3"} {
		if err := g.chat(p, "chào"); !errors.Is(err, ErrChatFull) {
			t.Fatalf("%s: err = %v, want ErrChatFull", p, err)
		}
	}
	if got := chatCount(g.eventLog()); got != 1 {
		t.Fatalf("%d messages in the log, want 1", got)
	}

	g.forceEnded()
	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))
	g.deal()
	g.step(engine.DefaultSettings().RoleView)
	g.must(g.chat("p3", "ván mới"))
}

func TestChatLimit_CountsOnlyMessagesThatWentThrough(t *testing.T) {
	g := chatGame(t)
	g.must(g.chat("p2", "chào"))
	_ = g.chat("p2", "   ")
	for range ChatBurst + 2 {
		_ = g.chat("p3", "chào")
	}
	var sent int
	_ = g.r.do(func() error { sent = g.r.chat.sent; return nil })
	if want := 1 + ChatBurst; sent != want {
		t.Fatalf("counted %d messages, want %d", sent, want)
	}
}
