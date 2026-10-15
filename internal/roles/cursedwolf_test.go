package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/roles"
)

// cursedTable is the ten-player table with the Cursed wolf: the Cursed wolf
// (An), one wolf (Binh), the Guard (Dung), the Seer (Em), the Hunter (Phuc),
// the Witch (Giang) and four villagers (Chi, Hoa, Khoa, Lan). With a Cursed
// wolf the wolf limit leaves room for one plain wolf only (rules §2).
func cursedTable(t testing.TB) *Game {
	return NewGame(t,
		Seats(all...),
		Cast("An", CursedWolf, "Binh", Werewolf, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Villager, "Lan", Villager))
}

// cursedVisitorTable is the same table with the Visitor in Khoa's seat.
func cursedVisitorTable(t testing.TB) *Game {
	return NewGame(t,
		Seats(all...),
		Cast("An", CursedWolf, "Binh", Werewolf, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Visitor, "Lan", Villager))
}

// passNight ends the night and a quiet day, ready for the next night.
func passNight(g *Game) {
	g.Dawn()
	g.Day().NobodyHanged()
}

// stepsOfNight lists the steps called on the given night.
func stepsOfNight(g *Game, night int) []engine.RoleID {
	var steps []engine.RoleID
	in := false
	for _, e := range g.Events() {
		switch ev := e.(type) {
		case engine.NightStarted:
			in = ev.Night == night
		case engine.NightStepCalled:
			if in {
				steps = append(steps, ev.Role)
			}
		}
	}
	return steps
}

func TestCursedWolf_IsCalledBeforeWolfSeerAndPack_OnNightsOneToThreeOnly(t *testing.T) {
	g := cursedTable(t)
	withCurse := []engine.RoleID{Guard, CursedWolf, engine.StepPhoneCheck, Werewolf, Witch, Seer}
	for night := 1; night <= 4; night++ {
		g.Night()
		passNight(g)
	}
	for night := 1; night <= 3; night++ {
		if got := stepsOfNight(g, night); !reflect.DeepEqual(got, withCurse) {
			t.Errorf("night %d calls %v, want %v", night, got, withCurse)
		}
	}
	if got, want := stepsOfNight(g, 4), []engine.RoleID{Guard, Werewolf, Witch, Seer}; !reflect.DeepEqual(got, want) {
		t.Errorf("night 4 calls %v, want %v: no curse, no phone check", got, want)
	}
}

func TestCursedWolf_CursesAnyLivingNonWolfExceptHimself(t *testing.T) {
	for _, target := range []string{"Chi", "Dung", "Em", "Phuc", "Giang", "Hoa"} {
		t.Run("can curse "+target, func(t *testing.T) {
			g := cursedTable(t)
			g.Night().CursedWolf("An").Curses(target)
			if team := g.State().Players[g.Player(target)].Team; team != engine.TeamWolves {
				t.Fatalf("%s is on team %v after the curse, want the wolves", target, team)
			}
		})
	}
	invalid := []struct {
		name   string
		target string
		dead   string
	}{
		{"himself", "An", ""},
		{"another wolf", "Binh", ""},
		{"a dead player", "Hoa", "Hoa"},
	}
	for _, tc := range invalid {
		t.Run("cannot curse "+tc.name, func(t *testing.T) {
			g := cursedTable(t)
			if tc.dead != "" {
				g.HandBuild(func(s *engine.State) { s.Players[engine.PlayerID(tc.dead)].Alive = false })
			}
			g.Night().CursedWolf("An").CannotCurse(tc.target, engine.ErrInvalidTarget)
		})
	}
}

func TestCursedWolf_CursesOnceInTheGame(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().CursedWolf("An").Curses("Phuc")
	n.DoExpectErr(roles.CursedWolfCurse{By: g.Player("An"), Target: g.Player("Em")}, engine.ErrAlreadyActed)
	passNight(g)
	g.Night().CursedWolf("An").CannotCurse("Em", engine.ErrAlreadyActed)
	if g.State().Players[g.Player("Em")].Team != engine.TeamVillage {
		t.Fatal("a second curse must change nothing")
	}
}

func TestCursedWolf_OnlyHeCanCurse(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().TurnOf(CursedWolf)
	n.DoExpectErr(roles.CursedWolfCurse{By: g.Player("Binh"), Target: g.Player("Em")}, engine.ErrNotYourTurn)
	n.DoExpectErr(roles.CursedWolfCurse{By: g.Player("Em"), Target: g.Player("Phuc")}, engine.ErrNotYourTurn)
}

func TestCursedWolf_CurseIsImmediate_TargetKeepsAbilityAndBitesWithThePackThatNight(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().CursedWolf("An").Curses("Phuc")
	if p := g.State().Players[g.Player("Phuc")]; p.Team != engine.TeamWolves || p.Role != Hunter {
		t.Fatalf("cursed player is %v/%v, want the wolves with the Hunter role", p.Team, p.Role)
	}
	n.PhoneCheck()
	g.As("Phuc").Cursed(true).OwnRole(Hunter).Pack("An", "Binh", "Phuc")
	g.As("Binh").Pack("An", "Binh", "Phuc")
	g.As("Hoa").NoPack()

	n.Wolf("Phuc").Bites("Em").Wolf("Binh").Bites("Em")
	g.Dawn().ExpectDead("Em")
}

func TestCursedWolf_NewMemberAndCurseShowOnlyAtThePhoneCheck(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().CursedWolf("An").Curses("Phuc")

	phuc := g.As("Phuc").Cursed(false).NoPack().PhoneNoticesInFeed(0).NoPhoneNotice()
	if team := phuc.View().Me.Team; team != engine.TeamVillage {
		t.Fatalf("before the phone check the cursed player's card shows %v, want the village", team)
	}
	g.As("Phuc").NoEvent(engine.TeamChanged{Player: g.Player("Phuc"), Team: engine.TeamWolves})
	g.As("Binh").Pack("An", "Binh")
	g.As("An").Event(roles.PlayerCursed{Wolf: g.Player("An"), Target: g.Player("Phuc")})
	g.As("Binh").NoEvent(roles.PlayerCursed{Wolf: g.Player("An"), Target: g.Player("Phuc")})

	n.PhoneCheck()
	g.As("Phuc").Cursed(true).PhoneNotice(true).PhoneNoticesInFeed(1)
	g.As("Binh").Pack("An", "Binh", "Phuc")
}

func TestCursedWolf_NeverCurses_BecomesAPlainWolfAfterNightThree(t *testing.T) {
	g := cursedTable(t)
	for night := 1; night <= 3; night++ {
		g.Night()
		passNight(g)
	}
	n := g.Night()
	n.DoExpectErr(roles.CursedWolfCurse{By: g.Player("An"), Target: g.Player("Phuc")}, engine.ErrNotYourTurn)
	g.As("An").OwnRole(CursedWolf).Cursed(false)
	n.Wolf("An").Bites("Em").Wolf("Binh").Bites("Em")
	g.Dawn().ExpectDead("Em")
}

func TestCursedWolf_BlockedOnNightOneOrTwo_CursesLater_BlockedOnNightThree_CurseIsLost(t *testing.T) {
	tests := []struct {
		name         string
		blockedNight int
		curseNight   int // 0: the curse is lost
	}{
		{"blocked on night 1", 1, 2},
		{"blocked on night 2", 2, 3},
		{"blocked on night 3", 3, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := cursedVisitorTable(t)
			for night := 1; night <= 4; night++ {
				n := g.Night()
				// The Visitor never visits the same player twice in a row.
				visit := map[bool]string{true: "Hoa", false: "Lan"}[night%2 == 0]
				if night == tc.blockedNight {
					visit = "An"
				}
				n.Visitor("Khoa").Visits(visit)
				switch {
				case night == tc.blockedNight:
					n.CursedWolf("An").CannotCurse("Phuc", engine.ErrBlocked)
				case night == tc.curseNight:
					n.CursedWolf("An").Curses("Phuc")
				case night > 3:
					if steps := stepsOfNight(g, night); contains(steps, CursedWolf) {
						t.Fatalf("night %d calls the Cursed wolf: %v", night, steps)
					}
				}
				passNight(g)
			}
			cursed := g.State().Players[g.Player("Phuc")].Team == engine.TeamWolves
			if cursed != (tc.curseNight != 0) {
				t.Fatalf("Phuc cursed=%v, want %v", cursed, tc.curseNight != 0)
			}
		})
	}
}

func contains(steps []engine.RoleID, role engine.RoleID) bool {
	for _, s := range steps {
		if s == role {
			return true
		}
	}
	return false
}

func TestCursedWolf_GuardCannotBlockTheCurse(t *testing.T) {
	g := cursedTable(t)
	g.Night().Guard("Dung").Protects("Phuc").CursedWolf("An").Curses("Phuc")
	if g.State().Players[g.Player("Phuc")].Team != engine.TeamWolves {
		t.Fatal("the Guard's shield must not stop the curse")
	}
}

func TestCursedWolf_SeerSeesTheCursedPlayerAsAWolf(t *testing.T) {
	tests := []struct {
		name  string
		curse bool
		wolf  bool
	}{
		{"not cursed", false, false},
		{"cursed the same night", true, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := cursedTable(t)
			n := g.Night()
			if tc.curse {
				n.CursedWolf("An").Curses("Phuc")
			}
			n.Seer("Em").Checks("Phuc")
			n.ExpectSeerSaw("Em", "Phuc", tc.wolf)
		})
	}
	// The Detective reads the same team; see detective_test.go.
}

func TestCursedWolf_CursedLover_PairTeamIsRecalculatedAtThePhoneCheck(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Em", "Phuc").PhoneCheck()
	g.As("Em").Lover("Phuc", Hunter, engine.LinkWhite) // told at the first phone check
	g.Dawn()
	g.Day().NobodyHanged()

	n := g.Night().CursedWolf("An").Curses("Phuc")
	if team := g.State().Players[g.Player("Phuc")].Team; team != engine.TeamWolves {
		t.Fatalf("the engine's team of the cursed lover is %v, want the wolves at once", team)
	}
	g.As("Em").Lover("Phuc", Hunter, engine.LinkWhite) // the phones still show the old colour

	n.PhoneCheck()
	g.As("Em").Lover("Phuc", Hunter, engine.LinkStriped)
	g.As("Phuc").Lover("Em", Seer, engine.LinkStriped)
}
