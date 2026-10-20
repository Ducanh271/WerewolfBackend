package engine

import (
	"testing"
	"time"
)

// The lobby and the end screen carry the current timers and sound mode, so a
// phone that comes back does not rebuild them from the defaults and events.
func TestView_LobbyAndEnded_CarryTimersAndSound(t *testing.T) {
	s := lobbyOf(t, 5)
	s = apply(t, s, ConfigureTimers{Timers: Timers{Day: 7 * time.Minute}})
	s = apply(t, s, ConfigureSound{Mode: AllPhonesSpeak})
	defaults := DefaultSettings()

	ended := Fold(s, GameEnded{Outcome: Outcome{Draw: true}})
	for name, st := range map[string]State{"lobby": s, "ended": ended} {
		v := ViewFor(st, nil, Viewer{Player: playerID(0)})
		if v.Timers == nil {
			t.Fatalf("%s: no timers", name)
		}
		if v.Timers.Day != 7*time.Minute {
			t.Errorf("%s: day = %s, want the 7m the host set", name, v.Timers.Day)
		}
		if v.Timers.NightStep != defaults.NightStep || v.Timers.ChiefSuccession != defaults.ChiefSuccession {
			t.Errorf("%s: timers = %+v, want the defaults for the times not changed", name, *v.Timers)
		}
		if v.Sound == nil || *v.Sound != AllPhonesSpeak {
			t.Errorf("%s: sound = %v, want AllPhonesSpeak", name, v.Sound)
		}
	}
}

func TestView_DuringTheGame_NoTimersOrSound(t *testing.T) {
	s := lobbyOf(t, 5)
	s.Phase = PhaseNight
	v := ViewFor(s, nil, Viewer{Player: playerID(0)})
	if v.Timers != nil || v.Sound != nil {
		t.Errorf("timers/sound = %+v/%v, want none during the game", v.Timers, v.Sound)
	}
}
