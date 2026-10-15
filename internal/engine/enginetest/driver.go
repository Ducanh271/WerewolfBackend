package enginetest

import (
	"errors"
	"fmt"

	"masoi/internal/engine"
)

// Driver turns the steps of a scenario into engine commands. The default
// driver plays the real night machine (T4) and the real day machine (T7).
type Driver interface {
	// StartNight returns the commands that bring the game into the next night
	// (for the first night: the role-view countdown runs out).
	StartNight(s engine.State) ([]engine.Command, error)
	// EndNight times out every role that has not acted (rules §3) and
	// resolves the night into dawn.
	EndNight(s engine.State) ([]engine.Command, error)
	// NobodyHanged ends the day without an execution.
	NobodyHanged(s engine.State) ([]engine.Command, error)
	// Hang ends the day with the execution of one player.
	Hang(s engine.State, who engine.PlayerID) ([]engine.Command, error)
}

// ErrWrongStage is returned by the default driver when a step is asked for in
// a game phase it cannot start from.
var ErrWrongStage = errors.New("the game is not in a phase this step can start from")

// realDriver drives the engine's own night machine.
type realDriver struct {
	chief engine.PlayerID // everybody votes for them when the scenario did not hold the election
}

// electionSeed settles the election of scenarios that do not play it.
const electionSeed = 1

// StartNight runs the role view and the Chief election out, then night 1.
// Like EndNight it works on a private copy because each step changes Seq.
func (d realDriver) StartNight(s engine.State) ([]engine.Command, error) {
	if s.Phase == engine.PhaseDusk { // night 2 and later follow a day
		return []engine.Command{engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq}}, nil
	}
	if s.Phase != engine.PhaseRoleView && s.Phase != engine.PhaseChiefElection {
		return nil, fmt.Errorf("%w: cannot start a night from %s", ErrWrongStage, s.Phase)
	}
	var cmds []engine.Command
	play := func(c engine.Command) error {
		evs, err := engine.Decide(s, c)
		if err != nil {
			return err
		}
		cmds = append(cmds, c)
		s = engine.Fold(s, evs...)
		return nil
	}
	if s.Phase == engine.PhaseRoleView {
		if err := play(engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq}); err != nil {
			return nil, err
		}
		if d.chief != "" {
			for _, id := range s.Seats {
				if err := play(engine.VoteChief{Voter: id, Candidate: d.chief}); err != nil {
					return nil, err
				}
			}
		}
	}
	err := play(engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq, Seed: electionSeed})
	return cmds, err
}

// EndNight times out every remaining step. The commands are worked out on a
// private copy of the state because each timeout changes Seq.
func (realDriver) EndNight(s engine.State) ([]engine.Command, error) {
	var cmds []engine.Command
	for s.Phase == engine.PhaseNight {
		cmd := engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq}
		evs, err := engine.Decide(s, cmd)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
		s = engine.Fold(s, evs...)
	}
	return cmds, nil
}
