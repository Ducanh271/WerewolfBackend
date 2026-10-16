package engine

import (
	"fmt"
	"slices"
	"time"
)

// SkipThresholdPercent: the day is skipped when more than this share of the
// living players press the skip button (rules §4).
const SkipThresholdPercent = 66

// Nominate is a public vote to put Target on the trial stand, or, with an empty
// Target, for "no one". A voter may change the choice until nominations close.
type Nominate struct {
	CommandMarker
	Voter  PlayerID
	Target PlayerID
}

// SkipDay presses (On) or releases the skip button. Skips count one per
// person; the Chief is not doubled.
type SkipDay struct {
	CommandMarker
	Voter PlayerID
	On    bool
}

// NominationCast records a voter's choice; a later one replaces it. A real
// target cancels the voter's own skip.
type NominationCast struct {
	EventMarker
	Voter  PlayerID
	Target PlayerID
}

// SkipSet records the skip button being pressed or released.
type SkipSet struct {
	EventMarker
	Voter PlayerID
	On    bool
}

// NominationsLocking starts (or restarts) the countdown after every living
// player has chosen. Anyone may still change their choice until it runs out.
type NominationsLocking struct {
	EventMarker
	Duration time.Duration
}

// NominationsReopened ends a countdown whose last change left nobody
// nominated: the day clock goes on and discussion continues.
type NominationsReopened struct {
	EventMarker
}

// dayVoter checks the parts of a day action every vote shares.
func dayVoter(s State, id PlayerID, phases ...Phase) error {
	if !slices.Contains(phases, s.Phase) {
		return fmt.Errorf("%w: cannot do this during %s", ErrWrongPhase, s.Phase)
	}
	p, ok := s.Players[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	if !p.Alive {
		return fmt.Errorf("%w: %s", ErrNotAlive, id)
	}
	return nil
}

func decideNominate(s State, cmd Nominate) ([]Event, error) {
	if err := dayVoter(s, cmd.Voter, PhaseDay, PhaseNominationLock); err != nil {
		return nil, err
	}
	if cmd.Target != "" {
		if _, err := LivingTarget(s, cmd.Target); err != nil {
			return nil, err
		}
		if cmd.Target == cmd.Voter {
			return nil, ErrSelfNomination
		}
		if s.Day.Spared[cmd.Target] {
			return nil, fmt.Errorf("%w: %s", ErrAlreadySpared, cmd.Target)
		}
	}
	// A real change restarts the countdown (rules §4 "mỗi lần có người đổi");
	// resending the same pick is not a change.
	changed := true
	if prev, had := s.Day.Nominations[cmd.Voter]; had {
		changed = prev.Target != cmd.Target
	}
	evs := []Event{NominationCast{Voter: cmd.Voter, Target: cmd.Target}}
	after := Fold(s, evs...)
	if !nominationsClosed(after) {
		return evs, nil
	}
	// Everybody has chosen. Each real change restarts the countdown; when the
	// change leaves nobody nominated, discussion goes on.
	if _, ok := topNominee(after); ok {
		if !changed {
			return evs, nil
		}
		return append(evs, NominationsLocking{Duration: s.Settings.NominationLock}), nil
	}
	if s.Phase == PhaseNominationLock {
		evs = append(evs, NominationsReopened{})
	}
	return evs, nil
}

func decideSkipDay(s State, cmd SkipDay) ([]Event, error) {
	if err := dayVoter(s, cmd.Voter, PhaseDay, PhaseNominationLock); err != nil {
		return nil, err
	}
	evs := []Event{SkipSet{Voter: cmd.Voter, On: cmd.On}}
	after := Fold(s, evs...)
	if len(after.Day.Skips)*100 > SkipThresholdPercent*livingCount(after) {
		evs = append(evs, DayEnded{})
	}
	return evs, nil
}

// nominationsClosed is true once every living player has made a choice.
func nominationsClosed(s State) bool {
	for id, p := range s.Players {
		if _, chose := s.Day.Nominations[id]; p.Alive && !chose {
			return false
		}
	}
	return true
}

// topNominee returns who goes on trial: the most weighted nomination votes, a
// tie going to the target whose earliest standing vote came first. ok is false
// when nobody was nominated.
func topNominee(s State) (top PlayerID, ok bool) {
	score := map[PlayerID]int{}
	first := map[PlayerID]int{}
	for voter, n := range s.Day.Nominations {
		if n.Target == "" {
			continue
		}
		score[n.Target] += voteWeight(s, voter)
		if f, seen := first[n.Target]; !seen || n.Order < f {
			first[n.Target] = n.Order
		}
	}
	best := 0
	for target, sc := range score {
		if sc > best || (sc == best && first[target] < first[top]) {
			top, best = target, sc
		}
	}
	return top, best > 0
}
