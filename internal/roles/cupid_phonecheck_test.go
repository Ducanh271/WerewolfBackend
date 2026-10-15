package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

// nightCalls plays a whole night by timing out every step and returns the
// steps that were called, in order.
func nightCalls(g *Game) []engine.RoleID {
	g.Night()
	var calls []engine.RoleID
	for {
		step, ok := engine.CurrentNightRole(g.State())
		if !ok {
			break
		}
		calls = append(calls, step)
		g.TimeOut()
	}
	g.Dawn()
	return calls
}

func countOf(calls []engine.RoleID, step engine.RoleID) int {
	n := 0
	for _, c := range calls {
		if c == step {
			n++
		}
	}
	return n
}

// Rules §3: the phone check runs on night 1 when Cupid OR the Cursed wolf is in
// the game, once, and on nights 2-3 only for the Cursed wolf.
func TestPhoneCheck_Night1RunsWithCupidOrCursedWolf_LaterNightsOnlyWithTheCursedWolf(t *testing.T) {
	tests := []struct {
		name   string
		swap   map[string]engine.RoleID
		nights [3]int // phone checks on nights 1, 2, 3
	}{
		{"Cupid and Cursed wolf", nil, [3]int{1, 1, 1}},
		{"Cupid only", map[string]engine.RoleID{"An": Werewolf}, [3]int{1, 0, 0}},
		{"Cursed wolf only", map[string]engine.RoleID{"Lan": Villager}, [3]int{1, 1, 1}},
		{"neither", map[string]engine.RoleID{"An": Werewolf, "Lan": Villager}, [3]int{0, 0, 0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := cupidTableWith(t, tc.swap)
			for night, want := range tc.nights {
				calls := nightCalls(g)
				if got := countOf(calls, engine.StepPhoneCheck); got != want {
					t.Fatalf("night %d: %d phone checks in %v, want %d", night+1, got, calls, want)
				}
				if night < 2 {
					g.Day().NobodyHanged()
				}
			}
		})
	}
}

// Rules §3: every living player gets ONE notice of the same shape, combining
// the curse status and the love status; only the values differ.
func TestPhoneCheck_EveryLivingPlayerGetsOneNoticeOfTheSameShape_OnlyValuesDiffer(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Em", "Chi").CursedWolf("An").Curses("Em").PhoneCheck()

	notices := map[engine.PlayerID]engine.PhoneChecked{}
	for _, e := range g.Events() {
		if pc, ok := e.(engine.PhoneChecked); ok {
			if _, twice := notices[pc.Player]; twice {
				t.Fatalf("%s got two notices", pc.Player)
			}
			notices[pc.Player] = pc
		}
	}
	if len(notices) != len(all) {
		t.Fatalf("%d notices for %d living players", len(notices), len(all))
	}
	for name, want := range map[string]engine.PhoneChecked{
		"Em":   {Cursed: true, Love: engine.LoveNotice{InLove: true, Partner: "Chi", PartnerRole: Villager, Colour: engine.LinkStriped}},
		"Chi":  {Cursed: false, Love: engine.LoveNotice{InLove: true, Partner: "Em", PartnerRole: Seer, Colour: engine.LinkStriped}},
		"Hoa":  {},
		"An":   {},
		"Lan":  {}, // Cupid is told nothing of the pair
		"Binh": {},
	} {
		want.Player = g.Player(name)
		if got := notices[g.Player(name)]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s got %+v, want %+v", name, got, want)
		}
	}
	// Each phone sees its own notice in the same type and fields.
	for _, name := range all {
		view := g.As(name).View().Me.PhoneCheck
		if view == nil {
			t.Fatalf("%s sees no phone check notice", name)
		}
	}
	g.As("Em").PhoneNotice(true).LoveNotice(engine.LoveNotice{InLove: true, Partner: "Chi", PartnerRole: Villager, Colour: engine.LinkStriped})
	g.As("Hoa").PhoneNotice(false).LoveNotice(engine.LoveNotice{})
	g.As("Lan").PhoneNotice(false).LoveNotice(engine.LoveNotice{})
}

// Rules §7: the pack does not see that one of its wolves is in love, and the
// lovers' notices go to the lovers alone.
func TestPhoneCheck_PackDoesNotSeeThatAWolfIsInLove(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Em", "Binh").PhoneCheck()
	g.As("Binh").Lover("Em", Seer, engine.LinkStriped)
	g.As("An").NoLover().PhoneNoticesInFeed(1).NoSecretsOfOthers()
	g.As("Hoa").NoLover().PhoneNoticesInFeed(1).NoSecretsOfOthers()
}
