package roles_test

import (
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

// detectiveKnightTable: ten players in a circle An, Binh, Chi, Dung, Em, Phuc,
// Giang, Hoa, Khoa, Lan (Lan sits next to An). An is the Cursed wolf, Dung a
// wolf, Chi the Rusty knight, Phuc the Detective, Lan the Chief; the others are
// villagers.
func detectiveKnightTable(t testing.TB) *Game {
	return NewGame(t, Seats(all...), Chief("Lan"),
		Cast("An", CursedWolf, "Binh", Villager, "Chi", RustyKnight, "Dung", Werewolf, "Em", Villager,
			"Phuc", Detective, "Giang", Villager, "Hoa", Villager, "Khoa", Villager, "Lan", Villager))
}

// Rules §6 "Thám tử": on night 2 Dung, poisoned by the Knight, is dead by the
// rules from the pack step on, but the table still shows him alive. For the
// Detective he counts as alive in everything: as a neighbour and as a wolf, so
// the result fits the table he sees. Giang, cursed on night 1, counts as a wolf.
func TestCross_DetectiveZone_WithADoomedWolfAndACursedPlayer(t *testing.T) {
	tests := []struct {
		name   string
		target string
		wolf   bool
	}{
		{"the doomed wolf is the nearest neighbour", "Em", true}, // Dung, Em, Phuc (skipping Dung: Binh, Em, Phuc)
		{"the doomed wolf himself", "Dung", true},                // Binh, Dung, Em
		{"the cursed player is a neighbour", "Hoa", true},        // Giang, Hoa, Khoa
		{"no wolf in the zone", "Khoa", false},                   // Hoa, Khoa, Lan
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := detectiveKnightTable(t)
			n := g.Night().CursedWolf("An").Curses("Giang").PhoneCheck()
			bite(n, "Chi", "An", "Dung", "Giang") // the Knight poisons Dung, next on his left
			g.Dawn().ExpectDead("Chi")
			g.Day().NobodyHanged()

			n = g.Night().TurnOf(Detective)
			if !engine.Doomed(g.State(), g.Player("Dung")) {
				t.Fatal("Dung should be dead by the rules from the pack step on")
			}
			for _, p := range g.As("Phuc").Dead("Chi").View().Players {
				if p.ID == g.Player("Dung") && !p.Alive {
					t.Fatal("the Detective must still see Dung alive")
				}
			}
			n.Detective("Phuc").Checks(tc.target).ExpectDetectiveSaw(tc.target, tc.wolf)
			g.Dawn().ExpectDead("Dung")
		})
	}
}
