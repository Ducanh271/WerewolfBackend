package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/roles"
)

// toStep times out the night's steps until step k (1 = Guard, 2 = wolf pack,
// 3 = Seer) is being called.
func toStep(g *Game, k int) {
	for ; k > 1; k-- {
		s := g.State()
		g.Do(engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	}
}

func TestNight_TargetMustBeSeatedAndAlive(t *testing.T) {
	tests := []struct {
		name string
		step int
		cmd  func(target engine.PlayerID) engine.Command
	}{
		{"guard", 1, func(to engine.PlayerID) engine.Command { return roles.GuardProtect{By: "Dung", Target: to} }},
		{"wolf", 2, func(to engine.PlayerID) engine.Command { return roles.WolfVote{By: "An", Target: to} }},
		{"seer", 3, func(to engine.PlayerID) engine.Command { return roles.SeerCheck{By: "Em", Target: to} }},
	}
	for _, tc := range tests {
		for _, target := range []engine.PlayerID{"Zed", "Phuc"} { // nobody, and dead
			t.Run(tc.name+"/"+string(target), func(t *testing.T) {
				g := newTable(t)
				g.State().Players["Phuc"].Alive = false
				g.Night()
				toStep(g, tc.step)
				g.DoExpectErr(tc.cmd(target), engine.ErrInvalidTarget)
			})
		}
	}
}

func TestNight_DeadPlayerCannotAct(t *testing.T) {
	g := newTable(t)
	g.State().Players["Dung"].Alive = false
	g.Night()
	g.DoExpectErr(roles.GuardProtect{By: "Dung", Target: "Phuc"}, engine.ErrNotAlive)
}
