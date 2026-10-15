package enginetest

import (
	"masoi/internal/engine"
	"masoi/internal/roles"
)

// Cupid is Cupid: Cupid("Lan").Pairs("An", "Binh"). He acts on night 1 only.
func (n *Night) Cupid(name string) *CupidActor { return &CupidActor{n: n, id: n.Player(name)} }

// CupidActor is Cupid acting.
type CupidActor struct {
	n  *Night
	id engine.PlayerID
}

// Pairs makes the two players a couple. The lovers learn of it at the phone
// check: see PhoneCheck and LoverNotice.
func (a *CupidActor) Pairs(first, second string) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Cupid)
	return a.n.Do(roles.CupidPair{By: a.id, A: a.n.Player(first), B: a.n.Player(second)})
}

// CannotPair chooses and expects the engine to refuse with want.
func (a *CupidActor) CannotPair(first, second string, want error) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Cupid)
	return a.n.DoExpectErr(roles.CupidPair{By: a.id, A: a.n.Player(first), B: a.n.Player(second)}, want)
}

// RunsOutOfTime lets Cupid's step end without a choice; the server then pairs
// two random living players with the given seed (rules §7).
func (a *CupidActor) RunsOutOfTime(seed int64) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Cupid)
	return a.n.Do(engine.PhaseTimeout{Phase: a.n.g.state.Phase, Seq: a.n.g.state.Seq, Seed: seed})
}

// Lovers returns the couple of the game, in no particular order, as names.
func (g *Game) Lovers() (a, b string, ok bool) {
	if g.state.Lovers == nil {
		return "", "", false
	}
	return string(g.state.Lovers.A), string(g.state.Lovers.B), true
}

// LoveNotice checks the love part of the phone check notice the viewer sees.
func (s *Seen) LoveNotice(want engine.LoveNotice) *Seen {
	s.g.t.Helper()
	if s.view.Me == nil || s.view.Me.PhoneCheck == nil || s.view.Me.PhoneCheck.Love != want {
		s.g.t.Fatalf("%s should see the love notice %+v, sees %+v", s.who, want, s.view.Me)
	}
	return s
}

// Doomed is declared in view_knight.go; Grief checks the grief notice on the
// viewer's own turn.
func (s *Seen) Grief(want bool) *Seen {
	s.g.t.Helper()
	got := s.view.Me != nil && s.view.Me.Turn != nil && s.view.Me.Turn.Grief
	if got != want {
		s.g.t.Fatalf("%s: grief notice shown=%v, want %v", s.who, got, want)
	}
	return s
}
