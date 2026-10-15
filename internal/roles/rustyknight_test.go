package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// knightBase is the table of these scenarios: ten players in a circle An,
// Binh, Chi, Dung, Em, Phuc, Giang, Hoa, Khoa, Lan (Lan sits next to An).
// Wolves: An, Chi, Giang. Binh is the Rusty knight. Dung is the Guard, Em the
// Seer, Hoa the Witch (also the Chief, who never dies in these scenarios); the
// others are villagers. No Visitor: a Visitor who sleeps through her step
// visits somebody at random (rules §6).
var knightBase = map[string]engine.RoleID{
	"An": Werewolf, "Binh": RustyKnight, "Chi": Werewolf, "Dung": Guard, "Em": Seer,
	"Phuc": Villager, "Giang": Werewolf, "Hoa": Witch, "Khoa": Villager, "Lan": Villager,
}

// knightTableWith moves the Rusty knight to another seat; Binh takes over the
// role that seat had.
func knightTableWith(t testing.TB, knight string) *Game {
	roles := map[string]engine.RoleID{}
	for name, role := range knightBase {
		roles[name] = role
	}
	if knight != "Binh" {
		roles["Binh"], roles[knight] = roles[knight], RustyKnight
	}
	var cast []any
	for _, name := range all {
		cast = append(cast, name, roles[name])
	}
	return NewGame(t, Seats(all...), Chief("Hoa"), Cast(cast...))
}

func knightTable(t testing.TB) *Game { return knightTableWith(t, "Binh") }

// bite has the named wolves vote for the victim.
func bite(n *Night, victim string, wolves ...string) *Night {
	for _, w := range wolves {
		n = n.Wolf(w).Bites(victim)
	}
	return n
}

func TestRustyKnight_BittenByThePack_KnightDiesNow_NearestWolfOnHisLeftDiesNextNight(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn().ExpectDead("Binh") // the wolf is poisoned, not dead yet
	g.As("Em").PublicTable().Dead("Binh")
	g.Day().NobodyHanged()

	bite(g.Night(), "Phuc", "An", "Giang") // Chi is dead by the rules: she cannot vote
	g.Dawn().ExpectDead("Chi", "Phuc").ExpectAnnouncement(i18n.DawnSomeoneDied)
	g.ExpectDawnNames("Chi", "Phuc") // seating order, no cause
}

func TestRustyKnight_WhichWolfIsPoisoned(t *testing.T) {
	tests := []struct {
		name   string
		knight string
		dead   []string // dead before the bite
		cursed string   // a villager on the wolf team
		voters []string
		want   string
	}{
		{"next seat is a wolf", "Binh", nil, "", []string{"An", "Chi", "Giang"}, "Chi"},
		{"dead wolf is skipped", "Binh", []string{"Chi"}, "", []string{"An", "Giang"}, "Giang"},
		{"villagers between are skipped", "Em", nil, "", []string{"An", "Chi", "Giang"}, "Giang"},
		{"cursed villager counts as a wolf", "Em", nil, "Phuc", []string{"An", "Chi", "Giang", "Phuc"}, "Phuc"},
		{"the circle closes after the last seat", "Lan", nil, "", []string{"An", "Chi", "Giang"}, "An"},
		{"a villager and the Visitor in between", "Khoa", nil, "", []string{"An", "Chi", "Giang"}, "An"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := knightTableWith(t, tc.knight)
			n := g.Night() // the Chief election is over: now change the table
			g.HandBuild(func(s *engine.State) {
				for _, name := range tc.dead {
					s.Players[g.Player(name)].Alive = false
				}
				if tc.cursed != "" {
					s.Players[g.Player(tc.cursed)].Team = engine.TeamWolves
				}
			})
			bite(n, tc.knight, tc.voters...)
			g.Dawn().ExpectDead(tc.knight)
			g.Day().NobodyHanged()
			g.Night()
			g.Dawn().ExpectDead(tc.want)
		})
	}
}

func TestRustyKnight_GuardShieldsHim_NobodyDiesAndNoWolfIsPoisoned(t *testing.T) {
	g := knightTable(t)
	bite(g.Night().Guard("Dung").Protects("Binh"), "Binh", "An", "Chi")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectNobodyDied()
}

func TestRustyKnight_WitchHealsHim_KnightLivesButTheWolfIsStillPoisoned(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi").Witch("Hoa").Heals()
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectDead("Chi")
}

func TestRustyKnight_GuardAndWitchBothSaveHim_NoWolfIsPoisoned(t *testing.T) {
	g := knightTable(t)
	bite(g.Night().Guard("Dung").Protects("Binh"), "Binh", "An", "Chi").Witch("Hoa").Heals()
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectNobodyDied()
}

func TestRustyKnight_OnlyTheBiteTriggers(t *testing.T) {
	t.Run("poisoned by the Witch", func(t *testing.T) {
		g := knightTable(t)
		g.Night().Witch("Hoa").Poisons("Binh")
		g.Dawn().ExpectDead("Binh")
		g.Day().NobodyHanged()
		g.Night()
		g.Dawn().ExpectNobodyDied()
	})
	t.Run("hanged", func(t *testing.T) {
		g := knightTable(t)
		g.Night()
		g.Dawn().ExpectNobodyDied()
		g.Day().Hanged("Binh")
		g.Night()
		g.Dawn().ExpectNobodyDied()
	})
	t.Run("shot by the Hunter", func(t *testing.T) {
		g := knightTable(t)
		g.HandBuild(func(s *engine.State) { s.Players[g.Player("Khoa")].Role = Hunter })
		bite(g.Night(), "Khoa", "An", "Chi")
		g.Dawn().ExpectDead("Khoa")
		g.Hunter("Khoa").Shoots("Binh")
		g.Day().NobodyHanged()
		g.Night()
		g.Dawn().ExpectNobodyDied()
	})
}

func TestRustyKnight_PoisonedWolfDiesEarlier_NothingMoreHappens(t *testing.T) {
	t.Run("hanged the next day", func(t *testing.T) {
		g := knightTable(t)
		bite(g.Night(), "Binh", "An", "Chi")
		g.Dawn().ExpectDead("Binh")
		g.Day().Hanged("Chi")
		bite(g.Night(), "Phuc", "An", "Giang")
		g.Dawn().ExpectDead("Phuc")
	})
	t.Run("poisoned by the Witch the same night", func(t *testing.T) {
		g := knightTable(t)
		bite(g.Night(), "Binh", "An", "Chi").Witch("Hoa").Poisons("Chi")
		g.Dawn().ExpectDead("Binh", "Chi")
		g.Day().NobodyHanged()
		g.Night()
		g.Dawn().ExpectNobodyDied()
	})
}

func TestRustyKnight_PoisonedWolfIsTheOnlyWolf_NobodyIsBittenThatNight(t *testing.T) {
	g := NewGame(t, Seats(all...), Chief("Hoa"),
		Cast("An", Villager, "Binh", RustyKnight, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Villager, "Giang", Villager, "Hoa", Witch, "Khoa", Villager, "Lan", Villager))
	g.Night().Wolf("Chi").Bites("Binh")
	g.Dawn().ExpectDead("Binh")
	g.Day().NobodyHanged()

	g.Night().Wolf("Chi").CannotBite("Phuc", engine.ErrNotYourTurn)
	g.Dawn().ExpectDead("Chi").ExpectAnnouncement(i18n.DawnSomeoneDied)
	g.ExpectDawnNames("Chi")
}

func TestRustyKnight_PoisonedWolfCannotVote_PackCountsTheOthers(t *testing.T) {
	g := knightTable(t)
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn()
	g.Day().NobodyHanged()

	// An alone is one vote of the two wolves left: not more than half.
	g.Night().Wolf("Chi").CannotBite("Phuc", engine.ErrNotYourTurn).Wolf("An").Bites("Phuc")
	g.Dawn().ExpectDead("Chi")
}

func TestRustyKnight_CursedKnightKeepsTheAbility(t *testing.T) {
	g := knightTable(t)
	g.HandBuild(func(s *engine.State) { s.Players[g.Player("Binh")].Team = engine.TeamWolves })
	bite(g.Night(), "Binh", "An", "Chi", "Giang") // 3 of the 4 wolves bite their own Knight
	g.Dawn().ExpectDead("Binh")
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectDead("Chi")
}

func TestRustyKnight_TheVisitorCannotBlockIt(t *testing.T) {
	g := NewGame(t, Seats(all...), Chief("Hoa"),
		Cast("An", Werewolf, "Binh", RustyKnight, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Villager, "Giang", Werewolf, "Hoa", Witch, "Khoa", Villager, "Lan", Visitor))
	bite(g.Night().Visitor("Lan").Visits("Binh"), "Binh", "An", "Chi")
	g.Dawn().ExpectDead("Binh")
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectDead("Chi")
}

func TestRustyKnight_PoisonedWolfWhoIsHunter_ShootsAfterTheAnnouncement(t *testing.T) {
	g := knightTable(t)
	g.HandBuild(func(s *engine.State) {
		s.Players[g.Player("Chi")].Role = Hunter // a cursed Hunter
	})
	bite(g.Night(), "Binh", "An", "Chi")
	g.Dawn().ExpectDead("Binh")
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectDead("Chi")
	g.As("Chi").CanShoot(true)
	g.Hunter("Chi").Shoots("Phuc")
}

func TestRustyKnight_PoisonedWolfWhoIsChief_ChoosesASuccessorAfterTheAnnouncement(t *testing.T) {
	g := knightTable(t)
	n := g.Night()
	g.HandBuild(func(s *engine.State) { s.Chief = g.Player("Chi") })
	bite(n, "Binh", "An", "Chi")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night()
	g.Dawn().ExpectDead("Chi")
	g.ExpectPhase(engine.PhaseChiefSuccession)
	g.NameSuccessor("Chi", "Phuc").ExpectChief("Phuc")
}
