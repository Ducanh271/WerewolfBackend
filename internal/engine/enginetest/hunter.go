package enginetest

import "masoi/internal/engine"

// Hunter is the dead Hunter's phone while he picks his target:
// g.Hunter("An").Shoots("Binh").
func (g *Game) Hunter(name string) *HunterActor { return &HunterActor{g: g, id: g.Player(name)} }

// HunterActor is the dead Hunter acting.
type HunterActor struct {
	g  *Game
	id engine.PlayerID
}

// Shoots takes the player with him.
func (a *HunterActor) Shoots(target string) *Game {
	a.g.t.Helper()
	a.g.Do(engine.HunterShoot{By: a.id, Target: a.g.Player(target)})
	return a.g
}

// CannotShoot shoots and expects the engine to refuse with want.
func (a *HunterActor) CannotShoot(target engine.PlayerID, want error) *Game {
	a.g.t.Helper()
	a.g.DoExpectErr(engine.HunterShoot{By: a.id, Target: target}, want)
	return a.g
}

// CanShoot checks whether the viewer's screen shows the shoot buttons.
func (s *Seen) CanShoot(want bool) *Seen {
	s.g.t.Helper()
	got := s.view.Me != nil && s.view.Me.CanShoot
	if got != want {
		s.g.t.Fatalf("%s: shoot buttons shown=%v, want %v", s.who, got, want)
	}
	return s
}

// Shooter checks whom every phone shows as choosing a target ("" for nobody).
func (s *Seen) Shooter(name string) *Seen {
	s.g.t.Helper()
	if s.view.Shooter != engine.PlayerID(name) {
		s.g.t.Fatalf("%s sees %q as the shooter, want %q", s.who, s.view.Shooter, name)
	}
	return s
}
