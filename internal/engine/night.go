package engine

import (
	"sort"
	"time"
)

// NightStarted records the start of a night.
type NightStarted struct {
	EventMarker
	Night int
}

// NightStepCalled is the narrator's cue: call this role now and wait Duration.
// It is the same for a living role and for a dead or absent one (rules §3,
// "gọi giả"); nothing in it says which, and the step always lasts the full
// Duration. It is only ever shown to the narrator, never to players.
type NightStepCalled struct {
	EventMarker
	Role     RoleID
	Duration time.Duration
}

// NightEnded records that the last step ran out and the night is over.
type NightEnded struct {
	EventMarker
}

// nightSteps lists the roles called on a night, in call order. The role list
// of a game is public (rules §2), so only the roles in the configuration are
// called, plus the wolf pack step, which every game has. Beyond that it depends on the night number only, never on who is
// alive, so a fake call (a dead player's role) cannot be told from a real one.
func nightSteps(s State, night int) []RoleID {
	type step struct {
		id    RoleID
		order int
	}
	var steps []step
	for id, role := range Registry {
		if always, ok := role.(AlwaysCalled); !(ok && always.AlwaysCalled()) && s.Settings.Roles[id] < 1 {
			continue
		}
		if call := role.NightCall(); call != nil && call.CalledOn(night) {
			steps = append(steps, step{id, call.Order})
		}
	}
	sort.Slice(steps, func(i, j int) bool {
		if steps[i].order != steps[j].order {
			return steps[i].order < steps[j].order
		}
		return steps[i].id < steps[j].id
	})
	ids := make([]RoleID, len(steps))
	for i, st := range steps {
		ids[i] = st.id
	}
	return ids
}

// CurrentNightRole returns the role being called right now. Roles use it to
// accept night actions only during their own step.
func CurrentNightRole(s State) (RoleID, bool) {
	if s.Phase != PhaseNight || s.NightStep < 1 {
		return "", false
	}
	steps := nightSteps(s, s.Night)
	if s.NightStep > len(steps) {
		return "", false
	}
	return steps[s.NightStep-1], true
}

// nextNightEvent is the call for the step after `called` steps, or the end of
// the night when none is left.
func nextNightEvent(s State, night, called int) Event {
	steps := nightSteps(s, night)
	if called >= len(steps) {
		return NightEnded{}
	}
	return NightStepCalled{Role: steps[called], Duration: s.Settings.NightStep}
}

// startNight begins a night; night 1 follows the Chief election.
func startNight(s State, night int) []Event {
	return []Event{NightStarted{Night: night}, nextNightEvent(s, night, 0)}
}

// decidePhaseTimeout moves the game on when a deadline passes. A night step
// ends only by timeout, never early, so every call lasts the same time.
func decidePhaseTimeout(s State, cmd PhaseTimeout) ([]Event, error) {
	if cmd.Phase != s.Phase || cmd.Seq != s.Seq {
		return nil, nil // stale
	}
	switch s.Phase {
	case PhaseRoleView:
		return []Event{ChiefElectionStarted{Duration: s.Settings.ChiefElection}}, nil
	case PhaseChiefElection:
		return append([]Event{ChiefElected{Chief: tallyChief(s, cmd.Seed)}}, startNight(s, 1)...), nil
	case PhaseChiefSuccession:
		return []Event{ChiefLost{Former: s.Succession.Former}}, nil
	case PhaseDawn, PhaseDay, PhaseNominationLock, PhaseTrial, PhaseVerdict, PhaseDusk:
		return decideDayTimeout(s, cmd)
	case PhaseNight:
		next := nextNightEvent(s, s.Night, s.NightStep)
		if _, over := next.(NightEnded); over {
			return append([]Event{next}, resolveNight(s)...), nil
		}
		return []Event{next}, nil
	}
	return nil, nil
}

func evolveNight(s State, e Event) (State, bool) {
	switch ev := e.(type) {
	case NightStarted:
		out := s.clone()
		out.Phase = PhaseNight
		out.Night = ev.Night
		out.NightStep = 0
		out.Pending = NightActions{}
		out.Seq++
		return out, true
	case NightStepCalled:
		out := s.clone()
		out.NightStep++
		out.Seq++
		return out, true
	case NightEnded:
		out := s.clone()
		out.Phase = PhaseDawn
		out.Seq++
		return out, true
	}
	return evolveResolution(s, e)
}
