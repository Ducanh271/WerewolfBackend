package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// Cross-role scenarios with the lovers (task T20). cupidTable: An (Cursed
// wolf), Binh (wolf), Chi (villager), Dung (Guard), Em (Seer), Phuc (Hunter),
// Giang (Witch), Hoa (villager, Chief), Khoa (villager), Lan (Cupid).

func expectWinners(t testing.TB, g *Game, names ...string) {
	t.Helper()
	var want []engine.PlayerID
	for _, n := range names {
		want = append(want, g.Player(n))
	}
	if got := g.State().Winners; !reflect.DeepEqual(got, want) {
		t.Fatalf("winners %v, want %v", got, want)
	}
	if got := g.As(names[0]).View().Winners; !reflect.DeepEqual(got, want) {
		t.Fatalf("the end screen shows winners %v, want %v", got, want)
	}
}

// The cursed Hunter in a cross-team couple is hanged. He shoots; the win check
// after the shot counts his lover alive, on her own team (the couple is
// broken). She dies of grief in the next night, announced in the morning, and
// the win check runs again then.
func TestCross_CursedHunterInACrossTeamCouple_IsHanged(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Phuc", "Chi").PhoneCheck()
	n.CursedWolf("An").Curses("Phuc").PhoneCheck()
	g.As("Chi").Lover("Phuc", Hunter, engine.LinkStriped)
	killOff(g, "An", "Phuc", "Chi", "Dung", "Em")
	g.Dawn().ExpectNobodyDied() // 1 wolf (An) against 2 villagers: the couple counts on neither side
	g.ExpectPhase(engine.PhaseDawn)

	g.Day().Hanged("Phuc")
	g.ExpectPhase(engine.PhaseHunterShot)
	g.Hunter("Phuc").Shoots("Dung")
	// An against Em and Chi, still alive until the night: the game goes on.
	g.ExpectPhase(engine.PhaseDusk).ExpectAlive("An", "Chi", "Em")

	g.Night().Wolf("An").Bites("Em")
	g.Dawn().ExpectDead("Chi", "Em").ExpectAnnouncement(i18n.WinWolves)
	expectWinner(t, g, engine.TeamWolves)
	expectWinners(t, g, "An", "Binh", "Phuc") // the couple was broken: Phuc wins with his team
}

// The Chief is a lover and is bitten: the Hunter, her lover, dies the same
// night. The morning names both in seating order, the Hunter shoots, the win
// check runs, then the Chief names a successor.
func TestCross_ChiefLoverBittenAtNight_HunterDiesWithHer_ShootsBeforeTheTitlePasses(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Hoa", "Phuc")
	n.Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa", "Phuc")
	wantDead(t, announcedDead(t, g), "Phuc", "Hoa")
	g.ExpectPhase(engine.PhaseHunterShot)
	g.Hunter("Phuc").Shoots("Binh")
	g.ExpectPhase(engine.PhaseChiefSuccession)
	g.NameSuccessor("Hoa", "Em").ExpectChief("Em")
	want := []string{"DawnAnnounced", "HunterRevealed", "HunterShotStarted", "HunterShot",
		"ChiefSuccessionStarted", "ChiefSucceeded"}
	if got := trace(g); !reflect.DeepEqual(got, want) {
		t.Fatalf("events %v, want %v", got, want)
	}
}

// Same night, but the shot kills the last wolf: the game ends at the win check
// and the dead Chief names nobody.
func TestCross_ChiefLoverBittenAtNight_HuntersShotEndsTheGame_NoSuccession(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Hoa", "Phuc")
	killOff(g, "An", "Hoa", "Phuc", "Em", "Dung", "Chi")
	g.HandBuild(func(s *engine.State) { s.Chief = g.Player("Hoa") })
	n.Wolf("An").Bites("Hoa") // the only wolf left
	g.Dawn().ExpectDead("Hoa", "Phuc")
	g.Hunter("Phuc").Shoots("An")
	expectWinner(t, g, engine.TeamVillage)
	for _, e := range g.Events() {
		if _, ok := e.(engine.ChiefSuccessionStarted); ok {
			t.Fatal("the game is over: the Chief must not name a successor")
		}
	}
}

// The pack bites its own wolf, who is in a cross-team couple: his lover dies
// the same night, and the win check of the morning counts what is left.
func TestCross_PackBitesItsOwnWolfInACrossTeamCouple_PartnerDiesToo(t *testing.T) {
	tests := []struct {
		name      string
		survivors []string // alive before the night
		winners   []string // empty: the game goes on
	}{
		{"the game goes on", []string{"An", "Binh", "Chi", "Dung", "Em"}, nil},
		{"the wolves win", []string{"An", "Binh", "Chi", "Dung"}, []string{"An", "Binh"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := cupidTable(t)
			g.Night().Cupid("Lan").Pairs("Binh", "Chi")
			g.Dawn()
			g.Day().NobodyHanged()
			killOff(g, tc.survivors...)
			g.Night().Wolf("An").Bites("Binh").Wolf("Binh").Bites("Binh")
			g.Dawn().ExpectDead("Binh", "Chi")
			wantDead(t, announcedDead(t, g), "Binh", "Chi")
			if tc.winners == nil {
				g.ExpectPhase(engine.PhaseDawn)
				return
			}
			expectWinner(t, g, engine.TeamWolves)
			expectWinners(t, g, tc.winners...)
		})
	}
}

// Rules §8: the wolves win while the cross-team couple is still alive; the
// wolf in the couple loses with it.
func TestWin_WolvesWinWhileTheCrossTeamCoupleLives_TheWolfLoverLosesWithIt(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("An", "Em")
	killOff(g, "An", "Em", "Binh", "Chi", "Dung")
	g.Dawn()
	g.Day().Hanged("Dung") // Binh against Chi: the couple counts on neither side
	expectWinner(t, g, engine.TeamWolves)
	expectWinners(t, g, "Binh")
}

func TestWin_CrossTeamCoupleWins_OnlyTheTwoLoversAreWinners(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("An", "Em")
	killOff(g, "An", "Em", "Chi")
	g.Dawn()
	g.Day().Hanged("Chi")
	expectWinner(t, g, engine.TeamLovers)
	expectWinners(t, g, "An", "Em")
}
