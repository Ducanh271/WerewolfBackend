package enginetest

import (
	"masoi/internal/engine"
	"masoi/internal/roles"
)

// Detective is the detective: Detective("An").Checks("Binh").
func (n *Night) Detective(name string) *DetectiveActor {
	return &DetectiveActor{n: n, id: n.Player(name)}
}

// DetectiveActor is the detective acting.
type DetectiveActor struct {
	n  *Night
	id engine.PlayerID
}

// Checks looks at one player and the neighbours; read the answer with
// ExpectDetectiveSaw.
func (a *DetectiveActor) Checks(target string) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Detective)
	return a.n.Do(roles.DetectiveCheck{By: a.id, Target: a.n.Player(target)})
}

// CannotCheck looks and expects the engine to refuse with want.
func (a *DetectiveActor) CannotCheck(target string, want error) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.Detective)
	return a.n.DoExpectErr(roles.DetectiveCheck{By: a.id, Target: a.n.Player(target)}, want)
}

// ExpectDetectiveSaw checks the answer for the zone around target: a wolf in
// it or not.
func (n *Night) ExpectDetectiveSaw(target string, wolf bool) *Night {
	n.g.t.Helper()
	for _, e := range n.g.events {
		r, ok := e.(roles.DetectiveChecked)
		if ok && r.Target == n.Player(target) {
			if r.HasWolf != wolf {
				n.g.t.Fatalf("expected the Detective to see wolf=%v around %s, but saw wolf=%v", wolf, target, r.HasWolf)
			}
			return n
		}
	}
	n.g.t.Fatalf("expected the Detective to have checked %s, but no such check happened", target)
	return n
}

// DetectiveResults checks how many Detective results the viewer knows of.
func (s *Seen) DetectiveResults(n int) *Seen {
	s.g.t.Helper()
	if got := count[roles.DetectiveChecked](s.feed); got != n {
		s.g.t.Fatalf("%s should know of %d Detective results, knows of %d", s.who, n, got)
	}
	return s
}
