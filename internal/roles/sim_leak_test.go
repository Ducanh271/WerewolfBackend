package roles_test

import (
	"fmt"

	"masoi/internal/engine"
	"masoi/internal/engine/enginetest"
	"masoi/internal/roles"
)

// checkViews is the secrecy check after every command: no living player's
// screen carries another player's role or team (lovers and a revealed Hunter
// excepted), the full state, or the pack unless they are a wolf.
func (g *simGame) checkViews() {
	if g.s.Phase == engine.PhaseEnded || !g.s.Dealt {
		return // after the end everybody sees everything (rules §4)
	}
	for _, id := range g.living() {
		v := engine.ViewFor(g.s, g.log, engine.PlayerViewer(id))
		me := g.s.Players[id]
		if v.Full || v.Secrets != nil {
			g.leak(id, "a living player got a full view")
		}
		if (v.Pack != nil) != (engine.VisibleTeam(me) == engine.TeamWolves) {
			g.leak(id, "pack view does not match the wolf team")
		}
		if v.Pack != nil {
			for _, m := range v.Pack.Members {
				if g.s.Players[m].Team != engine.TeamWolves && engine.VisibleTeam(g.s.Players[m]) == engine.TeamWolves {
					g.leak(id, "pack lists a non-wolf")
				}
			}
		}
		for _, pv := range v.Players {
			if pv.ID == id {
				continue
			}
			lover := v.Lover != nil && v.Lover.Partner == pv.ID
			revealed := g.s.Players[pv.ID].Flags[engine.FlagRevealed] != 0
			if pv.Team != engine.TeamNone && !lover || pv.Role != "" && !lover && !revealed {
				g.leak(id, fmt.Sprintf("sees %s as role %q team %q", pv.ID, pv.Role, pv.Team))
			}
		}
		if v.Me != nil && v.Me.Turn != nil && v.Me.Turn.BiteShown {
			if _, witch := engine.Registry[me.Role].(engine.BiteWitness); !witch {
				g.leak(id, "sees the bitten player without being the Witch")
			}
		}
	}
}

// checkEvents is the secrecy check on the event stream each living player may
// know of.
func (g *simGame) checkEvents() {
	for _, id := range g.living() {
		g.checkEventsOf(id)
	}
}

// checkEventsOf checks what one living player may know: no private result of
// somebody else, and no list of players in an order that tells more than the
// announcements did (see playerListLeak).
func (g *simGame) checkEventsOf(id engine.PlayerID) {
	if leak := playerListLeak(g.s, g.log, id); leak != "" {
		g.leak(id, leak)
	}
	me := g.s.Players[id]
	for _, e := range g.views.Visible(g.s, g.log, engine.PlayerViewer(id)) {
		switch ev := e.(type) {
		case engine.RolesDealt, engine.PlayerDied, engine.HunterChainNoted:
			g.leak(id, fmt.Sprintf("receives %T", e))
		case roles.SeerChecked:
			g.mustBe(id, ev.Seer == id, e)
		case roles.PlayerProtected:
			g.mustBe(id, ev.Guard == id, e)
		case roles.PlayerCursed:
			g.mustBe(id, ev.Wolf == id, e)
		case roles.WitchHealed:
			g.mustBe(id, ev.Witch == id, e)
		case roles.WitchPoisoned:
			g.mustBe(id, ev.Witch == id, e)
		case roles.WolfSeerChecked:
			g.mustBe(id, engine.VisibleTeam(me) == engine.TeamWolves, e)
		case roles.WolfVoted:
			g.mustBe(id, engine.VisibleTeam(me) == engine.TeamWolves, e)
		}
	}
}

func (g *simGame) mustBe(viewer engine.PlayerID, ok bool, e engine.Event) {
	if !ok {
		g.leak(viewer, fmt.Sprintf("receives somebody else's private result %T %+v", e, e))
	}
}

func (g *simGame) leak(viewer engine.PlayerID, what string) {
	panic(fmt.Sprintf("secret leak to %s (%s, night %d, phase %v): %s", viewer, g.s.Players[viewer].Role, g.s.Night, g.s.Phase, what))
}

// checkViewCache is the differential test of the fast view path (audit R5):
// every phone's view and events from the cache equal the full replay.
func (g *simGame) checkViewCache() {
	for _, v := range enginetest.Viewers(g.log) {
		if diff := enginetest.ViewCacheDiff(&g.views, g.s, g.log, v); diff != "" {
			panic(fmt.Sprintf("view cache differs from the full replay (%s): %s", g.desc, diff))
		}
	}
}
