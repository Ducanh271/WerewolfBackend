package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

func TestView_Seer_SeesOnlyOwnResults(t *testing.T) {
	g := newTable(t)
	g.Night().Seer("Em").Checks("An")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Seer("Em").Checks("Phuc")

	g.As("Em").SeerResults(2)
	for _, other := range []string{"An", "Binh", "Dung", "Phuc", "Hoa"} {
		g.As(other).SeerResults(0)
	}
}

// A cursed Seer keeps her power (rules §5); her results stay hers, the pack
// does not get them.
func TestView_CursedSeer_ResultsStayHers(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().CursedWolf("An").Curses("Em")
	n.Seer("Em").Checks("Giang")
	g.As("Em").SeerResults(1)
	g.As("An").SeerResults(0)
}

// witchTable is the usual table with the Witch in Lan's seat; she still has
// her heal potion.
func witchTable(t *testing.T) *Game {
	return NewGame(t,
		Seats(all...),
		Cast("An", Werewolf, "Binh", Werewolf, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Villager, "Giang", Villager, "Hoa", Villager, "Khoa", Villager, "Lan", Witch))
}

func TestView_Witch_SeesTheBittenOnlyOnHerTurnWhileSheCanHeal(t *testing.T) {
	bitePhuc := func(n *Night) *Night { return n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc") }
	tests := []struct {
		name   string
		night  func(g *Game, n *Night)
		bitten string
	}{
		{"her turn, heal potion left", func(_ *Game, n *Night) { bitePhuc(n).TurnOf(Witch) }, "Phuc"},
		{"protected by the Guard: she cannot tell", func(_ *Game, n *Night) {
			bitePhuc(n.Guard("Dung").Protects("Phuc")).TurnOf(Witch)
		}, "Phuc"},
		{"heal potion used", func(g *Game, n *Night) {
			g.HandBuild(func(s *engine.State) { s.Players["Lan"].Flags = map[string]int{"witch.heal_used": 1} })
			bitePhuc(n).TurnOf(Witch)
		}, ""},
		{"before her turn", func(_ *Game, n *Night) { bitePhuc(n) }, ""},
		{"after her turn", func(_ *Game, n *Night) { bitePhuc(n).TurnOf(Seer) }, ""},
		{"the pack did not agree", func(_ *Game, n *Night) {
			n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Giang").TurnOf(Witch)
		}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := witchTable(t)
			tc.night(g, g.Night())
			g.As("Lan").Bitten(tc.bitten)
			g.As("Em").Bitten("") // nobody else, ever
		})
	}
}

func TestView_CursedPlayer_IsNotifiedAtThePhoneCheck_AndOthersAreToldTheyAreNot(t *testing.T) {
	g := cursedTable(t)
	n := g.Night()
	g.As("Phuc").Cursed(false)
	n.CursedWolf("An").Curses("Phuc")
	g.As("Phuc").Cursed(false).NoPhoneNotice() // not before the phone check
	n.PhoneCheck()
	notice := func(who string, cursed bool) engine.PhoneChecked {
		return engine.PhoneChecked{Player: g.Player(who), Cursed: cursed}
	}
	g.As("Phuc").Cursed(true).PhoneNotice(true).Event(notice("Phuc", true)).NoEvent(notice("Phuc", false))
	g.As("An").Cursed(false).PhoneNotice(false).Event(notice("An", false))
	g.As("Hoa").PhoneNotice(false).Event(notice("Hoa", false)).NoEvent(notice("Phuc", true))
}

func TestView_Lovers_SeePartnerRoleAndLinkColour_CupidSeesNothing(t *testing.T) {
	tests := []struct {
		name   string
		a, b   string
		curse  string // cursed on the same night, before the phone check, if any
		colour engine.LinkColour
		roleA  engine.RoleID
		roleB  engine.RoleID
	}{
		{"two villagers: white", "Em", "Chi", "", engine.LinkWhite, Seer, Villager},
		{"two wolves: black", "An", "Binh", "", engine.LinkBlack, CursedWolf, Werewolf},
		{"a wolf and a villager: striped", "Em", "An", "", engine.LinkStriped, Seer, CursedWolf},
		{"one lover cursed: the colour changes", "Em", "Chi", "Chi", engine.LinkStriped, Seer, Villager},
		{"a lover cursed next to a wolf: black", "Em", "An", "Em", engine.LinkBlack, Seer, CursedWolf},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := cupidTable(t)
			n := g.Night().Cupid("Lan").Pairs(tc.a, tc.b)
			g.As(tc.a).NoLover() // nobody is told before the phone check
			g.As(tc.b).NoLover()
			if tc.curse != "" {
				n.CursedWolf("An").Curses(tc.curse)
			}
			n.PhoneCheck()
			g.As(tc.a).Lover(tc.b, tc.roleB, tc.colour).RoleOf(tc.b, tc.roleB)
			g.As(tc.b).Lover(tc.a, tc.roleA, tc.colour).RoleOf(tc.a, tc.roleA)
			g.As("Lan").OwnRole(Cupid).NoLover().NoRoleExceptOwn()
			g.As("Hoa").NoLover().NoRoleExceptOwn()
		})
	}
}

func TestView_DeadPlayers_SeeEverything(t *testing.T) {
	g := newTable(t)
	g.Night().Guard("Dung").Protects("Em").Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc").Seer("Em").Checks("Chi")
	g.Dawn().ExpectDead("Phuc")
	g.Day().NobodyHanged()
	g.Night().Guard("Dung").Protects("Giang").TurnOf(Seer)

	dead := g.As("Phuc").Full().Pack("An", "Binh", "Chi").SeerResults(1)
	if got := dead.View().Secrets.Pending.Protected; got != "Giang" {
		t.Fatalf("the dead should see tonight's protection, see %q", got)
	}
	if dead.View().NightCall == nil {
		t.Fatal("the dead should see the night calls")
	}
}
