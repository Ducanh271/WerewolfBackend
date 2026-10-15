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
	g := newTable(t)
	n := g.Night()
	g.Inject(cursed("Em"))
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
		{"visited tonight", func(g *Game, n *Night) {
			g.HandBuild(func(s *engine.State) { s.Pending.Visited = "Lan" })
			bitePhuc(n).TurnOf(Witch)
		}, ""},
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

func TestView_CursedPlayer_IsNotified(t *testing.T) {
	g := newTable(t)
	g.Night()
	g.As("Phuc").Cursed(false)
	g.Inject(cursed("Phuc"))
	g.As("Phuc").Cursed(true).Event(cursed("Phuc"))
	g.As("An").Cursed(false).NoEvent(cursed("Phuc"))
	g.As("Hoa").NoEvent(cursed("Phuc"))
}

func TestView_Lovers_SeePartnerRoleAndLinkColour_CupidSeesNothing(t *testing.T) {
	tests := []struct {
		name   string
		a, b   string
		curse  string // cursed after the pairing, if any
		colour engine.LinkColour
		roleA  engine.RoleID
		roleB  engine.RoleID
	}{
		{"two villagers: white", "Em", "Phuc", "", engine.LinkWhite, Seer, Villager},
		{"two wolves: black", "An", "Binh", "", engine.LinkBlack, Werewolf, Werewolf},
		{"a wolf and a villager: striped", "Em", "An", "", engine.LinkStriped, Seer, Werewolf},
		{"one lover cursed: the colour changes", "Em", "Phuc", "Phuc", engine.LinkStriped, Seer, Villager},
		{"both cursed in turn: black", "Em", "An", "Em", engine.LinkBlack, Seer, Werewolf},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			g.Night()
			g.HandBuild(func(s *engine.State) {
				s.Lovers = &engine.LoversLink{A: engine.PlayerID(tc.a), B: engine.PlayerID(tc.b)}
				s.Players["Lan"].Role = Cupid
			})
			if tc.curse != "" {
				g.Inject(cursed(tc.curse))
			}
			g.As(tc.a).Lover(tc.b, tc.roleB, tc.colour).RoleOf(tc.b, tc.roleB)
			g.As(tc.b).Lover(tc.a, tc.roleA, tc.colour).RoleOf(tc.a, tc.roleA)
			g.As("Lan").OwnRole(Cupid).NoLover().NoRoleExceptOwn()
			g.As("Hoa").NoLover().NoRoleExceptOwn()
		})
	}
}

func TestView_VisitedPlayer_SeesTheNoticeOnTheirTurn(t *testing.T) {
	tests := []struct {
		name    string
		visited string
		step    engine.RoleID
		viewer  string
		turn    bool
		notice  bool
	}{
		{"on her turn", "Em", Seer, "Em", true, true},
		{"before her turn", "Em", Guard, "Em", false, false},
		{"another player on their turn", "Em", Guard, "Dung", true, false},
		{"a visited wolf at the pack step", "An", Werewolf, "An", true, true},
		{"the other wolves are not blocked", "An", Werewolf, "Binh", true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			n := g.Night()
			g.HandBuild(func(s *engine.State) { s.Pending.Visited = engine.PlayerID(tc.visited) })
			n.TurnOf(tc.step)
			seen := g.As(tc.viewer).Visited(tc.notice)
			if tc.turn {
				seen.Turn(tc.step)
			} else {
				seen.NoTurn()
			}
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
