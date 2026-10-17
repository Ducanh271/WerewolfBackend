package room

import (
	"errors"
	"testing"

	"masoi/internal/engine"
)

// forceEnded ends the game right away, however it got dealt, so tests do not
// have to play a whole night out just to reach NewGame (audit K3).
func (g *game) forceEnded() {
	g.t.Helper()
	_ = g.r.do(func() error {
		evs := []engine.Event{engine.GameEnded{Outcome: engine.Outcome{Draw: true}}}
		g.r.state = engine.Fold(g.r.state, evs...)
		g.r.log = append(g.r.log, evs...)
		return nil
	})
}

func TestClient_NewGame_OnlyTheHostMay(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.forceEnded()

	if err := g.cl["p2"].Submit(as(engine.NewGame{})); !errors.Is(err, ErrNotHost) {
		t.Fatalf("non-host NewGame: err = %v, want ErrNotHost", err)
	}
	if g.snap().Phase != engine.PhaseEnded {
		t.Fatal("a non-host's NewGame changed the game")
	}

	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))
	if s := g.snap(); s.Phase != engine.PhaseLobby || s.Dealt {
		t.Fatalf("after host NewGame: phase = %s, dealt = %v, want Lobby/false", s.Phase, s.Dealt)
	}
}

func TestClient_NewGame_RejectedBeforeTheGameEnded(t *testing.T) {
	g := newGame(t, 5)
	g.deal()

	if err := g.cl["p1"].Submit(as(engine.NewGame{})); !errors.Is(err, engine.ErrWrongPhase) {
		t.Fatalf("NewGame mid-game: err = %v, want ErrWrongPhase", err)
	}
}

// TestClient_NewGame_KeepsSeatsAndRoles_HostCanDealAgain covers audit K3
// end to end through the room layer, including the seed the room (not the
// client) draws for the second deal.
func TestClient_NewGame_KeepsSeatsAndRoles_HostCanDealAgain(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	seats, roles := g.snap().Seats, g.snap().Settings.Roles
	g.forceEnded()
	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))

	s := g.snap()
	if len(s.Seats) != len(seats) {
		t.Fatalf("seats = %v, want kept as %v", s.Seats, seats)
	}
	if len(s.Settings.Roles) != len(roles) {
		t.Fatalf("role list = %v, want kept as %v", s.Settings.Roles, roles)
	}
	g.rnd.setSeed(2)
	if err := g.cl["p1"].Submit(as(engine.DealRoles{})); err != nil {
		t.Fatalf("second DealRoles: %v, want no error", err)
	}
	if !g.snap().Dealt {
		t.Fatal("second deal did not happen")
	}
}
