package engine

import (
	"fmt"
	"time"

	"masoi/internal/i18n"
)

// The Chief (Trưởng làng) is a public title, not a role (rules §7b): the
// holder's role and team never change and it does not count towards the wolf
// limit. The vote weight belongs to the day phase (rules §4, §7b).

// VoteChief is a player's public vote in the Chief election. A voter may vote
// for themselves and may change the vote until time is up.
type VoteChief struct {
	CommandMarker
	Voter     PlayerID
	Candidate PlayerID
}

// NameSuccessor is the dead Chief choosing who inherits the title.
type NameSuccessor struct {
	CommandMarker
	By        PlayerID
	Successor PlayerID
}

// ChiefElectionStarted opens the election phase.
type ChiefElectionStarted struct {
	EventMarker
	Duration time.Duration
}

// ChiefVoteCast is a public vote; a later one by the same voter replaces it.
type ChiefVoteCast struct {
	EventMarker
	Voter     PlayerID
	Candidate PlayerID
}

// ChiefElected ends the election.
type ChiefElected struct {
	EventMarker
	Chief PlayerID
}

// ChiefSuccessionStarted opens the time in which the dead Chief names a
// successor (rules §7b, Settings.ChiefSuccession). The title is empty
// meanwhile.
type ChiefSuccessionStarted struct {
	EventMarker
	Former   PlayerID
	Duration time.Duration
}

// ChiefSucceeded passes the title on.
type ChiefSucceeded struct {
	EventMarker
	From PlayerID
	To   PlayerID
}

// ChiefLost means the Chief did not name anybody in time: the title is gone.
type ChiefLost struct {
	EventMarker
	Former PlayerID
}

// Succession remembers whose title is being passed on and which phase to go
// back to afterwards.
type Succession struct {
	Former PlayerID
	Resume Phase
}

// AnnouncementKey implementations: all four are read aloud to everybody.
func (ChiefElectionStarted) AnnouncementKey() string { return i18n.ChiefElectionStarted }
func (ChiefElected) AnnouncementKey() string         { return i18n.ChiefElected }
func (ChiefSucceeded) AnnouncementKey() string       { return i18n.ChiefSucceeded }
func (ChiefLost) AnnouncementKey() string            { return i18n.ChiefLost }

func evolveChief(s State, e Event) (State, bool) {
	switch ev := e.(type) {
	case ChiefElectionStarted:
		out := s.clone()
		out.Phase = PhaseChiefElection
		out.ChiefVotes = map[PlayerID]PlayerID{}
		out.Seq++
		return out, true
	case ChiefVoteCast:
		out := s.clone()
		out.ChiefVotes[ev.Voter] = ev.Candidate
		return out, true
	case ChiefElected:
		out := s.clone()
		out.Chief = ev.Chief
		out.ChiefVotes = nil
		return out, true
	case ChiefSuccessionStarted:
		out := s.clone()
		out.Succession = Succession{Former: ev.Former, Resume: s.Phase}
		out.Chief = ""
		out.Phase = PhaseChiefSuccession
		out.Seq++
		return out, true
	case ChiefSucceeded:
		out := s.clone()
		out.Chief = ev.To
		return endSuccession(out), true
	case ChiefLost:
		return endSuccession(s.clone()), true
	}
	return s, false
}

func endSuccession(s State) State {
	s.Phase = s.Succession.Resume
	s.Succession = Succession{}
	s.Seq++
	return s
}

func decideVoteChief(s State, cmd VoteChief) ([]Event, error) {
	if s.Phase != PhaseChiefElection {
		return nil, fmt.Errorf("%w: cannot vote for the Chief during %s", ErrWrongPhase, s.Phase)
	}
	if _, err := LivingTarget(s, cmd.Voter); err != nil {
		return nil, fmt.Errorf("%w: voter %s", ErrUnknownPlayer, cmd.Voter)
	}
	if _, err := LivingTarget(s, cmd.Candidate); err != nil {
		return nil, err
	}
	return []Event{ChiefVoteCast{Voter: cmd.Voter, Candidate: cmd.Candidate}}, nil
}

func decideNameSuccessor(s State, cmd NameSuccessor) ([]Event, error) {
	if s.Phase != PhaseChiefSuccession {
		return nil, fmt.Errorf("%w: no Chief is passing on the title", ErrWrongPhase)
	}
	if cmd.By != s.Succession.Former {
		return nil, fmt.Errorf("%w: only the dead Chief names the successor", ErrNotYourTurn)
	}
	if _, err := LivingTarget(s, cmd.Successor); err != nil {
		return nil, err
	}
	return []Event{ChiefSucceeded{From: cmd.By, To: cmd.Successor}}, nil
}
