package roles_test

import (
	"reflect"
	"sort"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"

	. "masoi/internal/engine/enginetest"
)

// cupidTable is the usual ten-player table of the Cursed wolf with Cupid in
// Lan's seat: An (Cursed wolf), Binh (wolf), Chi (villager), Dung (Guard), Em
// (Seer), Phuc (Hunter), Giang (Witch), Hoa (Chief), Khoa (villager), Lan
// (Cupid).
func cupidTable(t testing.TB) *Game { return cupidTableWith(t, nil) }

// cupidTableWith is the same table with some seats given other roles. The
// caller keeps the role list legal (one Cupid, one Cursed wolf...).
func cupidTableWith(t testing.TB, override map[string]engine.RoleID) *Game {
	roles := map[string]engine.RoleID{"An": CursedWolf, "Binh": Werewolf, "Chi": Villager, "Dung": Guard, "Em": Seer,
		"Phuc": Hunter, "Giang": Witch, "Hoa": Villager, "Khoa": Villager, "Lan": Cupid}
	for who, role := range override {
		roles[who] = role
	}
	var cast []any
	for _, name := range all {
		cast = append(cast, name, roles[name])
	}
	return NewGame(t, Seats(all...), Chief("Hoa"), Cast(cast...))
}

func TestCupid_OnlyInGamesOfTenPlayersOrMore(t *testing.T) {
	rule, ok := engine.Registry[Cupid].(engine.MinPlayersRule)
	if !ok || rule.MinPlayers() != 10 {
		t.Fatalf("Cupid must need at least 10 players, got %v", engine.Registry[Cupid])
	}
}

func TestCupid_CalledOnNightOneOnly_FirstOfTheNight(t *testing.T) {
	g := cupidTable(t)
	g.Night()
	if first, _ := engine.CurrentNightRole(g.State()); first != Cupid {
		t.Fatalf("the first call of night 1 is %s, want Cupid", first)
	}
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night()
	if first, _ := engine.CurrentNightRole(g.State()); first == Cupid {
		t.Fatal("Cupid is called again on night 2")
	}
}

func TestCupid_PicksTwoDifferentLivingPlayers_HimselfAllowed(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want error
	}{
		{"two others", "Em", "Chi", nil},
		{"himself and another", "Lan", "An", nil},
		{"the same player twice", "Em", "Em", engine.ErrInvalidTarget},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n := cupidTable(t).Night()
			if tc.want != nil {
				n.Cupid("Lan").CannotPair(tc.a, tc.b, tc.want)
				return
			}
			n.Cupid("Lan").Pairs(tc.a, tc.b)
			a, b, ok := n.Game().Lovers()
			if !ok || !sameSet([]string{a, b}, []string{tc.a, tc.b}) {
				t.Fatalf("the couple is %s and %s, want %s and %s", a, b, tc.a, tc.b)
			}
		})
	}
}

func TestCupid_OnlyCupidChoosesOnlyOnce_AndOnlyOnHisStep(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Em", "Chi")
	n.Cupid("Lan").CannotPair("Em", "Dung", engine.ErrAlreadyActed)

	g = cupidTable(t)
	g.Night()
	g.DoExpectErr(roles.CupidPair{By: g.Player("Em"), A: g.Player("Em"), B: g.Player("Chi")}, engine.ErrNotYourTurn)
	g.Dawn()
	g.Day().NobodyHanged()
	g.Night()
	g.DoExpectErr(roles.CupidPair{By: g.Player("Lan"), A: g.Player("Em"), B: g.Player("Chi")}, engine.ErrNotYourTurn)
}

func TestCupid_StepRunsOutWithoutAChoice_ServerPairsTwoRandomLivingPlayers(t *testing.T) {
	pairs := map[string]bool{}
	for seed := int64(1); seed <= 20; seed++ {
		g := cupidTable(t)
		g.Night().Cupid("Lan").RunsOutOfTime(seed)
		a, b, ok := g.Lovers()
		if !ok || a == b {
			t.Fatalf("seed %d: the server paired %q and %q", seed, a, b)
		}
		pairs[a+"/"+b] = true

		again := cupidTable(t)
		again.Night().Cupid("Lan").RunsOutOfTime(seed)
		if a2, b2, _ := again.Lovers(); a2 != a || b2 != b {
			t.Fatalf("seed %d is not deterministic: %s/%s then %s/%s", seed, a, b, a2, b2)
		}
	}
	if len(pairs) < 5 {
		t.Fatalf("20 seeds gave only %d different couples", len(pairs))
	}
}

func sameSet(a, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	return reflect.DeepEqual(a, b)
}
