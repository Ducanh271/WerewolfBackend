package room

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// Every phone gets only ViewFor and Visible for itself, at every moment.
func TestBroadcast_EachPhoneGetsOnlyItsOwnProjection(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toNight1()
	wolf := g.who(roles.Werewolf)
	victim := g.who(roles.Villager)
	for id := range g.cl {
		if id != wolf && id != victim {
			victim = id
			break
		}
	}
	g.step(engine.DefaultSettings().NightStep) // the Guard's call passes
	g.must(g.cl[wolf].Submit(func(by engine.PlayerID) engine.Command {
		return roles.WolfVote{By: by, Target: victim}
	}))
	for i := 0; i < 2; i++ { // the pack's call, then the Seer's
		g.step(engine.DefaultSettings().NightStep)
	}

	state, log := g.snap(), g.eventLog()
	for id, sink := range g.sink {
		viewer := engine.PlayerViewer(id)
		got := sink.last().View.View
		if want := engine.ViewFor(state, viewer); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: view differs from ViewFor", id)
		}
		if known, want := sink.events(), engine.Visible(state, log, viewer); !reflect.DeepEqual(known, want) {
			t.Errorf("%s knows %d events, Visible gives %d", id, len(known), len(want))
		}
	}

	// Concrete leaks, from the point of view of a plain villager.
	bystander := engine.PlayerID("")
	for id, p := range state.Players {
		if p.Alive && p.Role == roles.Villager && id != victim {
			bystander = id
		}
	}
	view := g.sink[bystander].last().View.View
	if view.Pack != nil || view.Secrets != nil || view.Full {
		t.Fatal("a villager got the pack or hidden state")
	}
	for _, e := range g.sink[bystander].events() {
		switch e.(type) {
		case engine.RolesDealt, engine.PlayerDied, roles.WolfVoted:
			t.Fatalf("a villager was sent %T", e)
		}
	}
	if view.NightCall != nil {
		t.Fatalf("night is over, but a call is shown: %+v", view.NightCall)
	}
	if g.sink[wolf].last().View.View.Pack == nil {
		t.Fatal("the wolf must see its pack")
	}
}

// A real call and a fake one reach every phone identically; only the called
// player's own card differs.
func TestBroadcast_NightCallIsTheSameOnEveryPhone(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toNight1()
	calls := map[engine.PlayerID]engine.NightCallView{}
	for id, sink := range g.sink {
		nc := sink.last().View.View.NightCall
		if nc == nil {
			t.Fatalf("%s sees no night call", id)
		}
		calls[id] = *nc
	}
	guard := g.who(roles.Guard)
	for id, c := range calls {
		if c != calls[guard] {
			t.Errorf("%s sees %+v, the Guard sees %+v", id, c, calls[guard])
		}
		turn := g.sink[id].last().View.View.Me.Turn
		if (turn != nil) != (id == guard) {
			t.Errorf("%s: turn card = %v", id, turn)
		}
	}
}

func TestBroadcast_DeadlineTellsWhenTheStepEnds(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	got := g.sink["p2"].last().View.Deadline
	want := g.clk.Now().Add(engine.DefaultSettings().RoleView)
	if !got.Equal(want) {
		t.Fatalf("deadline = %v, want %v", got, want)
	}
}
