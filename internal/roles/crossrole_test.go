package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// Scenarios where several roles act on the same night (task T20).

// visitorKnightTable: ten players in a circle An, Binh, Chi, Dung, Em, Phuc,
// Giang, Hoa, Khoa, Lan. Wolves: An, Chi, Giang. Binh is the Rusty knight,
// Dung the Guard, Em the Seer, Phuc the Visitor, Hoa the Witch, Lan the Chief.
func visitorKnightTable(t testing.TB) *Game {
	return NewGame(t, Seats(all...), Chief("Lan"),
		Cast("An", Werewolf, "Binh", RustyKnight, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Visitor, "Giang", Werewolf, "Hoa", Witch, "Khoa", Villager, "Lan", Villager))
}

// The Visitor sleeps with the Witch on the night the pack bites the Knight: the
// Witch sees nothing and cannot heal, the Knight dies and the nearest wolf on
// his left (Chi) is poisoned; it dies when the pack step of the next night
// starts and is announced with the others.
func TestCross_VisitorBlocksTheWitch_TheNightTheKnightIsBitten(t *testing.T) {
	g := visitorKnightTable(t)
	n := g.Night().Visitor("Phuc").Visits("Hoa")
	bite(n, "Binh", "An", "Chi", "Giang").TurnOf(Witch)
	g.As("Hoa").Blocked(true).Bitten("").BiteShown(false)
	n.Witch("Hoa").CannotHeal(engine.ErrBlocked)
	g.Dawn().ExpectDead("Binh")
	g.Day().NobodyHanged()

	n = g.Night().Visitor("Phuc").Visits("Khoa").TurnOf(Werewolf)
	g.As("Chi").Doomed(true)
	g.As("An").PackDoomed("Chi")
	n.Wolf("Chi").CannotBite("Lan", engine.ErrNotYourTurn)
	bite(n, "Lan", "An", "Giang")
	g.Dawn().ExpectDead("Chi", "Lan")
	wantDead(t, announcedDead(t, g), "Chi", "Lan")
}

// The Wolf seer finds the Seer and the pack bites her, but the Guard shields
// her and the Witch heals her too: nobody dies, the morning says "được bảo vệ",
// and the heal potion is gone.
func TestCross_WolfSeerFindsTheSeer_PackBitesHer_GuardShieldsAndWitchHeals(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night().Guard("Dung").Protects("Em")
	n.WolfSeer("An").Checks("Em").ExpectWolfSeerSaw("Em", true)
	n.Wolf("An").Bites("Em").Wolf("Binh").Bites("Em").TurnOf(Witch)
	g.As("Giang").Bitten("Em") // she cannot know the Guard shielded her
	n.Witch("Giang").Heals()
	n.Seer("Em").Checks("An").ExpectSeerSaw("Em", "An", true)
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
	g.Day().NobodyHanged()
	n = g.Night().Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	n.Witch("Giang").CannotHeal(engine.ErrPotionUsed)
	g.Dawn().ExpectDead("Hoa")
}

// The last two wolves: one grieves (his lover was hanged by day) and is dead
// from the pack step, the Visitor sleeps with the other. Nobody can vote, so
// nobody is bitten; only the grieving wolf dies.
func TestCross_LastTwoWolves_OneGrieves_TheOtherIsVisited_NobodyIsBitten(t *testing.T) {
	g := cupidTableWith(t, map[string]engine.RoleID{"Khoa": Visitor})
	g.Night().Cupid("Lan").Pairs("Binh", "Chi").Visitor("Khoa").Visits("Hoa")
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("Chi")
	g.ExpectAlive("Binh")

	n := g.Night().Visitor("Khoa").Visits("An").TurnOf(Werewolf)
	g.As("Binh").Doomed(true).Grief(true)
	g.As("An").Blocked(true)
	n.Wolf("An").CannotBite("Em", engine.ErrBlocked)
	n.Wolf("Binh").CannotBite("Em", engine.ErrNotYourTurn)
	n.TurnOf(Witch)
	g.As("Giang").BiteShown(true).Bitten("")
	g.Dawn().ExpectDead("Binh").ExpectAnnouncement(i18n.DawnSomeoneDied)
	wantDead(t, announcedDead(t, g), "Binh")
}
