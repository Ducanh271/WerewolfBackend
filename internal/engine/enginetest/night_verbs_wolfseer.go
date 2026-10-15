package enginetest

import (
	"masoi/internal/engine"
	"masoi/internal/roles"
)

// WolfSeer is the Wolf seer: WolfSeer("An").Checks("Em"). He bites with the
// pack through Wolf("An").Bites(...).
func (n *Night) WolfSeer(name string) *WolfSeerActor {
	return &WolfSeerActor{n: n, id: n.Player(name)}
}

// WolfSeerActor is the Wolf seer acting.
type WolfSeerActor struct {
	n  *Night
	id engine.PlayerID
}

// Checks looks at one player; read the answer with ExpectWolfSeerSaw.
func (a *WolfSeerActor) Checks(target string) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.WolfSeer)
	return a.n.Do(roles.WolfSeerCheck{By: a.id, Target: a.n.Player(target)})
}

// CannotCheck looks and expects the engine to refuse with want.
func (a *WolfSeerActor) CannotCheck(target string, want error) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.WolfSeer)
	return a.n.DoExpectErr(roles.WolfSeerCheck{By: a.id, Target: a.n.Player(target)}, want)
}

// ExpectWolfSeerSaw checks the pack's answer for a player: the Seer or not.
func (n *Night) ExpectWolfSeerSaw(target string, seer bool) *Night {
	n.g.t.Helper()
	for _, e := range n.g.events {
		r, ok := e.(roles.WolfSeerChecked)
		if ok && r.Target == n.Player(target) {
			if r.IsSeer != seer {
				n.g.t.Fatalf("expected the Wolf seer to see %s as Seer=%v, but saw Seer=%v", target, seer, r.IsSeer)
			}
			return n
		}
	}
	n.g.t.Fatalf("expected the Wolf seer to have checked %s, but no such check happened", target)
	return n
}

// WolfSeerResults checks how many Wolf seer results the viewer knows of.
func (s *Seen) WolfSeerResults(n int) *Seen {
	s.g.t.Helper()
	if got := count[roles.WolfSeerChecked](s.feed); got != n {
		s.g.t.Fatalf("%s should know of %d Wolf seer results, knows of %d", s.who, n, got)
	}
	return s
}
