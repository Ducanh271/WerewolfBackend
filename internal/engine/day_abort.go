package engine

// CastAbortVote is a living player's vote to call the game off before it has
// a winner (audit K1, rules §8b). A later vote by the same player replaces
// the earlier one, so Yes: false withdraws it.
type CastAbortVote struct {
	CommandMarker
	Voter PlayerID
	Yes   bool
}

// AbortVoteCast records one player's current abort vote.
type AbortVoteCast struct {
	EventMarker
	Voter PlayerID
	Yes   bool
}

// dayPhases is every phase that counts as "by day" for the abort vote
// (rules §8b).
var dayPhases = []Phase{PhaseDay, PhaseNominationLock, PhaseTrial, PhaseVerdict, PhaseDusk}

func decideCastAbortVote(s State, cmd CastAbortVote) ([]Event, error) {
	if err := dayVoter(s, cmd.Voter, dayPhases...); err != nil {
		return nil, err
	}
	evs := []Event{AbortVoteCast{Voter: cmd.Voter, Yes: cmd.Yes}}
	if abortVoteDue(Fold(s, evs...)) {
		evs = append(evs, GameEnded{Outcome: Outcome{Draw: true, Aborted: true}})
	}
	return evs, nil
}

// abortVoteDue reports whether the pending abort votes, counted against the
// players alive right now, reach 2/3 of them: yes*3 >= alive*2 (rules §8b).
// Votes from players who died since casting no longer count, so a death can
// bring a pending vote over the line on its own, without a new cast.
func abortVoteDue(s State) bool {
	if len(s.Day.AbortVotes) == 0 {
		return false
	}
	yes, alive := 0, 0
	for id, p := range s.Players {
		if !p.Alive {
			continue
		}
		alive++
		if s.Day.AbortVotes[id] {
			yes++
		}
	}
	return alive > 0 && yes*3 >= alive*2
}
