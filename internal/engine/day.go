package engine

import (
	"time"

	"masoi/internal/i18n"
)

// Nomination is one player's current public choice. Target is empty for "no
// one". Order says when the choice was made, so a tie goes to the player whose
// earliest standing vote came first (rules §4); changing a vote takes a new
// place in the queue.
type Nomination struct {
	Target PlayerID
	Order  int
}

// DayState holds nominations, skip votes, the trial and what the day has
// already decided. The day clock itself lives in the room layer, which pauses
// it during the trial; the engine only hears that it ran out (ClockOut).
type DayState struct {
	Nominations map[PlayerID]Nomination // voter -> choice
	NextOrder   int
	Skips       map[PlayerID]bool
	Spared      map[PlayerID]bool // spared today: kept until the night
	ClockOut    bool              // the 7 minutes ran out; the verdict is the last step
	Defendant   PlayerID          // on trial, empty outside Trial and Verdict
	Verdicts    map[PlayerID]bool // voter -> true for kill, false for spare
}

func (d DayState) clone() DayState {
	out := d
	out.Nominations = copyMap(d.Nominations)
	out.Skips = copyMap(d.Skips)
	out.Spared = copyMap(d.Spared)
	out.Verdicts = copyMap(d.Verdicts)
	return out
}

func copyMap[K comparable, V any](in map[K]V) map[K]V {
	if in == nil {
		return nil
	}
	out := make(map[K]V, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// voteWeight is 2 for the Chief and 1 for everybody else (rules §7b).
func voteWeight(s State, voter PlayerID) int {
	if voter != "" && voter == s.Chief {
		return 2
	}
	return 1
}

// livingCount is the number of players still in the game.
func livingCount(s State) int {
	n := 0
	for _, p := range s.Players {
		if p.Alive {
			n++
		}
	}
	return n
}

// DayStarted opens the day after the dawn announcement.
type DayStarted struct {
	EventMarker
	Duration time.Duration
}

// DayEnded closes a day nobody was hanged in; the game waits at dusk.
type DayEnded struct {
	EventMarker
}

// AnnouncementKey: the narrator says that nobody was hanged.
func (DayEnded) AnnouncementKey() string { return i18n.DayNobodyHanged }

func evolveDay(s State, e Event) (State, bool) {
	switch ev := e.(type) {
	case DayStarted:
		out := s.clone()
		out.Phase = PhaseDay
		out.Day = DayState{}
		out.Seq++
		return out, true
	case DayEnded:
		out := s.clone()
		out.Phase = PhaseDusk
		out.Day.Defendant = ""
		out.Seq++
		return out, true
	case NominationsLocking:
		out := s.clone()
		out.Phase = PhaseNominationLock
		out.Seq++
		return out, true
	case NominationsReopened:
		out := s.clone()
		out.Phase = PhaseDay
		out.Seq++
		return out, true
	case NominationCast:
		out := s.clone()
		if out.Day.Nominations == nil {
			out.Day.Nominations = map[PlayerID]Nomination{}
		}
		out.Day.NextOrder++
		out.Day.Nominations[ev.Voter] = Nomination{Target: ev.Target, Order: out.Day.NextOrder}
		if ev.Target != "" {
			delete(out.Day.Skips, ev.Voter) // nominating cancels the skip
		}
		return out, true
	case SkipSet:
		out := s.clone()
		if out.Day.Skips == nil {
			out.Day.Skips = map[PlayerID]bool{}
		}
		if ev.On {
			out.Day.Skips[ev.Voter] = true
		} else {
			delete(out.Day.Skips, ev.Voter)
		}
		return out, true
	}
	return evolveTrial(s, e)
}
