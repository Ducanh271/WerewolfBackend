package engine

import (
	"fmt"
	"time"

	"masoi/internal/i18n"
)

// CastVerdict is a public vote to kill (Kill) or spare the player on trial. It
// may be changed until the vote closes. The player on trial cannot vote.
type CastVerdict struct {
	CommandMarker
	Voter PlayerID
	Kill  bool
}

// TrialStarted puts a player on the stand for their defense.
type TrialStarted struct {
	EventMarker
	Defendant PlayerID
	Duration  time.Duration
}

// DayClockExpired records that the 7 minutes ran out while somebody was
// nominated: that player's trial is the last step of the day.
type DayClockExpired struct {
	EventMarker
}

// VerdictStarted ends the defense and opens the kill or spare vote.
type VerdictStarted struct {
	EventMarker
	Duration time.Duration
}

// VerdictCast records a public vote; a later one by the same voter replaces it.
type VerdictCast struct {
	EventMarker
	Voter PlayerID
	Kill  bool
}

// Spared ends the trial without an execution. If the clock has not run out the
// day goes back to nominations; the spared player stays safe until the night.
type Spared struct {
	EventMarker
	Player PlayerID
}

// Hanged is the execution; PlayerDied comes with it. The role is not named.
type Hanged struct {
	EventMarker
	Player PlayerID
}

func (TrialStarted) AnnouncementKey() string   { return i18n.TrialStarted }
func (VerdictStarted) AnnouncementKey() string { return i18n.VerdictStarted }
func (Spared) AnnouncementKey() string         { return i18n.DaySpared }
func (Hanged) AnnouncementKey() string         { return i18n.DayHanged }

func startTrial(s State, defendant PlayerID) []Event {
	return []Event{TrialStarted{Defendant: defendant, Duration: s.Settings.Defense}}
}

func evolveTrial(s State, e Event) (State, bool) {
	switch ev := e.(type) {
	case DayClockExpired:
		out := s.clone()
		out.Day.ClockOut = true
		return out, true
	case TrialStarted:
		out := s.clone()
		out.Phase = PhaseTrial
		out.Day.Defendant = ev.Defendant
		out.Day.Verdicts = map[PlayerID]bool{}
		out.Seq++
		return out, true
	case VerdictStarted:
		out := s.clone()
		out.Phase = PhaseVerdict
		out.Seq++
		return out, true
	case VerdictCast:
		out := s.clone()
		out.Day.Verdicts[ev.Voter] = ev.Kill
		return out, true
	case Spared:
		out := s.clone()
		if out.Day.Spared == nil {
			out.Day.Spared = map[PlayerID]bool{}
		}
		out.Day.Spared[ev.Player] = true
		// The spared list stays for the whole day; only the votes start over.
		out.Day.Nominations, out.Day.Skips, out.Day.Verdicts = nil, nil, nil
		out.Day.Defendant = ""
		out.Phase = PhaseDay
		out.Seq++
		return out, true
	case Hanged:
		out := s.clone()
		out.Day.Defendant = ""
		out.Phase = PhaseDusk
		out.Seq++
		return out, true
	}
	return s, false
}

func decideCastVerdict(s State, cmd CastVerdict) ([]Event, error) {
	if err := dayVoter(s, cmd.Voter, PhaseVerdict); err != nil {
		return nil, err
	}
	if cmd.Voter == s.Day.Defendant {
		return nil, ErrOnTrial
	}
	evs := []Event{VerdictCast{Voter: cmd.Voter, Kill: cmd.Kill}}
	after := Fold(s, evs...)
	if verdictsComplete(after) {
		evs = append(evs, resolveVerdict(after)...)
	}
	return evs, nil
}

// verdictsComplete is true once every living player except the defendant voted.
func verdictsComplete(s State) bool {
	for id, p := range s.Players {
		if _, voted := s.Day.Verdicts[id]; p.Alive && id != s.Day.Defendant && !voted {
			return false
		}
	}
	return true
}

// resolveVerdict hangs only when the kill votes beat the spare votes (rules
// §4); a tie spares. It shares afterDeaths with the night, so the Hunter and
// the Chief's succession follow a hanging exactly as they follow a bite.
func resolveVerdict(s State) []Event {
	kill, spare := 0, 0
	for voter, k := range s.Day.Verdicts {
		if k {
			kill += voteWeight(s, voter)
		} else {
			spare += voteWeight(s, voter)
		}
	}
	who := s.Day.Defendant
	if kill > spare {
		evs := []Event{PlayerDied{Player: who}, Hanged{Player: who}}
		return append(evs, afterDeaths(s, evs, []PlayerID{who})...)
	}
	evs := []Event{Spared{Player: who}}
	if s.Day.ClockOut {
		evs = append(evs, DayEnded{})
	}
	return evs
}

func decideDayTimeout(s State, cmd PhaseTimeout) ([]Event, error) {
	switch s.Phase {
	case PhaseDawn:
		return []Event{DayStarted{Duration: s.Settings.Day}}, nil
	case PhaseNominationLock:
		// Nobody changed their mind, or the day clock ran out meanwhile: either
		// way the top nominee goes on trial, the clock case being the last one.
		if top, ok := topNominee(s); ok {
			if cmd.DayClock {
				return append([]Event{DayClockExpired{}}, startTrial(s, top)...), nil
			}
			return startTrial(s, top), nil
		}
		return []Event{NominationsReopened{}}, nil
	case PhaseDay:
		// The 7 minutes are up: the top nominee still goes through the trial.
		if top, ok := topNominee(s); ok {
			return append([]Event{DayClockExpired{}}, startTrial(s, top)...), nil
		}
		return []Event{DayEnded{}}, nil
	case PhaseTrial:
		return []Event{VerdictStarted{Duration: s.Settings.Verdict}}, nil
	case PhaseVerdict:
		return resolveVerdict(s), nil
	case PhaseDusk:
		return startNight(s, s.Night+1), nil
	}
	return nil, fmt.Errorf("%w: no day timer in %s", ErrWrongPhase, s.Phase)
}
