package roles_test

import (
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

func TestView_RustyKnight_PoisonedWolfAndPackSeeItOnThePackStepOnly(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn()
	g.As("Chi").Doomed(false).PackDoomed() // still alive and unaware
	g.Day().NobodyHanged()

	n := g.Night().TurnOf(Guard)
	g.As("Chi").Doomed(false).PackDoomed() // the pack step has not started

	n.TurnOf(Werewolf)
	g.As("Chi").Turn(Werewolf).Doomed(true).PackDoomed("Chi")
	g.As("An").Turn(Werewolf).Doomed(false).PackDoomed("Chi")
	g.As("Giang").Doomed(false).PackDoomed("Chi")
}

func TestView_RustyKnight_NobodyOutsideThePackLearnsOfItDuringTheNight(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn()
	g.Day().NobodyHanged()
	n := g.Night().TurnOf(Werewolf)
	n.TurnOf(Seer)

	for _, who := range []string{"Em", "Hoa", "Phuc", "Dung"} {
		g.As(who).PublicTable().NoPack().Doomed(false).NoSecretsOfOthers()
		for _, p := range g.As(who).View().Players {
			if p.ID == g.Player("Chi") && !p.Alive {
				t.Fatalf("%s sees Chi dead during the night", who)
			}
		}
		g.As(who).NoEvent(engine.WolfPoisoned{Wolf: g.Player("Chi"), Night: 1})
	}
}

func TestView_RustyKnight_DeadViewersSeeTheWholeStory(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn()
	g.As("Binh").Full().Event(engine.WolfPoisoned{Wolf: g.Player("Chi"), Night: 1})
}
