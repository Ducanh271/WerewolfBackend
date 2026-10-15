package room

import (
	"time"

	"masoi/internal/engine"
)

type armKey struct {
	phase engine.Phase
	seq   int
}

func (r *Room) stopTimers() {
	if r.phaseT != nil {
		r.phaseT.Stop()
		r.phaseT = nil
	}
	if r.dayT != nil {
		r.dayT.Stop()
		r.dayT = nil
	}
}

// phaseDuration is how long the current phase lasts, if it is timed.
func (r *Room) phaseDuration() (time.Duration, bool) {
	set := r.state.Settings
	switch r.state.Phase {
	case engine.PhaseRoleView:
		return set.RoleView, true
	case engine.PhaseChiefElection:
		return set.ChiefElection, true
	case engine.PhaseChiefSuccession:
		return set.ChiefSuccession, true
	case engine.PhaseNight:
		return set.NightStep, true
	case engine.PhaseDawn, engine.PhaseDusk:
		return set.Announce, true
	case engine.PhaseDay:
		return r.day.remaining, true
	case engine.PhaseNominationLock:
		return set.NominationLock, true
	case engine.PhaseTrial:
		return set.Defense, true
	case engine.PhaseVerdict:
		return set.Verdict, true
	case engine.PhaseHunterShot:
		return set.HunterShot, true
	}
	return 0, false
}

// rearm runs after every change of state. Each phase or night step gets one
// timer, set for its full time and never cut short. The day clock pauses
// outside the day and the nomination countdown.
func (r *Room) rearm() {
	now := r.clock.Now()
	phase := r.state.Phase
	r.day.setRunning(phase == engine.PhaseDay || phase == engine.PhaseNominationLock)

	key := armKey{phase, r.state.Seq}
	if key == r.armedFor && (r.phaseT != nil || r.dayT != nil) {
		return
	}
	r.armedFor = key
	r.stopTimers()
	r.deadline = time.Time{}

	if d, ok := r.phaseDuration(); ok {
		r.deadline = now.Add(d)
		r.phaseT = r.clock.AfterFunc(d, func() { r.fire(engine.PhaseTimeout{Phase: key.phase, Seq: key.seq}) })
	}
	if phase == engine.PhaseNominationLock {
		// The 7 minutes keep running through the countdown.
		r.dayT = r.clock.AfterFunc(r.day.remaining, func() {
			r.fire(engine.PhaseTimeout{Phase: key.phase, Seq: key.seq, DayClock: true})
		})
	}
}

// fire turns a deadline into a PhaseTimeout command. The seed for an election
// that votes did not settle is drawn here, never taken from a client. A
// timeout whose phase or seq has moved on is stale and the engine ignores it.
func (r *Room) fire(t engine.PhaseTimeout) {
	r.post(func() {
		t.Seed = r.rnd.Seed()
		_ = r.execute(t)
	})
}
