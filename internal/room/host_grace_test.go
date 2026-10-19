package room

import (
	"testing"
	"time"

	"masoi/internal/engine"
)

// Rules §2 "Chủ phòng": from the deal on, a dead phone keeps the title; once
// the game has ended the 60-second rule applies again, counted from when the
// host's phone dropped, so the table is never stuck without anyone able to
// press "Ván mới".

func TestHost_LostMidGame_PassesAsSoonAsTheGameEnds(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.cl["p1"].Disconnect()
	g.step(2 * hostDisconnectGrace)
	if h := g.host(); h != "p1" {
		t.Fatalf("host mid-game = %s, want p1 (a dead phone keeps the title until the end)", h)
	}

	g.forceEnded()
	g.step(0)

	if h := g.host(); h != "p2" {
		t.Fatalf("host after the end = %s, want p2 (first to join among the connected)", h)
	}
	g.must(g.cl["p2"].Submit(as(engine.NewGame{})))
}

func TestHost_DroppedJustBeforeTheEnd_GetsTheRestOfTheMinute(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.cl["p1"].Disconnect()
	g.step(50 * time.Second)
	g.forceEnded()

	g.step(9 * time.Second) // 59 seconds since the drop
	if h := g.host(); h != "p1" {
		t.Fatalf("host 59 s after the drop = %s, want p1", h)
	}
	g.step(time.Second) // 60
	if h := g.host(); h != "p2" {
		t.Fatalf("host 60 s after the drop = %s, want p2", h)
	}
}

func TestHost_DroppedAfterTheEnd_ReconnectingWithinTheMinute_KeepsTheTitle(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.forceEnded()
	g.cl["p1"].Disconnect()
	g.step(30 * time.Second)

	cl, err := g.r.RejoinToken(g.cred["p1"].Token, g.sink["p1"])
	if err != nil {
		t.Fatal(err)
	}
	g.cl["p1"] = cl
	g.step(hostDisconnectGrace) // a still-armed timer would have fired by now

	if h := g.host(); h != "p1" {
		t.Fatalf("host = %s, want p1 (back within the minute)", h)
	}
	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))
}

func TestHost_DroppedAfterTheEnd_StayingAway_PassesAfterTheMinute(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.forceEnded()
	g.cl["p1"].Disconnect()
	g.cl["p2"].Disconnect()

	g.step(hostDisconnectGrace)

	if h := g.host(); h != "p3" {
		t.Fatalf("host = %s, want p3 (first to join among the connected)", h)
	}
}
