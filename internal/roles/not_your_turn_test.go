package roles_test

import (
	"errors"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"

	. "masoi/internal/engine/enginetest"
)

// A player who does not hold a role may send that role's command (a modified
// app, devtools). The answer must never depend on their own state: a visited
// villager who got "blocked" while the others get "not your turn" would learn
// that the Visitor came to them (rules §3, §6 "Nguyệt Nữ"). So whoever does
// not hold the role gets exactly engine.ErrNotYourTurn, the same code and the
// same text, whether blocked, dead by the rules, dead, or simply someone else.

// nightCommands is every role command, with the test of who holds it: the
// role's holder, or anybody on the wolf team for the pack's vote.
var nightCommands = []struct {
	name  string
	holds func(p *engine.Player) bool
	build func(by, target engine.PlayerID) engine.Command
}{
	{"cupidPair", hasRole(Cupid), func(by, t engine.PlayerID) engine.Command { return roles.CupidPair{By: by, A: t, B: by} }},
	{"visitorVisit", hasRole(Visitor), func(by, t engine.PlayerID) engine.Command { return roles.VisitorVisit{By: by, Target: t} }},
	{"guardProtect", hasRole(Guard), func(by, t engine.PlayerID) engine.Command { return roles.GuardProtect{By: by, Target: t} }},
	{"cursedWolfCurse", hasRole(CursedWolf), func(by, t engine.PlayerID) engine.Command { return roles.CursedWolfCurse{By: by, Target: t} }},
	{"wolfSeerCheck", hasRole(WolfSeer), func(by, t engine.PlayerID) engine.Command { return roles.WolfSeerCheck{By: by, Target: t} }},
	{"wolfVote", func(p *engine.Player) bool { return p.Team == engine.TeamWolves },
		func(by, t engine.PlayerID) engine.Command { return roles.WolfVote{By: by, Target: t} }},
	{"witchHeal", hasRole(Witch), func(by, _ engine.PlayerID) engine.Command { return roles.WitchHeal{By: by} }},
	{"witchPoison", hasRole(Witch), func(by, t engine.PlayerID) engine.Command { return roles.WitchPoison{By: by, Target: t} }},
	{"seerCheck", hasRole(Seer), func(by, t engine.PlayerID) engine.Command { return roles.SeerCheck{By: by, Target: t} }},
	{"detectiveCheck", hasRole(Detective), func(by, t engine.PlayerID) engine.Command { return roles.DetectiveCheck{By: by, Target: t} }},
}

func hasRole(role engine.RoleID) func(p *engine.Player) bool {
	return func(p *engine.Player) bool { return p.Role == role }
}

// twelveTable has every role with a night step, the Hunter and the Rusty
// knight. Lan the Knight sits right of Minh, Nga and An (left = next seat).
func twelveTable(t *testing.T) *Game {
	return NewGame(t,
		Seats("An", "Binh", "Chi", "Dung", "Em", "Phuc", "Giang", "Hoa", "Khoa", "Lan", "Minh", "Nga"),
		Cast("An", CursedWolf, "Binh", WolfSeer, "Chi", Cupid, "Dung", Visitor, "Em", Guard, "Phuc", Witch,
			"Giang", Seer, "Hoa", Detective, "Khoa", Hunter, "Lan", RustyKnight, "Minh", Villager, "Nga", Villager),
		Chief("Giang"))
}

// probe is one night step: every player who does not hold a command's role
// sends it, and each must get the bare ErrNotYourTurn. It records the states
// the senders were in, so the test can prove it met all of them.
func probe(t *testing.T, s engine.State, seen map[string]bool) {
	t.Helper()
	step, _ := engine.CurrentNightRole(s)
	for _, c := range nightCommands {
		for _, id := range s.Seats {
			p := s.Players[id]
			if c.holds(p) {
				continue
			}
			_, err := engine.Decide(s, c.build(id, "Giang"))
			if !errors.Is(err, engine.ErrNotYourTurn) || err.Error() != engine.ErrNotYourTurn.Error() {
				t.Fatalf("step %s: %s (%s) sent %s and got %q, want exactly %q",
					step, id, p.Role, c.name, err, engine.ErrNotYourTurn)
			}
			switch {
			case !p.Alive:
				seen["dead"] = true
			case engine.Doomed(s, id):
				seen["doomed"] = true
			case engine.Blocked(s, id):
				seen["blocked"] = true
			default:
				seen["wrong role"] = true
			}
		}
	}
}

func TestNight_PlayerWithoutTheRole_AlwaysGetsTheSameNotYourTurn_WhateverTheirState(t *testing.T) {
	g := twelveTable(t)
	seen := map[string]bool{}

	// Night 1: Cupid pairs the Guard with Minh; the pack bites the Knight, so
	// An, the nearest wolf on his left, is poisoned for night 2.
	n := g.Night()
	probe(t, g.State(), seen) // Cupid's step
	n.Cupid("Chi").Pairs("Em", "Minh").Visitor("Dung").Visits("Khoa").Guard("Em").Protects("Giang").
		Wolf("An").Bites("Lan").Wolf("Binh").Bites("Lan")
	g.Dawn().ExpectDead("Lan")

	// Day 1: Minh is hanged, so Em the Guard grieves and dies at her own step.
	g.Day().Hanged("Minh")

	// Night 2: dead (Lan, Minh), visited (Nga), dead by the rules (Em from
	// the Guard's step, An from the pack's step) and everybody else.
	g.Night().Visitor("Dung").Visits("Nga")
	for {
		s := g.State()
		if _, ok := engine.CurrentNightRole(s); !ok {
			break
		}
		probe(t, s, seen)
		g.Do(engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	}
	for _, state := range []string{"dead", "doomed", "blocked", "wrong role"} {
		if !seen[state] {
			t.Errorf("no sender was %s: the test did not cover that state", state)
		}
	}
}

// The holder still gets the detail: they see all of it on their own screen.
func TestNight_HolderStillGetsTheReason(t *testing.T) {
	g := twelveTable(t)
	g.Night().Visitor("Dung").Visits("Giang").TurnOf(Seer)
	_, err := engine.Decide(g.State(), roles.SeerCheck{By: g.Player("Giang"), Target: g.Player("An")})
	if !errors.Is(err, engine.ErrBlocked) {
		t.Fatalf("the visited Seer: err = %v, want ErrBlocked", err)
	}
}
