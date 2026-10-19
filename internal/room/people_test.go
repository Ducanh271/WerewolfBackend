package room

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
)

// The lobby and end screens list who is in the room (plan T30e), so a phone
// that comes back needs nothing but its latest view.

func TestView_Lobby_ListsEverybodyInTheRoom(t *testing.T) {
	g := newGame(t, 3)
	g.cl["p3"].Disconnect()

	want := []Person{
		{ID: "p1", Nickname: "An", Connected: true, IsHost: true},
		{ID: "p2", Nickname: "Binh", Connected: true},
		{ID: "p3", Nickname: "Chi"},
	}
	for _, p := range []engine.PlayerID{"p1", "p2"} {
		if got := g.sink[p].last().View.People; !reflect.DeepEqual(got, want) {
			t.Errorf("%s sees people %+v, want %+v", p, got, want)
		}
	}

	cl, err := g.r.RejoinToken(g.cred["p3"].Token, &recSink{})
	if err != nil {
		t.Fatal(err)
	}
	want[2].Connected = true
	if got := g.sink["p1"].last().View.People; !reflect.DeepEqual(got, want) {
		t.Errorf("after p3 is back, p1 sees %+v, want %+v", got, want)
	}
	g.cl["p3"] = cl
}

func TestView_People_OnlyInTheLobbyAndOnceTheGameHasEnded(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	if got := g.sink["p2"].last().View.People; got != nil {
		t.Fatalf("people during the game = %+v, want none", got)
	}
	g.cl["p1"].Disconnect()
	g.forceEnded()

	got := g.sink["p2"].last().View.People
	if len(got) != 5 || got[0].Connected || !got[0].IsHost {
		t.Fatalf("people at the end = %+v, want all five, p1 the host and disconnected", got)
	}
}
