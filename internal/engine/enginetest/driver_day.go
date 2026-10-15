package enginetest

import (
	"fmt"

	"masoi/internal/engine"
)

// dayRun works out the commands of a day step on a private copy of the state,
// because every vote and timeout changes it.
type dayRun struct {
	s    engine.State
	cmds []engine.Command
}

func (r *dayRun) play(c engine.Command) error {
	evs, err := engine.Decide(r.s, c)
	if err != nil {
		return err
	}
	r.cmds = append(r.cmds, c)
	r.s = engine.Fold(r.s, evs...)
	return nil
}

func (r *dayRun) timeout() error {
	return r.play(engine.PhaseTimeout{Phase: r.s.Phase, Seq: r.s.Seq})
}

// living lists the living players in seating order.
func (r *dayRun) living() []engine.PlayerID {
	var out []engine.PlayerID
	for _, id := range r.s.Seats {
		if r.s.Players[id].Alive {
			out = append(out, id)
		}
	}
	return out
}

// NobodyHanged presses skip for the living players, in seating order, until
// the day gives up. Votes already cast do not matter: skipping is last.
func (realDriver) NobodyHanged(s engine.State) ([]engine.Command, error) {
	r := &dayRun{s: s}
	if r.s.Phase != engine.PhaseDay && r.s.Phase != engine.PhaseNominationLock {
		return nil, fmt.Errorf("%w: cannot end the day from %s", ErrWrongStage, r.s.Phase)
	}
	for _, id := range r.living() {
		if r.s.Phase != engine.PhaseDay && r.s.Phase != engine.PhaseNominationLock {
			break
		}
		if err := r.play(engine.SkipDay{Voter: id, On: true}); err != nil {
			return nil, err
		}
	}
	if r.s.Phase != engine.PhaseDusk {
		return nil, fmt.Errorf("%w: the day did not end (now %s)", ErrWrongStage, r.s.Phase)
	}
	return r.cmds, nil
}

// Hang has everybody else nominate the player (who chooses "no one"), lets the
// defense run out, and has everybody but the defendant vote to kill. Whatever
// is already done is skipped, so a scenario may play any part of it by hand.
func (realDriver) Hang(s engine.State, who engine.PlayerID) ([]engine.Command, error) {
	r := &dayRun{s: s}
	if r.s.Phase == engine.PhaseDay {
		for _, id := range r.living() {
			target := who
			if id == who {
				target = ""
			}
			if err := r.play(engine.Nominate{Voter: id, Target: target}); err != nil {
				return nil, err
			}
		}
	}
	if r.s.Phase == engine.PhaseNominationLock { // nobody changes their mind
		if err := r.timeout(); err != nil {
			return nil, err
		}
	}
	if r.s.Phase == engine.PhaseTrial && r.s.Day.Defendant == who {
		if err := r.timeout(); err != nil {
			return nil, err
		}
	}
	if r.s.Phase == engine.PhaseVerdict && r.s.Day.Defendant == who {
		for _, id := range r.living() {
			if id == who {
				continue
			}
			if err := r.play(engine.CastVerdict{Voter: id, Kill: true}); err != nil {
				return nil, err
			}
		}
	}
	if p := r.s.Players[who]; p == nil || p.Alive {
		return nil, fmt.Errorf("%w: %s was not hanged (phase %s)", ErrWrongStage, who, r.s.Phase)
	}
	return r.cmds, nil
}
