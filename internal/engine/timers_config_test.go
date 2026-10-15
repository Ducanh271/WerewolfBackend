package engine

import (
	"errors"
	"testing"
	"time"
)

func TestTimers_DefaultsAreInsideTheirLimits(t *testing.T) {
	set := DefaultSettings()
	for _, f := range (&Timers{}).fields(&set) {
		if *f.slot < f.limit.min || *f.slot > f.limit.max {
			t.Errorf("default %s = %s is outside %s..%s", f.name, *f.slot, f.limit.min, f.limit.max)
		}
	}
}

func TestTimers_ConfigureChangesOnlyWhatIsAsked(t *testing.T) {
	s := apply(t, lobbyOf(t, 5), ConfigureTimers{Timers: Timers{Day: 3 * time.Minute, NightStep: 20 * time.Second}})
	want := DefaultSettings()
	want.Day, want.NightStep = 3*time.Minute, 20*time.Second
	if s.Settings.Day != want.Day || s.Settings.NightStep != want.NightStep || s.Settings.Defense != want.Defense ||
		s.Settings.Announce != want.Announce || s.Settings.RoleView != want.RoleView {
		t.Fatalf("settings = %+v", s.Settings)
	}
}

func TestTimers_ConfigureRejectsOutOfRange(t *testing.T) {
	tests := []struct {
		name string
		t    Timers
	}{
		{"night step too short", Timers{NightStep: 4 * time.Second}},
		{"phone check too short", Timers{PhoneCheck: 4 * time.Second}},
		{"phone check too long", Timers{PhoneCheck: 11 * time.Minute}},
		{"defense too long", Timers{Defense: 11 * time.Minute}},
		{"day too short", Timers{Day: 30 * time.Second}},
		{"day too long", Timers{Day: time.Hour}},
		{"countdown too long", Timers{NominationLock: 2 * time.Minute}},
		{"announce pause too long", Timers{Announce: 2 * time.Minute}},
		{"negative", Timers{Verdict: -time.Second}},
		{"one bad value spoils the rest", Timers{Day: 5 * time.Minute, HunterShot: time.Second}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lobby := lobbyOf(t, 5)
			_, err := Decide(lobby, ConfigureTimers{Timers: tc.t})
			if !errors.Is(err, ErrTimerOutOfRange) {
				t.Fatalf("err = %v, want ErrTimerOutOfRange", err)
			}
		})
	}
}

func TestTimers_ConfigureRejectedAfterTheDeal(t *testing.T) {
	dealt := apply(t, tableOf(t, 6, 1, false), DealRoles{Seed: 1})
	_, err := Decide(dealt, ConfigureTimers{Timers: Timers{Day: 5 * time.Minute}})
	if !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("err = %v, want ErrWrongPhase", err)
	}
}

func TestTimers_PhoneCheckDefaultsToFiveSeconds_AndCanBeChanged(t *testing.T) {
	if got := DefaultSettings().PhoneCheck; got != 5*time.Second {
		t.Fatalf("default = %s, want 5s", got)
	}
	s := apply(t, lobbyOf(t, 5), ConfigureTimers{Timers: Timers{PhoneCheck: 8 * time.Second}})
	if s.Settings.PhoneCheck != 8*time.Second {
		t.Fatalf("phone check = %s, want 8s", s.Settings.PhoneCheck)
	}
}
