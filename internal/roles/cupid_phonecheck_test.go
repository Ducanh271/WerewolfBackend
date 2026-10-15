package roles_test

import (
	"reflect"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/i18n"

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

const (
	phoneCheck = engine.StepPhoneCheck
	loversMeet = engine.StepLoversMeet
)

// Rules §3: on night 1 with Cupid, Cupid picks, everybody looks at their phone
// (love notice), then the lovers open their eyes and look at each other; the
// Visitor and the rest follow. With the Cursed wolf, its own phone check stays
// right after its step, on nights 1-3. Both run on a fixed schedule.
func TestNightOrder_CupidThenPhoneCheckThenLovers_CursedWolfKeepsItsOwnCheck(t *testing.T) {
	visitor := map[string]engine.RoleID{"Khoa": Visitor}
	tests := []struct {
		name   string
		swap   map[string]engine.RoleID
		nights [4][]engine.RoleID
	}{
		{"Cupid and Cursed wolf", nil, [4][]engine.RoleID{
			{Cupid, phoneCheck, loversMeet, Visitor, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer},
			{Visitor, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer},
			{Visitor, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
		}},
		{"Cupid only", map[string]engine.RoleID{"An": Werewolf}, [4][]engine.RoleID{
			{Cupid, phoneCheck, loversMeet, Visitor, Guard, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
		}},
		{"Cursed wolf only", map[string]engine.RoleID{"Lan": Villager}, [4][]engine.RoleID{
			{Visitor, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer},
			{Visitor, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer},
			{Visitor, Guard, CursedWolf, phoneCheck, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
		}},
		{"neither", map[string]engine.RoleID{"An": Werewolf, "Lan": Villager}, [4][]engine.RoleID{
			{Visitor, Guard, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
			{Visitor, Guard, Werewolf, Witch, Seer},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			swap := map[string]engine.RoleID{}
			for k, v := range visitor {
				swap[k] = v
			}
			for k, v := range tc.swap {
				swap[k] = v
			}
			g := cupidTableWith(t, swap)
			for night, want := range tc.nights {
				if got := nightCalls(g); !reflect.DeepEqual(got, want) {
					t.Fatalf("night %d calls %v, want %v", night+1, got, want)
				}
				if g.State().Phase == engine.PhaseEnded {
					t.Fatal("the game ended before night 4")
				}
				g.Day().NobodyHanged()
			}
		})
	}
}

// noticesPerCheck groups the phone check notices of the game by check.
func noticesPerCheck(t testing.TB, g *Game) []map[engine.PlayerID]engine.PhoneChecked {
	var out []map[engine.PlayerID]engine.PhoneChecked
	for _, e := range g.Events() {
		switch ev := e.(type) {
		case engine.PhoneCheckAnnounced:
			out = append(out, map[engine.PlayerID]engine.PhoneChecked{})
		case engine.PhoneChecked:
			cur := out[len(out)-1]
			if _, twice := cur[ev.Player]; twice {
				t.Fatalf("%s got two notices in one check", ev.Player)
			}
			cur[ev.Player] = ev
		}
	}
	return out
}

// Rules §3: at each check every living player gets ONE notice of the same
// shape, combining the curse status and the love status; only the values
// differ. The check after Cupid comes before the curse; the one after the
// curse repeats the love part with the new colour.
func TestPhoneCheck_EachCheckGivesEveryLivingPlayerOneNoticeOfTheSameShape(t *testing.T) {
	g := cupidTable(t)
	n := g.Night().Cupid("Lan").Pairs("Em", "Chi").PhoneCheck()
	g.As("Em").PhoneNotice(false).LoveNotice(engine.LoveNotice{InLove: true, Partner: "Chi", PartnerRole: Villager, Colour: engine.LinkWhite})
	g.As("Hoa").PhoneNotice(false).LoveNotice(engine.LoveNotice{})
	n.CursedWolf("An").Curses("Em").PhoneCheck()

	checks := noticesPerCheck(t, g)
	if len(checks) != 2 {
		t.Fatalf("%d phone checks on night 1, want 2", len(checks))
	}
	inLove := func(partner string, role engine.RoleID, c engine.LinkColour) engine.LoveNotice {
		return engine.LoveNotice{InLove: true, Partner: engine.PlayerID(partner), PartnerRole: role, Colour: c}
	}
	wants := []map[string]engine.PhoneChecked{
		{ // after Cupid: nobody is cursed yet
			"Em":  {Love: inLove("Chi", Villager, engine.LinkWhite)},
			"Chi": {Love: inLove("Em", Seer, engine.LinkWhite)},
			"Hoa": {}, "An": {}, "Lan": {}, "Binh": {},
		},
		{ // after the curse
			"Em":  {Cursed: true, Love: inLove("Chi", Villager, engine.LinkStriped)},
			"Chi": {Love: inLove("Em", Seer, engine.LinkStriped)},
			"Hoa": {}, "An": {}, "Lan": {}, "Binh": {}, // Cupid is told nothing of the pair
		},
	}
	for i, check := range checks {
		if len(check) != len(all) {
			t.Fatalf("check %d: %d notices for %d living players", i+1, len(check), len(all))
		}
		for name, want := range wants[i] {
			want.Player = g.Player(name)
			if got := check[g.Player(name)]; !reflect.DeepEqual(got, want) {
				t.Errorf("check %d: %s got %+v, want %+v", i+1, name, got, want)
			}
		}
	}
	for _, name := range all {
		if g.As(name).View().Me.PhoneCheck == nil {
			t.Fatalf("%s sees no phone check notice", name)
		}
	}
	g.As("Em").PhoneNotice(true).LoveNotice(inLove("Chi", Villager, engine.LinkStriped))
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

// Rules §3: the lovers' step lasts 10 seconds by default and opens with its
// voice line. Every phone sees the same call; only the two lovers are called,
// and nobody acts in it.
func TestLoversMeet_TenSeconds_OnlyTheLoversAreCalled(t *testing.T) {
	g := cupidTable(t)
	g.Night().Cupid("Lan").Pairs("Em", "Binh").LoversMeet()
	if got := g.State().Settings.LoversMeet; got != 10*time.Second {
		t.Fatalf("the lovers' step lasts %v by default, want 10s", got)
	}
	for _, name := range all {
		seen := g.As(name).NightCall(loversMeet, g.State().Settings.LoversMeet)
		if seen.View().Me.PhoneCheck != nil {
			t.Fatalf("%s still sees the phone check notice", name)
		}
		switch name {
		case "Em", "Binh":
			seen.Turn(loversMeet)
		default:
			seen.NoTurn().NoLover()
		}
	}
	g.As("Em").Lover("Binh", Werewolf, engine.LinkStriped)
	announced := 0
	for _, e := range g.Events() {
		if a, ok := e.(engine.LoversMeetAnnounced); ok {
			announced++
			if a.AnnouncementKey() != i18n.NightLoversMeet {
				t.Fatalf("voice line key %q, want %q", a.AnnouncementKey(), i18n.NightLoversMeet)
			}
		}
	}
	if announced != 1 {
		t.Fatalf("the voice line was announced %d times, want once", announced)
	}
	g.As("Hoa").NoSecretsOfOthers()
}
