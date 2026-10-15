package enginetest

import "masoi/internal/engine"

// Doomed checks the notice on the viewer's pack turn: the wolf was poisoned
// by the Rusty knight, is dead from now on and cannot vote.
func (s *Seen) Doomed(want bool) *Seen {
	s.g.t.Helper()
	got := s.view.Me != nil && s.view.Me.Turn != nil && s.view.Me.Turn.Doomed
	if got != want {
		s.g.t.Fatalf("%s: doomed notice shown=%v, want %v", s.who, got, want)
	}
	return s
}

// PackDoomed checks which wolves the viewer sees as unable to vote because of
// the Rusty knight's poison (none when called with no names).
func (s *Seen) PackDoomed(names ...string) *Seen {
	s.g.t.Helper()
	var got []string
	if s.view.Pack != nil {
		for _, id := range s.view.Pack.Doomed {
			got = append(got, string(id))
		}
	}
	if !sameNames(got, names) {
		s.g.t.Fatalf("%s should see these wolves unable to vote: %v, sees %v", s.who, names, got)
	}
	return s
}

// ExpectDawnNames checks the names the narrator reads at the last dawn, in
// the order they are read.
func (g *Game) ExpectDawnNames(names ...string) *Game {
	g.t.Helper()
	var read []engine.PlayerID
	for _, e := range g.events {
		if d, ok := e.(engine.DawnAnnounced); ok {
			read = d.Dead
		}
	}
	got := make([]string, len(read))
	for i, id := range read {
		got[i] = string(id)
	}
	if len(got) != len(names) {
		g.t.Fatalf("the dawn should read the names %v, read %v", names, got)
	}
	for i := range got {
		if got[i] != names[i] {
			g.t.Fatalf("the dawn should read the names %v, read %v", names, got)
		}
	}
	return g
}
