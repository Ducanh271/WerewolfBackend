package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

// detectiveTable: ten players in a circle An, Binh, Chi, Dung, Em, Phuc,
// Giang, Hoa, Khoa, Lan (Lan sits next to An). Wolves: An and Giang. Em is the
// Detective, Hoa the Visitor, everybody else a villager.
func detectiveTable(t testing.TB) *Game {
	return NewGame(t, Seats(all...),
		Cast("An", Werewolf, "Giang", Werewolf, "Em", Detective, "Hoa", Visitor,
			"Binh", Villager, "Chi", Villager, "Dung", Villager, "Phuc", Villager, "Khoa", Villager, "Lan", Villager))
}

// The 8-player refusal is exercised through the kit in enginetest/kit_test.go;
// here the real role must carry the rule.
func TestDetective_OnlyInGamesOfMoreThanEightPlayers(t *testing.T) {
	rule, ok := engine.Registry[Detective].(engine.MinPlayersRule)
	if !ok || rule.MinPlayers() != 9 {
		t.Fatalf("the Detective must need at least 9 players, got %v", engine.Registry[Detective])
	}
	NewGame(t, Seats("A", "B", "C", "D", "E", "F", "G", "H", "I"),
		Cast("A", Detective, "B", Werewolf, "C", Villager, "D", Villager, "E", Villager,
			"F", Villager, "G", Villager, "H", Villager, "I", Villager))
}

func TestDetective_UsedOncePerGame(t *testing.T) {
	g := detectiveTable(t)
	g.Night().Detective("Em").Checks("Chi").ExpectDetectiveSaw("Chi", false)
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Detective("Em").CannotCheck("Binh", engine.ErrAbilityUsed)
}

func TestDetective_ChecksLivingPlayersIncludingHimself(t *testing.T) {
	g := detectiveTable(t)
	g.Night().Detective("Em").Checks("Em").ExpectDetectiveSaw("Em", false)

	g = detectiveTable(t)
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Chi")].Alive = false })
	g.Night().Detective("Em").CannotCheck("Chi", engine.ErrInvalidTarget)
}

func TestDetective_SeesOnlyWolfOrNoWolfInTheZone(t *testing.T) {
	tests := []struct {
		target string
		wolf   bool
	}{
		{"An", true},    // the wolf himself
		{"Binh", true},  // wolf next door (An)
		{"Chi", false},  // Binh, Dung: villagers
		{"Em", false},   // Dung, Phuc
		{"Lan", true},   // the circle closes: Lan sits next to An
		{"Hoa", true},   // Giang next door
		{"Khoa", false}, // Hoa, Lan
	}
	for _, tc := range tests {
		t.Run(tc.target, func(t *testing.T) {
			detectiveTable(t).Night().Detective("Em").Checks(tc.target).ExpectDetectiveSaw(tc.target, tc.wolf)
		})
	}
}

func TestDetective_CursedPlayerCountsAsWolf(t *testing.T) {
	g := detectiveTable(t)
	g.State().Players[g.Player("Phuc")].Team = engine.TeamWolves
	g.Night().Detective("Em").Checks("Dung").ExpectDetectiveSaw("Dung", false)
	g = detectiveTable(t)
	g.State().Players[g.Player("Phuc")].Team = engine.TeamWolves
	g.Night().Detective("Em").Checks("Em").ExpectDetectiveSaw("Em", true)
}

func TestDetective_WolfInALoverPairCountsAsWolf(t *testing.T) {
	g := NewGame(t, Seats(all...),
		Cast("An", Werewolf, "Giang", Werewolf, "Em", Detective, "Hoa", Visitor,
			"Binh", Villager, "Chi", Villager, "Dung", Villager, "Phuc", Villager, "Khoa", Villager, "Lan", Cupid))
	g.Night().Cupid("Lan").Pairs("Giang", "Khoa").Detective("Em").Checks("Phuc").ExpectDetectiveSaw("Phuc", true)
}

func TestDetective_NeighboursSkipDeadPlayers(t *testing.T) {
	// Chi's neighbours are Binh and Dung. With Binh dead the zone reaches An.
	g := detectiveTable(t)
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Binh")].Alive = false })
	g.Night().Detective("Em").Checks("Chi").ExpectDetectiveSaw("Chi", true)

	// A dead wolf next to the target does not count.
	g = detectiveTable(t)
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("An")].Alive = false })
	g.Night().Detective("Em").Checks("Binh").ExpectDetectiveSaw("Binh", false)
}

func TestDetective_FewerThanThreeLivingPlayers_ZoneIsEveryoneAlive(t *testing.T) {
	g := detectiveTable(t)
	g.HandBuild(func(s *engine.State) {
		for id, p := range s.Players {
			p.Alive = id == g.Player("Em") || id == g.Player("An")
		}
	})
	g.Night().Detective("Em").Checks("Em").ExpectDetectiveSaw("Em", true)

	g = detectiveTable(t)
	g.HandBuild(func(s *engine.State) {
		for id, p := range s.Players {
			p.Alive = id == g.Player("Em") || id == g.Player("Chi")
		}
	})
	g.Night().Detective("Em").Checks("Chi").ExpectDetectiveSaw("Chi", false)
}

func TestDetective_ResultIsSeenByTheDetectiveOnly(t *testing.T) {
	g := detectiveTable(t)
	g.Night().Detective("Em").Checks("An")
	g.As("Em").DetectiveResults(1)
	for _, other := range []string{"An", "Giang", "Hoa", "Binh", "Lan"} {
		g.As(other).DetectiveResults(0)
	}
}

func TestDetective_BlockedCannotActAndKeepsTheCharge(t *testing.T) {
	g := detectiveTable(t)
	n := g.Night().Visitor("Hoa").Visits("Em")
	n.Detective("Em").CannotCheck("An", engine.ErrBlocked)
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Detective("Em").Checks("An").ExpectDetectiveSaw("An", true)
}

func TestDetective_StillCalledEveryNightAfterUse_SameCue(t *testing.T) {
	g := detectiveTable(t)
	g.Night().Detective("Em").Checks("An")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Detective("Em").CannotCheck("An", engine.ErrAbilityUsed)

	var calls []engine.NightStepCalled
	for _, e := range g.Events() {
		if c, ok := e.(engine.NightStepCalled); ok && c.Role == Detective {
			calls = append(calls, c)
		}
	}
	if len(calls) != 2 || calls[0] != calls[1] {
		t.Fatalf("Detective calls %v: want one per night, identical", calls)
	}
}
