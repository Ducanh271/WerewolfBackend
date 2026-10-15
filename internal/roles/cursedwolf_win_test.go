package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/i18n"
)

// killOff leaves only the named players alive.
func killOff(g *Game, survivors ...string) {
	keep := map[engine.PlayerID]bool{}
	for _, name := range survivors {
		keep[g.Player(name)] = true
	}
	g.HandBuild(func(s *engine.State) {
		for id, p := range s.Players {
			p.Alive = keep[id]
		}
		s.Chief = "" // the title is not what these scenarios are about
	})
}

// Rules §8: the win is checked and announced by day only, on the situation of
// that moment. A curse that puts the wolves level with the village ends nothing
// during the night.
func TestWin_CurseMakesWolvesEqual_NightRunsOn_WolvesWinWhenTheMorningIsAnnounced(t *testing.T) {
	g := cursedTable(t)
	n := g.Night()
	killOff(g, "An", "Binh", "Chi", "Hoa", "Lan") // 2 wolves, 3 villagers
	n.CursedWolf("An").Curses("Chi")              // 3 wolves, 2 villagers
	g.ExpectPhase(engine.PhaseNight)
	n.PhoneCheck()
	g.ExpectPhase(engine.PhaseNight)
	n.Wolf("Chi").Bites("Lan") // the new wolf still bites; the pack is An, Binh, Chi
	n.Wolf("Binh").Bites("Lan")
	g.ExpectPhase(engine.PhaseNight)

	d := g.Dawn().ExpectDead("Lan").ExpectAnnouncement(i18n.WinWolves)
	_ = d
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamWolves {
		t.Fatalf("winner %v, want the wolves", w)
	}
}

func TestWin_CurseMakesWolvesEqual_NobodyDiesInTheNight_WolvesStillWinInTheMorning(t *testing.T) {
	g := cursedTable(t)
	n := g.Night()
	killOff(g, "An", "Binh", "Chi", "Hoa", "Lan")
	n.CursedWolf("An").Curses("Chi")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.WinWolves)
	g.ExpectPhase(engine.PhaseEnded)
}

func TestWin_CurseMakesWolvesEqual_ButWitchPoisonsAWolfThatNight_GameGoesOn(t *testing.T) {
	g := cursedTable(t)
	n := g.Night()
	killOff(g, "An", "Binh", "Giang", "Chi", "Hoa", "Lan") // 2 wolves, 4 villagers
	n.CursedWolf("An").Curses("Chi")                       // 3 wolves, 3 villagers: level
	n.Witch("Giang").Poisons("Binh")                       // 2 wolves, 3 villagers by the morning
	g.Dawn().ExpectDead("Binh")
	if phase := g.State().Phase; phase == engine.PhaseEnded {
		t.Fatal("the game ended though the wolves are fewer than the village when the morning is announced")
	}
	g.ExpectAlive("An", "Chi", "Giang", "Hoa", "Lan")
	g.Day().NobodyHanged()
	g.Night() // and play goes on
}
