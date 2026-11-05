package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

// onlineTable is the Cupid table of cupid_test.go, played online (rules §1b).
func onlineTable(t testing.TB) *Game {
	return NewGame(t, Online(), Seats(all...), Chief("Hoa"),
		Cast("An", CursedWolf, "Binh", Werewolf, "Chi", Villager, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Villager, "Lan", Cupid))
}

// Rules §1b: online there is no phone check, so nights run without that step;
// the lovers' step stays, right after Cupid.
func TestOnline_NoPhoneCheck_TheLoversStepStays(t *testing.T) {
	g := onlineTable(t)
	first := []engine.RoleID{Cupid, loversMeet, Guard, CursedWolf, Werewolf, Witch, Seer}
	if got := nightCalls(g); !reflect.DeepEqual(got, first) {
		t.Fatalf("night 1 = %v, want %v", got, first)
	}
	g.Day().NobodyHanged()
	second := []engine.RoleID{Guard, CursedWolf, Werewolf, Witch, Seer}
	if got := nightCalls(g); !reflect.DeepEqual(got, second) {
		t.Fatalf("night 2 = %v, want %v", got, second)
	}
}

// Rules §1b: as soon as Cupid shoots, the two see it on their own phones, while
// Cupid's step still runs. Nobody else hears anything, and the lovers do not
// learn who Cupid is.
func TestOnline_CupidShoots_TheTwoAreToldAtOnce(t *testing.T) {
	g := onlineTable(t)
	g.Night().Cupid("Lan").Pairs("Chi", "Em")
	if step, _ := engine.CurrentNightRole(g.State()); step != Cupid {
		t.Fatalf("the step is %v, want Cupid's to run on", step)
	}
	g.As("Chi").Lover("Em", Seer, engine.LinkWhite).PhoneNoticesInFeed(1).NoPhoneNotice().NoSecretsOfOthers()
	g.As("Em").Lover("Chi", Villager, engine.LinkWhite).PhoneNoticesInFeed(1).NoPhoneNotice()
	for _, name := range []string{"An", "Binh", "Dung", "Hoa", "Lan"} {
		g.As(name).NoLover().PhoneNoticesInFeed(0)
	}
}

// Rules §7: when Cupid runs out of time the server pairs two at random; they
// are told then.
func TestOnline_CupidRunsOutOfTime_TheTwoAreToldThen(t *testing.T) {
	g := onlineTable(t)
	g.Night().Cupid("Lan").RunsOutOfTime(1)
	a, b, ok := g.Lovers()
	if !ok {
		t.Fatal("no couple after the timeout")
	}
	g.As(a).PhoneNoticesInFeed(1)
	g.As(b).PhoneNoticesInFeed(1)
	if step, _ := engine.CurrentNightRole(g.State()); step != loversMeet {
		t.Fatalf("the step is %v, want the lovers' step", step)
	}
}

// Rules §1b: the cursed player sees it at once and the pack sees its new
// member. The others are told nothing.
func TestOnline_CursedWolfCurses_TheTargetIsToldAtOnce(t *testing.T) {
	g := onlineTable(t)
	g.Night().Cupid("Lan").Pairs("Hoa", "Khoa").CursedWolf("An").Curses("Chi")
	if step, _ := engine.CurrentNightRole(g.State()); step != CursedWolf {
		t.Fatalf("the step is %v, want the Cursed wolf's to run on", step)
	}
	g.As("Chi").Cursed(true).PhoneNoticesInFeed(1).NoPhoneNotice()
	if team := g.State().Players["Chi"].Flags[engine.FlagTeamHidden]; team != 0 {
		t.Fatalf("Chi's new team is still hidden")
	}
	g.As("Binh").Pack("An", "Binh", "Chi")
	g.As("Dung").Cursed(false).PhoneNoticesInFeed(0)
}

// Rules §7: both lovers see the new colour as soon as one is cursed.
func TestOnline_ALoverIsCursed_BothSeeTheNewColourAtOnce(t *testing.T) {
	g := onlineTable(t)
	n := g.Night().Cupid("Lan").Pairs("Chi", "Em")
	n.CursedWolf("An").Curses("Chi")
	g.As("Chi").Lover("Em", Seer, engine.LinkStriped).Cursed(true).PhoneNoticesInFeed(2)
	g.As("Em").Lover("Chi", Villager, engine.LinkStriped).Cursed(false).PhoneNoticesInFeed(2)
}

// Telling once is enough: later commands of the night send nothing again.
func TestOnline_NoticesAreNotRepeated(t *testing.T) {
	g := onlineTable(t)
	n := g.Night().Cupid("Lan").Pairs("Chi", "Em")
	n.LoversMeet()
	n.Guard("Dung").Protects("Hoa")
	g.As("Chi").PhoneNoticesInFeed(1)
}

// Rules §3: the same table in person keeps both phone checks; the mode is all
// that differs.
func TestInPerson_KeepsThePhoneChecks(t *testing.T) {
	g := cupidTable(t)
	got := nightCalls(g)
	want := []engine.RoleID{Cupid, phoneCheck, loversMeet, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("night 1 = %v, want %v", got, want)
	}
}
