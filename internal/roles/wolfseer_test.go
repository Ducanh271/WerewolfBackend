package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

// wolfSeerTable is the ten-player table with the Wolf seer (An) and the Cursed
// wolf (Binh), which fill the wolf limit of 3 between them (rules §2), the
// Guard (Dung), the Seer (Em), the Hunter (Phuc), the Witch (Giang) and four
// villagers (Chi, Hoa, Khoa, Lan).
func wolfSeerTable(t testing.TB) *Game {
	return NewGame(t,
		Seats(all...),
		Cast("An", WolfSeer, "Binh", CursedWolf, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Villager, "Lan", Villager))
}

// wolfSeerVisitorTable is the same table with the Visitor in Khoa's seat.
func wolfSeerVisitorTable(t testing.TB) *Game {
	return NewGame(t,
		Seats(all...),
		Cast("An", WolfSeer, "Binh", CursedWolf, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Visitor, "Lan", Villager))
}

func TestWolfSeer_IsCalledEveryNight_AfterCursedWolfAndPhoneCheck_BeforePack(t *testing.T) {
	g := wolfSeerTable(t)
	for night := 1; night <= 4; night++ {
		g.Night()
		passNight(g)
	}
	withCurse := []engine.RoleID{Guard, CursedWolf, engine.StepPhoneCheck, WolfSeer, Werewolf, Witch, Seer}
	for night := 1; night <= 3; night++ {
		if got := stepsOfNight(g, night); !reflect.DeepEqual(got, withCurse) {
			t.Errorf("night %d calls %v, want %v", night, got, withCurse)
		}
	}
	if got, want := stepsOfNight(g, 4), []engine.RoleID{Guard, WolfSeer, Werewolf, Witch, Seer}; !reflect.DeepEqual(got, want) {
		t.Errorf("night 4 calls %v, want %v", got, want)
	}
}

func TestWolfSeer_ChecksOnePlayer_ResultIsSeerOrNot(t *testing.T) {
	tests := []struct {
		target string
		seer   bool
	}{
		{"Em", true},
		{"Chi", false},
		{"Dung", false},
		{"Giang", false},
		{"Binh", false}, // another wolf
		{"An", false},   // himself
	}
	for _, tc := range tests {
		t.Run(tc.target, func(t *testing.T) {
			g := wolfSeerTable(t)
			g.Night().WolfSeer("An").Checks(tc.target).ExpectWolfSeerSaw(tc.target, tc.seer)
		})
	}
}

func TestWolfSeer_ChecksTwiceOrTheDead_IsRefused(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night()
	n.WolfSeer("An").Checks("Chi").WolfSeer("An").CannotCheck("Em", engine.ErrAlreadyActed)

	g = wolfSeerTable(t)
	g.Night().Wolf("Binh").Bites("Hoa").Wolf("An").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa")
	g.Day().NobodyHanged()
	g.Night().WolfSeer("An").CannotCheck("Hoa", engine.ErrInvalidTarget)
}

func TestWolfSeer_ResultIsSharedWithThePack_BeforeTheyChooseWhomToBite(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night().WolfSeer("An").Checks("Em")
	g.As("An").WolfSeerResults(1).WolfVotesInFeed(0)
	g.As("Binh").WolfSeerResults(1).WolfVotesInFeed(0)
	for _, other := range []string{"Em", "Chi", "Dung", "Giang"} {
		g.As(other).WolfSeerResults(0)
	}
	n.Wolf("Binh").Bites("Em") // the pack now bites on what it knows
}

func TestWolfSeer_ResultReachesAPlayerCursedEarlierThatNight_FromThePhoneCheckOn(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night().CursedWolf("Binh").Curses("Chi").PhoneCheck()
	n.WolfSeer("An").Checks("Em")
	g.As("Chi").Cursed(true).Pack("An", "Binh", "Chi").WolfSeerResults(1)
	g.As("Hoa").WolfSeerResults(0)
}

func TestWolfSeer_ViewShowsThePackLikeAnyWolf_CursedPlayerIncludedAfterThePhoneCheck(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night().CursedWolf("Binh").Curses("Chi")
	g.As("An").Pack("An", "Binh")
	n.PhoneCheck()
	g.As("An").Pack("An", "Binh", "Chi").NoRoleExceptOwn()
	g.As("Binh").Pack("An", "Binh", "Chi")
	g.As("Hoa").NoPack()
}

func TestWolfSeer_CursedSeerIsStillSeer_RoleCheckNotTeamCheck(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night().CursedWolf("Binh").Curses("Em").PhoneCheck()
	n.WolfSeer("An").Checks("Em").ExpectWolfSeerSaw("Em", true)
}

func TestWolfSeer_BitesWithThePackAndCountsTowardTheVote(t *testing.T) {
	g := wolfSeerTable(t)
	n := g.Night()
	n.Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	g.As("Binh").WolfVote("An", "Hoa").WolfVote("Binh", "Hoa")
	g.Dawn().ExpectDead("Hoa")

	// Two voters: one vote is not more than half, so nobody is bitten.
	g = wolfSeerTable(t)
	g.Night().Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Lan")
	g.Dawn().ExpectNobodyDied()
}

func TestWolfSeer_SeerCheckingTheWolfSeer_SeesWolf(t *testing.T) {
	g := wolfSeerTable(t)
	g.Night().Seer("Em").Checks("An").ExpectSeerSaw("Em", "An", true)
}

func TestWolfSeer_BlockedByVisitor_LosesTheCheckAndTheBiteVote(t *testing.T) {
	g := wolfSeerVisitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("An")
	n.WolfSeer("An").CannotCheck("Em", engine.ErrBlocked)
	g.As("An").Blocked(true)
	n.Wolf("An").CannotBite("Hoa", engine.ErrBlocked)
	g.As("Binh").WolfSeerResults(0)

	// The pack counts over the wolves still able to vote: Binh alone bites.
	n.Wolf("Binh").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa")
}
