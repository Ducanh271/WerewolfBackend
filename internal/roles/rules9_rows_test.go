package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"

	. "masoi/internal/engine/enginetest"
)

// The rows of rules §9 "Tình huống đã chốt" not already covered elsewhere.
// The others: Guard and Witch save the same player
// (TestNight_GuardAndWitchSaveSameTarget_AnnouncesProtected), poison through
// the Guard (TestWitch_PoisonKillsAnyoneEvenHerselfAndTheGuardCannotBlockIt),
// "Không cắn ai" (TestNight_NoBiteAndGuard_NobodyIsProtectedSoTheAnnouncementIsTheUsualOne),
// the only wolf visited (TestVisitor_TheOnlyWolfVisited_NobodyIsBitten), a
// cursed lover (TestCursedWolf_CursedLover_PairTeamIsRecalculatedAtThePhoneCheck),
// a lover dying by day (TestLove_LoverHangedByDay_..., TestLove_LoverShotByTheHunter_...)
// and the cross-team couple among others (TestWin_CrossTeamCouple_...).

// Row "Nguyền trúng Tiên tri, Bảo vệ, Thợ săn": they join the wolves and keep
// their ability.
func TestRules9_CurseHitsSeerGuardOrHunter_JoinsTheWolvesAndKeepsTheAbility(t *testing.T) {
	tests := []struct {
		name   string
		cursed string
		role   engine.RoleID
		uses   func(g *Game)
	}{
		{"the Seer still checks", "Em", Seer, func(g *Game) {
			g.Night().Seer("Em").Checks("Binh").ExpectSeerSaw("Em", "Binh", true)
			g.Dawn()
		}},
		{"the Guard still shields", "Dung", Guard, func(g *Game) {
			n := g.Night().Guard("Dung").Protects("Hoa")
			n.Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa").Wolf("Dung").Bites("Hoa")
			g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
		}},
		{"the Hunter still shoots", "Phuc", Hunter, func(g *Game) {
			g.Night().Witch("Giang").Poisons("Phuc")
			g.Dawn().ExpectDead("Phuc")
			g.As("Phuc").CanShoot(true)
			g.Hunter("Phuc").Shoots("Hoa")
			g.ExpectDead("Phuc", "Hoa")
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := cursedTable(t)
			g.Night().CursedWolf("An").Curses(tc.cursed).PhoneCheck()
			g.As(tc.cursed).Cursed(true).OwnRole(tc.role).Pack("An", "Binh", tc.cursed)
			passNight(g)
			tc.uses(g)
		})
	}
}

// Row "Bầy sói cắn người vừa bị nguyền": the pack may bite the player cursed
// that very night, who dies like anybody.
func TestRules9_PackBitesThePlayerCursedTheSameNight_TheyDie(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().CursedWolf("An").Curses("Hoa").PhoneCheck()
	n.Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa") // two of the three voters
	g.Dawn().ExpectDead("Hoa").ExpectAnnouncement(i18n.DawnSomeoneDied)
}

// Row "Thợ săn chết theo người yêu hoặc vì thuốc độc": he still shoots.
func TestRules9_HunterDiesWithHisLoverInTheNight_StillShoots(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Phuc", "Chi")
	n.Wolf("An").Bites("Chi").Wolf("Binh").Bites("Chi")
	g.Dawn().ExpectDead("Chi", "Phuc").ExpectAnnouncement(i18n.HunterRevealed)
	wantDead(t, announcedDead(t, g), "Chi", "Phuc")
	g.As("Phuc").CanShoot(true)
	g.Hunter("Phuc").Shoots("Binh")
	g.ExpectDead("Chi", "Phuc", "Binh")
}

func TestRules9_HunterPoisonedByTheWitch_StillShoots(t *testing.T) {
	g := hunterTable(t)
	g.Night().Witch("Giang").Poisons("Phuc")
	g.Dawn().ExpectDead("Phuc")
	g.ExpectPhase(engine.PhaseHunterShot)
	g.As("Phuc").CanShoot(true)
	g.Hunter("Phuc").Shoots("An")
	g.ExpectDead("Phuc", "An")
}
