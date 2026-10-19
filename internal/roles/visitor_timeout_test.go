package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

// When the Visitor runs out of time the server visits for her (rules §6
// "Nguyệt Nữ"): a random valid player, never herself nor last night's player.

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
