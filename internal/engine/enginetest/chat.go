package enginetest

import (
	"reflect"

	"masoi/internal/engine"
)

// Chat is the player's message on the public channel (rules §4 "Nhắn tin").
func (g *Game) Chat(name, text string) *Game {
	g.t.Helper()
	g.Do(engine.SendChat{By: g.Player(name), Text: text})
	return g
}

// ChatExpectErr sends a message that must be refused with want.
func (g *Game) ChatExpectErr(name, text string, want error) *Game {
	g.t.Helper()
	g.DoExpectErr(engine.SendChat{By: engine.PlayerID(name), Text: text}, want)
	return g
}

// ChatReads checks every message this phone can read, in order.
func (s *Seen) ChatReads(want ...engine.ChatSent) *Seen {
	s.g.t.Helper()
	var got []engine.ChatSent
	for _, e := range s.feed {
		if m, ok := e.(engine.ChatSent); ok {
			got = append(got, m)
		}
	}
	if len(got) != len(want) || (len(got) > 0 && !reflect.DeepEqual(got, want)) {
		s.g.t.Fatalf("%s reads the chat as %+v, want %+v", s.who, got, want)
	}
	return s
}
