package room

import (
	"errors"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
)

func TestSubmit_ActsOnlyAsTheConnectedPlayer(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	// p2's phone tries to nominate in p3's name.
	err := g.cl["p2"].Submit(as(engine.Nominate{Voter: "p3", Target: "p1"}))
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
	if len(g.snap().Day.Nominations) != 0 {
		t.Fatal("a forged command changed the game")
	}
	// The honest way: the room fills in who is acting.
	g.must(g.cl["p2"].Submit(func(by engine.PlayerID) engine.Command {
		return engine.Nominate{Voter: by, Target: "p1"}
	}))
	if _, ok := g.snap().Day.Nominations["p2"]; !ok {
		t.Fatal("p2's own nomination was not recorded under p2")
	}
}

func TestSetup_OnlyTheHostMayConfigureAndDeal(t *testing.T) {
	g := newGame(t, 5)
	order := []engine.PlayerID{"p1", "p2", "p3", "p4", "p5"}
	for name, c := range map[string]engine.Command{
		"seats": engine.SetSeats{Order: order}, "roles": engine.ConfigureRoles{}, "deal": engine.DealRoles{},
		"timers": engine.ConfigureTimers{}, "sound": engine.ConfigureSound{Mode: engine.AllPhonesSpeak},
	} {
		if err := g.cl["p2"].Submit(as(c)); !errors.Is(err, ErrNotHost) {
			t.Errorf("%s from a non-host: %v, want ErrNotHost", name, err)
		}
	}
	if g.snap().Dealt || len(g.snap().Seats) != 0 {
		t.Fatal("a non-host changed the setup")
	}
}

func TestSetup_LobbyCommandsAreRejectedAfterTheDeal(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	host := g.cl["p1"]
	for name, c := range map[string]engine.Command{
		"seats":  engine.SetSeats{Order: []engine.PlayerID{"p5", "p4", "p3", "p2", "p1"}},
		"roles":  engine.ConfigureRoles{Roles: map[engine.RoleID]int{"Villager": 5}},
		"deal":   engine.DealRoles{},
		"timers": engine.ConfigureTimers{Timers: engine.Timers{Day: 5 * time.Minute}},
		"sound":  engine.ConfigureSound{Mode: engine.AllPhonesSpeak},
	} {
		if err := host.Submit(as(c)); !errors.Is(err, engine.ErrWrongPhase) {
			t.Errorf("host %s after the deal: %v, want ErrWrongPhase", name, err)
		}
	}
	if err := host.Leave(); !errors.Is(err, engine.ErrWrongPhase) {
		t.Errorf("Leave after the deal: %v, want ErrWrongPhase", err)
	}
	if _, _, err := g.r.Join("Late", &recSink{}); !errors.Is(err, engine.ErrWrongPhase) {
		t.Errorf("Join after the deal: %v, want ErrWrongPhase", err)
	}
}

func TestRoom_OnlineModeIsNotSupportedYet(t *testing.T) {
	r, err := New(Config{Code: "X", Mode: engine.Online})
	if !errors.Is(err, engine.ErrModeUnsupported) || r != nil {
		t.Fatalf("New(Online) = %v, %v, want ErrModeUnsupported", r, err)
	}
}

type failingStore struct{ store.Memory }

func (*failingStore) Append(string, ...engine.Event) error { return errors.New("disk full") }

func TestRoom_EventsAreStoredBeforeAnyoneIsTold(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	stored, _ := g.store.Load("K7QP")
	if got := g.eventLog(); len(stored) != len(got) || len(got) == 0 {
		t.Fatalf("store has %d events, the room %d", len(stored), len(got))
	}

	r, _ := New(Config{Code: "F", Store: &failingStore{}})
	defer r.Close()
	sink := &recSink{}
	if _, _, err := r.Join("An", sink); err == nil {
		t.Fatal("join succeeded although the event could not be stored")
	}
	if len(sink.msgs) != 0 {
		t.Fatal("somebody was told about an event that was not stored")
	}
}
