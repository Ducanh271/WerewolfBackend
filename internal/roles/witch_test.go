package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	"masoi/internal/roles"

	. "masoi/internal/engine/enginetest"
)

// packBites has two of the three wolves bite the player.
func packBites(n *Night, victim string) *Night {
	return n.Wolf("An").Bites(victim).Wolf("Binh").Bites(victim)
}

func TestWitch_OneHealAndOnePoisonPerGame_BothUsableTheSameNight(t *testing.T) {
	g := witchTable(t)
	n := g.Night()
	packBites(n, "Phuc").Witch("Lan").Heals().Witch("Lan").Poisons("Hoa")
	g.Dawn().ExpectDead("Hoa")
	g.Day().NobodyHanged()

	n = g.Night()
	packBites(n, "Giang").
		Witch("Lan").CannotHeal(engine.ErrPotionUsed).
		Witch("Lan").CannotPoison("Khoa", engine.ErrPotionUsed)
	g.Dawn().ExpectDead("Giang")
}

func TestWitch_SeesTheBiteOnlyWhileSheHasTheHealPotion(t *testing.T) {
	g := witchTable(t)
	packBites(g.Night(), "Phuc").TurnOf(Witch)
	g.As("Lan").Bitten("Phuc").BiteShown(true)
	g.Do(engine.PhaseTimeout{Phase: g.State().Phase, Seq: g.State().Seq}) // she says nothing
	g.As("Lan").Bitten("").BiteShown(false)
	g.Dawn().ExpectDead("Phuc")
	g.Day().NobodyHanged()

	// Heal used: the next night she is told nothing.
	g2 := witchTable(t)
	packBites(g2.Night(), "Phuc").Witch("Lan").Heals()
	g2.Dawn().ExpectNobodyDied()
	g2.Day().NobodyHanged()
	packBites(g2.Night(), "Giang").TurnOf(Witch)
	g2.As("Lan").Bitten("").BiteShown(false)
}

func TestWitch_CanHealHerself(t *testing.T) {
	g := witchTable(t)
	packBites(g.Night(), "Lan").Witch("Lan").Heals()
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
}

func TestNight_GuardAndWitchSaveSameTarget_AnnouncesProtected(t *testing.T) {
	g := witchTable(t)
	n := g.Night().Guard("Dung").Protects("Phuc")
	packBites(n, "Phuc").Witch("Lan").Heals()
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
	g.Day().NobodyHanged()

	// The heal potion is gone all the same.
	packBites(g.Night(), "Giang").Witch("Lan").CannotHeal(engine.ErrPotionUsed)
}

func TestWitch_GuardProtectsButWitchDoesNotKnow_SheSeesTheBite(t *testing.T) {
	g := witchTable(t)
	packBites(g.Night().Guard("Dung").Protects("Phuc"), "Phuc").TurnOf(Witch)
	g.As("Lan").Bitten("Phuc").BiteShown(true)
}

func TestWitch_NobodyBitten_SheSeesNobodyAndKeepsThePotion(t *testing.T) {
	g := witchTable(t)
	n := g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Giang").Wolf("Chi").Bites("Hoa")
	n.TurnOf(Witch)
	g.As("Lan").Bitten("").BiteShown(true)
	n.Witch("Lan").CannotHeal(engine.ErrNothingToHeal)
	g.Dawn().ExpectNobodyDied()
	if g.State().Players[g.Player("Lan")].Flags["witch.heal_used"] != 0 {
		t.Fatal("a refused heal must not use up the potion")
	}
}

func TestWitch_PoisonKillsAnyoneEvenHerselfAndTheGuardCannotBlockIt(t *testing.T) {
	tests := []struct {
		name   string
		guards string
		target string
	}{
		{"a villager", "", "Phuc"},
		{"a wolf", "", "Chi"},
		{"herself", "", "Lan"},
		{"a player the Guard shields", "Phuc", "Phuc"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := witchTable(t)
			n := g.Night()
			if tc.guards != "" {
				n.Guard("Dung").Protects(tc.guards)
			}
			n.Witch("Lan").Poisons(tc.target)
			g.Dawn().ExpectDead(tc.target).ExpectAnnouncement(i18n.DawnSomeoneDied)
		})
	}
}

func TestWitch_VisitedByTheVisitor_SeesNothingAndCannotAct(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Giang")
	packBites(n, "Phuc").TurnOf(Witch)
	g.As("Giang").Blocked(true).Bitten("").BiteShown(false)
	n.Witch("Giang").CannotHeal(engine.ErrBlocked).Witch("Giang").CannotPoison("Hoa", engine.ErrBlocked)
	g.Dawn().ExpectDead("Phuc")
}

func TestWitch_BitesAndPoisonsTwoPlayers_AnnouncedInSeatingOrder(t *testing.T) {
	g := witchTable(t)
	packBites(g.Night(), "Hoa").Witch("Lan").Poisons("Em")
	g.Dawn().ExpectDead("Em", "Hoa").ExpectAnnouncement(i18n.DawnSomeoneDied)
	var dead []engine.PlayerID
	for _, e := range g.Events() {
		if d, ok := e.(engine.DawnAnnounced); ok {
			dead = d.Dead
		}
	}
	if want := []engine.PlayerID{g.Player("Em"), g.Player("Hoa")}; !reflect.DeepEqual(dead, want) {
		t.Fatalf("announced %v, want seating order %v", dead, want)
	}
}

func TestWitch_PoisonsTheBittenPlayer_DiesOnceInTheAnnouncement(t *testing.T) {
	g := witchTable(t)
	packBites(g.Night(), "Phuc").Witch("Lan").Poisons("Phuc")
	g.Dawn().ExpectDead("Phuc")
	for _, e := range g.Events() {
		if d, ok := e.(engine.DawnAnnounced); ok && len(d.Dead) != 1 {
			t.Fatalf("announced %v, want one name", d.Dead)
		}
	}
}

func TestWitch_OnlyTheWitchCanBrew(t *testing.T) {
	g := witchTable(t)
	n := g.Night()
	packBites(n, "Phuc").TurnOf(Witch)
	g.DoExpectErr(roles.WitchHeal{By: g.Player("Em")}, engine.ErrNotYourTurn)
}
