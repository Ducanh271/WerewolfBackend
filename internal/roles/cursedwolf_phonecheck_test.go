package roles_test

import (
	"reflect"
	"testing"
	"time"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/i18n"
)

// noticesOfNight maps each living player who got a phone check notice on the
// given night to what it said.
func noticesOfNight(g *Game, night int) map[engine.PlayerID]bool {
	out := map[engine.PlayerID]bool{}
	in := false
	for _, e := range g.Events() {
		switch ev := e.(type) {
		case engine.NightStarted:
			in = ev.Night == night
		case engine.PhoneChecked:
			if in {
				out[ev.Player] = ev.Cursed
			}
		}
	}
	return out
}

func TestPhoneCheck_AlwaysRunsOnNightsOneToThree_EveryLivingPlayerGetsANotice(t *testing.T) {
	tests := []struct {
		name  string
		table func(t testing.TB) *Game
		setup func(g *Game)
		night func(g *Game) // what happens on the night under test
		want  map[string]bool
		dead  []string
	}{
		{"nobody is cursed", cursedTable, nil, func(g *Game) { g.Night() }, map[string]bool{}, nil},
		{"the curse is cast that night", cursedTable, nil,
			func(g *Game) { g.Night().CursedWolf("An").Curses("Phuc") }, map[string]bool{"Phuc": true}, nil},
		{"the curse was cast on an earlier night", cursedTable, func(g *Game) {
			g.Night().CursedWolf("An").Curses("Phuc")
			passNight(g)
		}, func(g *Game) { g.Night() }, map[string]bool{"Phuc": true}, nil},
		{"the Cursed wolf is dead", cursedTable, func(g *Game) {
			g.HandBuild(func(s *engine.State) { s.Players["An"].Alive = false })
		}, func(g *Game) { g.Night() }, map[string]bool{}, []string{"An"}},
		{"the Cursed wolf is blocked", cursedVisitorTable, nil, func(g *Game) {
			g.Night().Visitor("Khoa").Visits("An").CursedWolf("An").CannotCurse("Phuc", engine.ErrBlocked)
		}, map[string]bool{}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := tc.table(t)
			if tc.setup != nil {
				tc.setup(g)
			}
			night := g.State().Night + 1
			if tc.setup == nil {
				night = 1
			}
			for _, name := range tc.dead {
				g.HandBuild(func(s *engine.State) { s.Players[engine.PlayerID(name)].Alive = false })
			}
			tc.night(g)
			g.Dawn()
			got := noticesOfNight(g, night)
			for _, id := range g.State().Seats {
				_, noticed := got[id]
				alive := !contains2(tc.dead, string(id))
				if noticed != alive {
					t.Errorf("%s: notice=%v, alive=%v", id, noticed, alive)
				}
				if got[id] != tc.want[string(id)] {
					t.Errorf("%s: cursed notice=%v, want %v", id, got[id], tc.want[string(id)])
				}
			}
		})
	}
}

func contains2(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// Rules §3: the two notices have the same shape, so nobody can tell from the
// screen who was cursed; only the one value differs.
func TestPhoneCheck_BothNoticesAreIndistinguishableExceptTheValue(t *testing.T) {
	g := cursedTable(t)
	n := g.Night().CursedWolf("An").Curses("Phuc").PhoneCheck()
	_ = n

	cursed := g.As("Phuc").PhoneNotice(true)
	plain := g.As("Hoa").PhoneNotice(false)
	g.As("An").PhoneNotice(false) // the Cursed wolf himself is not cursed

	a, b := cursed.View().Me.PhoneCheck, plain.View().Me.PhoneCheck
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		t.Fatalf("notice kinds differ: %T and %T", a, b)
	}
	b.Cursed = true
	if *a != *b {
		t.Fatalf("notices differ in more than the value: %+v and %+v", a, b)
	}

	feedNotice := func(s *Seen) engine.PhoneChecked {
		var found []engine.PhoneChecked
		for _, e := range s.Feed() {
			if pc, ok := e.(engine.PhoneChecked); ok {
				found = append(found, pc)
			}
		}
		if len(found) != 1 {
			t.Fatalf("a player must get exactly one notice, got %v", found)
		}
		return found[0]
	}
	x, y := feedNotice(g.As("Phuc")), feedNotice(g.As("Hoa"))
	if !x.Cursed || y.Cursed {
		t.Fatalf("notices say cursed=%v and cursed=%v, want true and false", x.Cursed, y.Cursed)
	}
	x.Player, y.Player, x.Cursed, y.Cursed = "", "", false, false
	if !reflect.DeepEqual(x, y) {
		t.Fatalf("notice events differ in more than the player and the value: %+v and %+v", x, y)
	}

	// The step itself is the same cue on every phone, and nobody is "called".
	for _, name := range all {
		g.As(name).NightCall(engine.StepPhoneCheck, g.State().Settings.PhoneCheck).NoTurn()
	}
}

func TestPhoneCheck_LastsItsOwnTime_AndTheSpeakerReadsTheVoiceLine(t *testing.T) {
	g := cursedTable(t)
	g.Night().PhoneCheck()
	g.As("Hoa").NightCall(engine.StepPhoneCheck, 5*time.Second)
	announced := 0
	for _, e := range g.Events() {
		if a, ok := e.(engine.PhoneCheckAnnounced); ok {
			announced++
			if a.AnnouncementKey() != i18n.NightPhoneCheck {
				t.Fatalf("voice line key %q, want %q", a.AnnouncementKey(), i18n.NightPhoneCheck)
			}
		}
	}
	if announced != 1 {
		t.Fatalf("the voice line was announced %d times, want once", announced)
	}
}
