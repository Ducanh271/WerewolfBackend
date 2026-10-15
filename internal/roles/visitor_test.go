package roles_test

import (
	"fmt"
	"strings"
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/i18n"
	"masoi/internal/roles"
)

// visitorTable is the ten-player table with the Visitor in Khoa's seat: three
// wolves (An, Binh, Chi), the Guard (Dung), the Seer (Em), the Hunter (Phuc),
// the Witch (Giang), the Visitor (Khoa) and two villagers (Hoa, Lan).
func visitorTable(t testing.TB) *Game {
	return NewGame(t,
		Seats(all...),
		Cast("An", Werewolf, "Binh", Werewolf, "Chi", Werewolf, "Dung", Guard, "Em", Seer,
			"Phuc", Hunter, "Giang", Witch, "Hoa", Villager, "Khoa", Visitor, "Lan", Villager))
}

// blockedNow lists who is blocked at this moment, by seat order.
func blockedNow(g *Game) []string {
	var out []string
	for _, id := range g.State().Seats {
		if engine.Blocked(g.State(), id) {
			out = append(out, string(id))
		}
	}
	return out
}

func TestVisitor_IsCalledFirstEveryNight(t *testing.T) {
	g := visitorTable(t)
	n := g.Night()
	g.As("Khoa").Turn(Visitor)
	g.As("Dung").NoTurn()
	n.Visitor("Khoa").Visits("Hoa")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night()
	g.As("Khoa").Turn(Visitor)
}

func TestVisitor_CannotVisitHerself(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").CannotVisit("Khoa", engine.ErrInvalidTarget)
}

func TestVisitor_CannotVisitTheSamePlayerTwoNightsInARow(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("Hoa")
	g.Dawn()
	g.Day().NobodyHanged()
	n := g.Night()
	n.Visitor("Khoa").CannotVisit("Hoa", engine.ErrSameVisitorTarget)
	n.Visitor("Khoa").Visits("Lan")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Visitor("Khoa").Visits("Hoa") // two nights ago is fine
}

func TestVisitor_CannotVisitTheDead(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("Lan").Wolf("An").Bites("Hoa").Wolf("Binh").Bites("Hoa")
	g.Dawn().ExpectDead("Hoa")
	g.Day().NobodyHanged()
	g.Night().Visitor("Khoa").CannotVisit("Hoa", engine.ErrInvalidTarget)
}

func TestVisitor_VisitedPlayerSeesAGenericNoticeOnTheirTurn(t *testing.T) {
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
		{"a blocked wolf at the pack step", "An", Werewolf, "An", true, true},
		{"the other wolves are not blocked", "An", Werewolf, "Binh", true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := visitorTable(t)
			g.Night().Visitor("Khoa").Visits(tc.visited).TurnOf(tc.step)
			seen := g.As(tc.viewer).Blocked(tc.notice)
			if tc.turn {
				seen.Turn(tc.step)
			} else {
				seen.NoTurn()
			}
		})
	}
}

// The blocked player must never learn who blocked them (rules §3). The public
// role list and the "call the Visitor" cue are public to everybody, so what is
// checked is everything private to the player: their own part of the view, the
// secrets and the events they may see.
func TestVisitor_ViewOfTheBlockedPlayer_NeverNamesTheBlocker(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Em")
	n.Wolf("An").Bites("Phuc").TurnOf(Seer)
	for _, name := range []string{"Em", "Dung", "An"} {
		seen := g.As(name)
		v := seen.View()
		private := fmt.Sprintf("%+v %+v %+v", v.Me, v.Secrets, v.Pack)
		for _, bad := range []string{"Nguyệt Nữ", string(Visitor), "Khoa", "Visited", "Visitor"} {
			if strings.Contains(private, bad) {
				t.Errorf("%s's private view mentions %q: %s", name, bad, private)
			}
		}
		for _, e := range seen.Feed() {
			if _, leak := e.(roles.PlayerVisited); leak {
				t.Errorf("%s may see the visit event", name)
			}
		}
	}
	g.As("Khoa").Feed() // the Visitor herself may know
	if text := i18n.Vi(i18n.NightBlocked); strings.Contains(text, "Nguyệt Nữ") || text != "Bạn không thể sử dụng chức năng đêm nay." {
		t.Errorf("blocked notice = %q", text)
	}
}

func TestVisitor_OnlyTheVisitorKnowsWhomSheVisited(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("Em")
	visits := func(name string) int {
		count := 0
		for _, e := range g.As(name).Feed() {
			if _, ok := e.(roles.PlayerVisited); ok {
				count++
			}
		}
		return count
	}
	if visits("Khoa") != 1 || visits("Em") != 0 || visits("An") != 0 {
		t.Fatalf("visit event seen by Khoa=%d Em=%d An=%d, want 1 0 0", visits("Khoa"), visits("Em"), visits("An"))
	}
}

func TestVisitor_GuardVisited_ProtectsNobody(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Dung")
	n.Guard("Dung").CannotProtect("Phuc", engine.ErrBlocked)
	n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc").ExpectAnnouncement(i18n.HunterRevealed)
}

func TestVisitor_SeerVisited_GetsNoResult(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Em")
	n.Seer("Em").CannotCheck("An", engine.ErrBlocked)
	g.As("Em").SeerResults(0)
}

func TestVisitor_WitchVisited_SeesNothingAndCannotAct(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Giang")
	packBites(n, "Hoa").TurnOf(Witch)
	g.As("Giang").Blocked(true).Bitten("").BiteShown(false)
	n.Witch("Giang").CannotHeal(engine.ErrBlocked).Witch("Giang").CannotPoison("Lan", engine.ErrBlocked)
	g.Dawn().ExpectDead("Hoa")
}

func TestVisitor_HunterVisited_StillShootsByDay(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Phuc")
	packBites(n, "Phuc")
	g.Dawn().ExpectDead("Phuc")
	g.Hunter("Phuc").Shoots("An")
	g.ExpectDead("Phuc", "An")
}

func TestVisitor_WolfVisited_CannotVoteAndThePackMajorityCountsTheOthers(t *testing.T) {
	tests := []struct {
		name  string
		votes map[string]string // wolf -> bitten player
		dead  string
	}{
		{"the two free wolves agree: 2 of 2", map[string]string{"Binh": "Hoa", "Chi": "Hoa"}, "Hoa"},
		{"the two free wolves split: nobody is bitten", map[string]string{"Binh": "Hoa", "Chi": "Lan"}, ""},
		{"one free wolf votes alone: 1 of 2 is not over half", map[string]string{"Binh": "Hoa"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := visitorTable(t)
			n := g.Night().Visitor("Khoa").Visits("An")
			n.Wolf("An").CannotBite("Hoa", engine.ErrBlocked)
			for _, wolf := range []string{"Binh", "Chi"} {
				if target, ok := tc.votes[wolf]; ok {
					n.Wolf(wolf).Bites(target)
				}
			}
			d := g.Dawn()
			if tc.dead == "" {
				d.ExpectNobodyDied()
			} else {
				d.ExpectDead(tc.dead)
			}
		})
	}
}

func TestVisitor_TheOnlyWolfVisited_NobodyIsBitten(t *testing.T) {
	g := NewGame(t,
		Seats("An", "Dung", "Em", "Phuc", "Giang", "Hoa", "Khoa"),
		Cast("An", Werewolf, "Dung", Guard, "Em", Seer, "Phuc", Villager, "Giang", Villager,
			"Hoa", Villager, "Khoa", Visitor))
	n := g.Night().Visitor("Khoa").Visits("An")
	n.Wolf("An").CannotBite("Phuc", engine.ErrBlocked)
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
}

func TestVisitor_GuardVisitedOnNight2_CanProtectOnNight3WhoHeProtectedOnNight1(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("Hoa").Guard("Dung").Protects("Phuc")
	g.Dawn()
	g.Day().NobodyHanged()

	n := g.Night().Visitor("Khoa").Visits("Dung")
	n.Guard("Dung").CannotProtect("Lan", engine.ErrBlocked)
	g.Dawn()
	g.Day().NobodyHanged()

	g.Night().Visitor("Khoa").Visits("Hoa").Guard("Dung").Protects("Phuc") // nights 1 and 3 are not consecutive
}

func TestVisitor_GuardProtectedOnNight2_StillCannotRepeatOnNight3(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("Hoa").Guard("Dung").Protects("Phuc")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Visitor("Khoa").Visits("Lan").Guard("Dung").Protects("Em")
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night().Visitor("Khoa").Visits("Hoa").Guard("Dung").CannotProtect("Em", engine.ErrSameGuardTarget)
}

func TestVisitor_BittenWhileVisiting_DiesAndTheVisitStillTakesEffect(t *testing.T) {
	g := visitorTable(t)
	n := g.Night().Visitor("Khoa").Visits("Em")
	n.Wolf("An").Bites("Khoa").Wolf("Binh").Bites("Khoa")
	n.Seer("Em").CannotCheck("An", engine.ErrBlocked)
	g.Dawn().ExpectDead("Khoa")
}

func TestVisitor_RunsOutOfTime_ServerVisitsARandomValidPlayer(t *testing.T) {
	for seed := int64(0); seed < 60; seed++ {
		g := visitorTable(t)
		g.Night().Visitor("Khoa").RunsOutOfTime(seed)
		got := blockedNow(g)
		if len(got) != 1 || got[0] == "Khoa" {
			t.Fatalf("seed %d: blocked %v, want exactly one player other than the Visitor", seed, got)
		}
	}
}

func TestVisitor_RunsOutOfTime_SameSeedSamePlayerAndSeedsSpreadOverThePlayers(t *testing.T) {
	pick := func(seed int64) string {
		g := visitorTable(t)
		g.Night().Visitor("Khoa").RunsOutOfTime(seed)
		return blockedNow(g)[0]
	}
	seen := map[string]bool{}
	for seed := int64(0); seed < 60; seed++ {
		first := pick(seed)
		if again := pick(seed); again != first {
			t.Fatalf("seed %d picked %s then %s", seed, first, again)
		}
		seen[first] = true
	}
	if len(seen) < 4 {
		t.Fatalf("60 seeds only ever picked %v: not random enough", seen)
	}
}

func TestVisitor_RunsOutOfTime_NeverPicksHerselfOrLastNightsPlayer(t *testing.T) {
	for seed := int64(0); seed < 60; seed++ {
		g := visitorTable(t)
		g.Night().Visitor("Khoa").Visits("Hoa")
		g.Dawn()
		g.Day().NobodyHanged()
		g.Night().Visitor("Khoa").RunsOutOfTime(seed)
		got := blockedNow(g)
		if len(got) != 1 || got[0] == "Khoa" || got[0] == "Hoa" {
			t.Fatalf("seed %d: night 2 blocked %v, want one player who is neither Khoa nor Hoa", seed, got)
		}
	}
}

func TestVisitor_RunsOutOfTime_NobodyValid_NobodyIsVisited(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("An")
	g.Dawn()
	g.Day().NobodyHanged()
	n := g.Night()
	g.HandBuild(func(s *engine.State) { // only Khoa and An are left, and An was visited last night
		for id, p := range s.Players {
			p.Alive = id == "Khoa" || id == "An"
		}
	})
	n.Visitor("Khoa").RunsOutOfTime(7)
	if got := blockedNow(g); len(got) != 0 {
		t.Fatalf("blocked %v, want nobody", got)
	}
}

func TestVisitor_Dead_StillCalledButNobodyIsVisited(t *testing.T) {
	g := visitorTable(t)
	g.Night().Visitor("Khoa").Visits("Lan").Wolf("An").Bites("Khoa").Wolf("Binh").Bites("Khoa")
	g.Dawn().ExpectDead("Khoa")
	g.Day().NobodyHanged()
	n := g.Night()
	n.TurnOf(Visitor) // a fake call: the step is still there
	n.Visitor("Khoa").RunsOutOfTime(3)
	if got := blockedNow(g); len(got) != 0 {
		t.Fatalf("a dead Visitor blocked %v", got)
	}
}
