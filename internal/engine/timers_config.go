package engine

import (
	"fmt"
	"time"
)

// Timers are the phase times the host may change (rules §2). A zero field
// means "leave as it is"; anything else must lie within its limits.
type Timers struct {
	RoleView        time.Duration
	ChiefElection   time.Duration
	NightStep       time.Duration
	PhoneCheck      time.Duration
	Announce        time.Duration
	Day             time.Duration
	NominationLock  time.Duration
	Defense         time.Duration
	Verdict         time.Duration
	HunterShot      time.Duration
	ChiefSuccession time.Duration
}

// ConfigureTimers changes the phase times. Like the other setup commands it
// works only in the lobby; the room lets only the host send it.
type ConfigureTimers struct {
	CommandMarker
	Timers Timers
}

// TimersConfigured records the changed times.
type TimersConfigured struct {
	EventMarker
	Timers Timers
}

// timerLimit is the allowed range of one phase time.
type timerLimit struct{ min, max time.Duration }

var (
	shortTimer = timerLimit{5 * time.Second, 10 * time.Minute}
	dayTimer   = timerLimit{time.Minute, 30 * time.Minute}
	pauseTimer = timerLimit{5 * time.Second, time.Minute}
)

// fields lists every timer with its name, limits and slots, so validation and
// application share one table.
func (t *Timers) fields(set *Settings) []timerField {
	return []timerField{
		{"RoleView", shortTimer, &t.RoleView, &set.RoleView},
		{"ChiefElection", shortTimer, &t.ChiefElection, &set.ChiefElection},
		{"NightStep", shortTimer, &t.NightStep, &set.NightStep},
		{"PhoneCheck", shortTimer, &t.PhoneCheck, &set.PhoneCheck},
		{"Announce", pauseTimer, &t.Announce, &set.Announce},
		{"Day", dayTimer, &t.Day, &set.Day},
		{"NominationLock", pauseTimer, &t.NominationLock, &set.NominationLock},
		{"Defense", shortTimer, &t.Defense, &set.Defense},
		{"Verdict", shortTimer, &t.Verdict, &set.Verdict},
		{"HunterShot", shortTimer, &t.HunterShot, &set.HunterShot},
		{"ChiefSuccession", shortTimer, &t.ChiefSuccession, &set.ChiefSuccession},
	}
}

type timerField struct {
	name  string
	limit timerLimit
	asked *time.Duration
	slot  *time.Duration
}

func decideConfigureTimers(s State, cmd ConfigureTimers) ([]Event, error) {
	if s.Phase != PhaseLobby || s.Dealt {
		return nil, fmt.Errorf("%w: cannot change timers during %s", ErrWrongPhase, s.Phase)
	}
	t, set := cmd.Timers, s.Settings
	for _, f := range t.fields(&set) {
		d := *f.asked
		if d != 0 && (d < f.limit.min || d > f.limit.max) {
			return nil, fmt.Errorf("%w: %s is %s, allowed %s to %s", ErrTimerOutOfRange, f.name, d, f.limit.min, f.limit.max)
		}
	}
	return []Event{TimersConfigured{Timers: t}}, nil
}

// applyTimers copies the non-zero times into the settings.
func applyTimers(set Settings, t Timers) Settings {
	for _, f := range t.fields(&set) {
		if *f.asked != 0 {
			*f.slot = *f.asked
		}
	}
	return set
}
