package enginetest

import (
	"sort"

	"masoi/internal/engine"
)

// ExpectAlive checks that every named player is alive.
func (g *Game) ExpectAlive(names ...string) *Game {
	g.t.Helper()
	for _, name := range names {
		if p := g.state.Players[g.Player(name)]; !p.Alive {
			g.t.Fatalf("expected %s alive, but %s is dead", name, name)
		}
	}
	return g
}

// ExpectDead checks that every named player is dead (now, whenever they died).
func (g *Game) ExpectDead(names ...string) *Game {
	g.t.Helper()
	for _, name := range names {
		if p := g.state.Players[g.Player(name)]; p.Alive {
			g.t.Fatalf("expected %s dead, but %s is alive", name, name)
		}
	}
	return g
}

// ExpectRole checks who holds which role.
func (g *Game) ExpectRole(name string, role engine.RoleID) *Game {
	g.t.Helper()
	if got := g.state.Players[g.Player(name)].Role; got != role {
		g.t.Fatalf("expected %s to be %s, but %s is %s", name, role, name, got)
	}
	return g
}

// ExpectPhase checks the engine phase.
func (g *Game) ExpectPhase(phase engine.Phase) *Game {
	g.t.Helper()
	if g.state.Phase != phase {
		g.t.Fatalf("expected phase %s, but the game is in %s", phase, g.state.Phase)
	}
	return g
}

// sameNames compares two name lists ignoring order.
func sameNames(a, b []string) bool {
	a, b = append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
