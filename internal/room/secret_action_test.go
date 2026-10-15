package room

import (
	"fmt"
	"sort"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// Rules §3: a real night call and a fake one must look the same on every
// phone. If a phone got a message whenever the role being called acts, the
// table would know that role is alive. So a secret night action sends nothing
// at all to any other phone, except to those the rules let see its result
// (the pack sees the Wolf seer's answer and each other's bite votes).

// dealTwelve seats twelve players and deals a table where every role with a
// night step plays, plus the Hunter and the Rusty knight.
func (g *game) dealTwelve() {
	g.t.Helper()
	host := g.cl["p1"]
	var order []engine.PlayerID
	for i := 1; i <= 12; i++ {
		order = append(order, engine.PlayerID(fmt.Sprintf("p%d", i)))
	}
	cast := map[engine.RoleID]int{
		roles.CursedWolf: 1, roles.WolfSeer: 1, roles.Cupid: 1, roles.Visitor: 1, roles.Guard: 1, roles.Witch: 1,
		roles.Seer: 1, roles.Detective: 1, roles.Hunter: 1, roles.RustyKnight: 1, roles.Villager: 2,
	}
	g.must(host.Submit(as(engine.SetSeats{Order: order})))
	g.must(host.Submit(as(engine.ConfigureRoles{Roles: cast})))
	g.must(host.Submit(as(engine.DealRoles{})))
}

// holders lists the players of a role in seat order.
func (g *game) holders(role engine.RoleID) []engine.PlayerID {
	s := g.snap()
	var out []engine.PlayerID
	for _, id := range s.Seats {
		if s.Players[id].Role == role {
			out = append(out, id)
		}
	}
	return out
}

func (g *game) wolves() map[engine.PlayerID]bool {
	out := map[engine.PlayerID]bool{}
	for id, p := range g.snap().Players {
		if p.Team == engine.TeamWolves {
			out[id] = true
		}
	}
	return out
}

func (g *game) received() map[engine.PlayerID]int {
	out := map[engine.PlayerID]int{}
	for id, s := range g.sink {
		s.mu.Lock()
		out[id] = len(s.msgs)
		s.mu.Unlock()
	}
	return out
}

type secretAction struct {
	by  engine.PlayerID
	cmd engine.Command
}

// expectSilent sends the actions and checks who got a message: the actor
// must (their own result), anybody in mayHear may, nobody else may.
func (g *game) expectSilent(step engine.RoleID, mayHear map[engine.PlayerID]bool, actions ...secretAction) {
	g.t.Helper()
	before := g.received()
	for _, a := range actions {
		g.must(g.cl[a.by].Submit(as(a.cmd)))
	}
	after := g.received()
	actors := map[engine.PlayerID]bool{}
	for _, a := range actions {
		actors[a.by] = true
		if after[a.by] == before[a.by] {
			g.t.Fatalf("%s: the actor %s got nothing back: the check is not looking at live phones", step, a.by)
		}
	}
	var told []string
	for id := range g.sink {
		if after[id] != before[id] && !actors[id] && !mayHear[id] {
			told = append(told, fmt.Sprintf("%s (%s)", id, g.snap().Players[id].Role))
		}
	}
	sort.Strings(told)
	if len(told) > 0 {
		g.t.Fatalf("%s: a secret action sent messages to %v", step, told)
	}
}

func TestRoom_SecretNightAction_SendsNothingToAnyOtherPhone(t *testing.T) {
	g := newGame(t, 12)
	g.dealTwelve()
	g.toNight1() // night 1: nobody is dead, so "any other phone" is every other phone

	one := func(role engine.RoleID) engine.PlayerID { return g.holders(role)[0] }
	villagers := g.holders(roles.Villager)
	none := map[engine.PlayerID]bool{}
	acted := map[engine.RoleID]bool{}

	for {
		s := g.snap()
		step, ok := engine.CurrentNightRole(s)
		if !ok {
			break
		}
		switch step {
		case roles.Cupid:
			g.expectSilent(step, none, secretAction{one(roles.Cupid), roles.CupidPair{By: one(roles.Cupid), A: villagers[0], B: villagers[1]}})
		case roles.Visitor:
			g.expectSilent(step, none, secretAction{one(roles.Visitor), roles.VisitorVisit{By: one(roles.Visitor), Target: one(roles.Hunter)}})
		case roles.Guard:
			g.expectSilent(step, none, secretAction{one(roles.Guard), roles.GuardProtect{By: one(roles.Guard), Target: one(roles.Seer)}})
		case roles.CursedWolf: // a lover is cursed: nothing shows before the phone check, not even to the couple
			g.expectSilent(step, none, secretAction{one(roles.CursedWolf), roles.CursedWolfCurse{By: one(roles.CursedWolf), Target: villagers[1]}})
		case roles.WolfSeer:
			g.expectSilent(step, g.wolves(), secretAction{one(roles.WolfSeer), roles.WolfSeerCheck{By: one(roles.WolfSeer), Target: one(roles.Seer)}})
		case roles.Werewolf:
			var votes []secretAction
			for id := range g.wolves() {
				votes = append(votes, secretAction{id, roles.WolfVote{By: id, Target: one(roles.Detective)}})
			}
			g.expectSilent(step, g.wolves(), votes...)
		case roles.Witch:
			g.expectSilent(step, none,
				secretAction{one(roles.Witch), roles.WitchHeal{By: one(roles.Witch)}},
				secretAction{one(roles.Witch), roles.WitchPoison{By: one(roles.Witch), Target: villagers[0]}})
		case roles.Seer:
			g.expectSilent(step, none, secretAction{one(roles.Seer), roles.SeerCheck{By: one(roles.Seer), Target: one(roles.CursedWolf)}})
		case roles.Detective:
			g.expectSilent(step, none, secretAction{one(roles.Detective), roles.DetectiveCheck{By: one(roles.Detective), Target: one(roles.Hunter)}})
		}
		acted[step] = true

		// The step running out is public: every phone hears of it.
		before := g.received()
		g.step(engine.NightStepDuration(s, step))
		after := g.received()
		for id := range g.sink {
			if after[id] == before[id] {
				t.Fatalf("the end of step %s did not reach %s", step, id)
			}
		}
	}
	for _, role := range []engine.RoleID{roles.Cupid, roles.Visitor, roles.Guard, roles.CursedWolf, roles.WolfSeer,
		roles.Werewolf, roles.Witch, roles.Seer, roles.Detective} {
		if !acted[role] {
			t.Errorf("step %s was never played", role)
		}
	}
}
