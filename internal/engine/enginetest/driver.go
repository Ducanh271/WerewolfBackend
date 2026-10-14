package enginetest

import (
	"errors"

	"masoi/internal/engine"
)

// Driver turns the steps of a scenario into engine commands. The default
// driver plays the real night machine (T4); the day half arrives with T7.
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

// ErrDayNotBuilt is returned by the default driver for the day and for nights
// after the first.
var ErrDayNotBuilt = errors.New("the day phase machine is not built yet (task T7)")

// realDriver drives the engine's own night machine.
type realDriver struct{}

func (realDriver) StartNight(s engine.State) ([]engine.Command, error) {
	if s.Phase != engine.PhaseRoleView {
		return nil, ErrDayNotBuilt // night 2 and later follow a day
	}
	return []engine.Command{engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq}}, nil
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

func (realDriver) NobodyHanged(engine.State) ([]engine.Command, error) {
	return nil, ErrDayNotBuilt
}

func (realDriver) Hang(engine.State, engine.PlayerID) ([]engine.Command, error) {
	return nil, ErrDayNotBuilt
}
