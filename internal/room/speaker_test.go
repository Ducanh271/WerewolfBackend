package room

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
)

func TestInPersonSpeaker_HostThenNextConnectedBySeat(t *testing.T) {
	order := []engine.PlayerID{"a", "b", "c", "d"}
	up := func(ids ...engine.PlayerID) func(engine.PlayerID) bool {
		return func(p engine.PlayerID) bool {
			for _, id := range ids {
				if id == p {
					return true
				}
			}
			return false
		}
	}
	tests := []struct {
		name      string
		host      engine.PlayerID
		connected []engine.PlayerID
		want      []engine.PlayerID
	}{
		{"host is connected", "b", []engine.PlayerID{"a", "b", "c", "d"}, []engine.PlayerID{"b"}},
		{"host gone: next by seat", "b", []engine.PlayerID{"a", "c", "d"}, []engine.PlayerID{"c"}},
		{"host gone, next gone too", "b", []engine.PlayerID{"a", "d"}, []engine.PlayerID{"d"}},
		{"wraps around the circle", "d", []engine.PlayerID{"a", "b"}, []engine.PlayerID{"a"}},
		{"nobody connected", "a", nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InPersonSpeaker{}.Speakers(Table{Host: tc.host, Order: order, Connected: up(tc.connected...)})
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("speakers = %v, want %v", got, tc.want)
			}
		})
	}
}

func speakerFlags(g *game) map[engine.PlayerID]bool {
	out := map[engine.PlayerID]bool{}
	for id, sink := range g.sink {
		if g.cl[id] != nil && g.liveClient(id) {
			out[id] = sink.last().View.Speaker
		}
	}
	return out
}

func (g *game) liveClient(id engine.PlayerID) bool {
	live := false
	_ = g.r.do(func() error { live = g.r.conns[id] != nil; return nil })
	return live
}

func onlySpeaker(t *testing.T, g *game, want engine.PlayerID) {
	t.Helper()
	for id, speaks := range speakerFlags(g) {
		if speaks != (id == want) {
			t.Fatalf("speaker flags = %v, want only %s", speakerFlags(g), want)
		}
	}
}

func TestSpeaker_HostSpeaksAndPassesBySeatWhenDisconnected(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	onlySpeaker(t, g, "p1")
	g.cl["p1"].Disconnect()
	onlySpeaker(t, g, "p2")
	g.cl["p2"].Disconnect()
	onlySpeaker(t, g, "p3")

	cl, err := g.r.RejoinToken(g.cred["p1"].Token, g.sink["p1"])
	if err != nil {
		t.Fatal(err)
	}
	g.cl["p1"] = cl
	onlySpeaker(t, g, "p1") // after the deal the host stays the host
}

func TestHost_PassesToTheNextPlayerBeforeTheDeal(t *testing.T) {
	g := newGame(t, 5)
	g.cl["p1"].Disconnect() // lost connection: the seat is kept, the powers pass
	order := []engine.PlayerID{"p1", "p2", "p3", "p4", "p5"}
	if err := g.cl["p1"].Submit(as(engine.SetSeats{Order: order})); err == nil {
		t.Fatal("a disconnected phone must not act")
	}
	g.must(g.cl["p2"].Submit(as(engine.SetSeats{Order: order})))
	if _, ok := g.snap().Players["p1"]; !ok {
		t.Fatal("a dropped connection must keep the player's place")
	}

	g.must(g.cl["p2"].Leave()) // leaving for good removes the player
	if _, ok := g.snap().Players["p2"]; ok {
		t.Fatal("p2 should be gone")
	}
	if len(g.snap().Seats) != 0 {
		t.Fatal("seats must be redone after someone leaves")
	}
	onlySpeaker(t, g, "p3")
	g.must(g.cl["p3"].Submit(as(engine.SetSeats{Order: []engine.PlayerID{"p1", "p3", "p4", "p5"}})))
}
