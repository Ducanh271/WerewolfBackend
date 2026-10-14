package enginetest

import (
	"errors"

	"masoi/internal/engine"
)

// Driver turns the steps of a scenario into engine commands. The engine has no
// night or day phase machine yet, so the kit does not hardcode one: the task
// that builds it (T4 for nights, T7 for days) supplies the real driver and
// makes it the default.
type Driver interface {
	// StartNight returns the commands that bring the game into the next night
	// (for the first night: role viewing, Chief election, host start).
	StartNight(s engine.State) ([]engine.Command, error)
	// EndNight times out every role that has not acted (rules §3) and
	// resolves the night into dawn.
	EndNight(s engine.State) ([]engine.Command, error)
	// NobodyHanged ends the day without an execution.
	NobodyHanged(s engine.State) ([]engine.Command, error)
	// Hang ends the day with the execution of one player.
	Hang(s engine.State, who engine.PlayerID) ([]engine.Command, error)
}

// ErrPhasesNotBuilt is returned by the default driver.
var ErrPhasesNotBuilt = errors.New("the night/day phase machine is not built yet (tasks T4, T7)")

type unbuiltDriver struct{}

func (unbuiltDriver) StartNight(engine.State) ([]engine.Command, error) {
	return nil, ErrPhasesNotBuilt
}

func (unbuiltDriver) EndNight(engine.State) ([]engine.Command, error) {
	return nil, ErrPhasesNotBuilt
}

func (unbuiltDriver) NobodyHanged(engine.State) ([]engine.Command, error) {
	return nil, ErrPhasesNotBuilt
}

func (unbuiltDriver) Hang(engine.State, engine.PlayerID) ([]engine.Command, error) {
	return nil, ErrPhasesNotBuilt
}
