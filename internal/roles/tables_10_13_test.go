package roles_test

import (
	"errors"
	"fmt"
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/i18n"
)

// Tables of 10-13 players are the common case, and the only ones where the
// Cursed wolf and the Wolf seer share the wolf limit of 3 (rules §2).

// thirteen are the seats of the 13-player table.
var thirteen = append(append([]string(nil), all...), "Mai", "Nga", "Oanh")

// Variant A, 10 players: Cursed wolf (An) + Wolf seer (Binh) and no plain wolf,
// the full set of village roles. The curse takes the third wolf slot.
func TestTable10_CursedWolfAndWolfSeerWithoutPlainWolf_PlaysToWolvesWin(t *testing.T) {
	g := NewGame(t, Seats(all...), Chief("Giang"),
		Cast("An", CursedWolf, "Binh", WolfSeer, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Visitor, "Khoa", Villager, "Lan", Villager))

	// Night 1: Lan is cursed; the pack knows the Seer and bites her, but the Guard shields her.
	n := g.Night().Visitor("Hoa").Visits("Khoa").Guard("Dung").Protects("Em")
	n.CursedWolf("An").Curses("Lan").PhoneCheck()
	n.WolfSeer("Binh").Checks("Em").ExpectWolfSeerSaw("Em", true)
	n.Wolf("An").Bites("Em").Wolf("Binh").Bites("Em").Wolf("Lan").Bites("Em")
	n.Seer("Em").Checks("An").ExpectSeerSaw("Em", "An", true)
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
	g.Day().Hanged("Khoa")

	// Night 2: the Guard guards someone else, the Seer dies.
	n = g.Night().Visitor("Hoa").Visits("Chi").Guard("Dung").Protects("Dung")
	n.WolfSeer("Binh").Checks("Phuc").ExpectWolfSeerSaw("Phuc", false)
	n.Wolf("An").Bites("Em").Wolf("Binh").Bites("Em").Wolf("Lan").Bites("Em")
	g.Dawn().ExpectDead("Em")
	g.Day().Hanged("Chi")

	// Night 3: 3 wolves against 3 villagers: the wolves win at dawn.
	n = g.Night().Visitor("Hoa").Visits("Phuc").Guard("Dung").Protects("Giang")
	n.Wolf("An").Bites("Dung").Wolf("Binh").Bites("Dung").Wolf("Lan").Bites("Dung")
	g.Dawn().ExpectDead("Dung").ExpectAnnouncement(i18n.WinWolves)
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamWolves {
		t.Fatalf("winner %v, want wolves", w)
	}
}

// Variant B, 13 players: Wolf seer (An) + 2 plain wolves, no Cursed wolf.
func TestTable13_WolfSeerAndTwoPlainWolves_PlaysToVillageWin(t *testing.T) {
	g := NewGame(t, Seats(thirteen...), Chief("Mai"),
		Cast("An", WolfSeer, "Binh", Werewolf, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Villager, "Lan", Villager,
			"Mai", Villager, "Nga", Villager, "Oanh", Visitor))

	n := g.Night().Visitor("Oanh").Visits("Nga")
	n.WolfSeer("An").Checks("Em").ExpectWolfSeerSaw("Em", true)
	n.Wolf("An").Bites("Em").Wolf("Binh").Bites("Em").Wolf("Chi").Bites("Em")
	g.Dawn().ExpectDead("Em")
	g.Day().Hanged("An")

	n = g.Night().Visitor("Oanh").Visits("Mai")
	n.Wolf("Binh").Bites("Giang").Wolf("Chi").Bites("Giang")
	g.Dawn().ExpectDead("Giang")
	g.Day().Hanged("Binh")

	n = g.Night().Visitor("Oanh").Visits("Nga")
	n.Wolf("Chi").Bites("Lan")
	g.Dawn().ExpectDead("Lan")
	g.Day().Hanged("Chi")
	g.ExpectPhase(engine.PhaseEnded)
	if w := g.State().Winner; w == nil || *w != engine.TeamVillage {
		t.Fatalf("winner %v, want village", w)
	}
}

// lobbyOf is a lobby of n seated players.
func lobbyOf(t *testing.T, n int) engine.State {
	t.Helper()
	s := engine.NewState()
	order := make([]engine.PlayerID, n)
	for i := range order {
		order[i] = engine.PlayerID(fmt.Sprintf("P%d", i+1))
		s = engine.Fold(s, engine.PlayerJoined{Player: order[i], Nickname: string(order[i])})
	}
	return engine.Fold(s, engine.SeatsSet{Order: order})
}

// villageRoles is the full village side of an n-player table next to the given
// wolf roles, with villagers filling the rest.
func villageRoles(n, wolves int, wolfRoles map[engine.RoleID]int) map[engine.RoleID]int {
	list := map[engine.RoleID]int{Guard: 1, Seer: 1, Witch: 1, Hunter: 1, Visitor: 1}
	for role, c := range wolfRoles {
		list[role] = c
	}
	list[Villager] = n - 5 - wolves
	return list
}

func TestTables10To13_WolfLimitIsThree_InBothVariants(t *testing.T) {
	for n := 10; n <= 13; n++ {
		variants := []struct {
			name      string
			wolves    int // wolves on the table at the deal
			allowed   bool
			wolfRoles map[engine.RoleID]int
		}{
			{"A: Cursed wolf + Wolf seer", 2, true, map[engine.RoleID]int{CursedWolf: 1, WolfSeer: 1}},
			{"A + a plain wolf: the curse slot is over the limit", 3, false, map[engine.RoleID]int{CursedWolf: 1, WolfSeer: 1, Werewolf: 1}},
			{"B: Wolf seer + 2 plain wolves", 3, true, map[engine.RoleID]int{WolfSeer: 1, Werewolf: 2}},
			{"B + a third plain wolf", 4, false, map[engine.RoleID]int{WolfSeer: 1, Werewolf: 3}},
		}
		for _, v := range variants {
			t.Run(fmt.Sprintf("%d players, %s", n, v.name), func(t *testing.T) {
				list := villageRoles(n, v.wolves, v.wolfRoles)
				s := lobbyOf(t, n)
				_, err := engine.Decide(s, engine.ConfigureRoles{Roles: list})
				if !v.allowed {
					if !errors.Is(err, engine.ErrTooManyWolves) {
						t.Fatalf("err = %v, want ErrTooManyWolves", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("a legal table was refused: %v", err)
				}
				for seed := int64(1); seed <= 20; seed++ { // whatever the shuffle
					g := NewGame(t, Seats(thirteen[:n]...), Roles(list), Seed(seed))
					wolves := 0
					for _, id := range g.State().Seats {
						if g.State().Players[id].Team == engine.TeamWolves {
							wolves++
						}
					}
					if wolves != v.wolves {
						t.Fatalf("seed %d: %d wolves dealt, want %d", seed, wolves, v.wolves)
					}
				}
			})
		}
	}
}
