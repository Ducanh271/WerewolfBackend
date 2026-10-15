package roles

import (
	"fmt"

	"masoi/internal/engine"
)

// Seer checks one player a night and learns wolf or not (rules §6 "Tiên tri").
type seer struct{}

// Seer is the ID of the seer.
const Seer engine.RoleID = "Seer"

// SeerCheck is the Seer's choice for tonight.
type SeerCheck struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// SeerChecked is the Seer's private result. Anyone on the wolf team, including
// a cursed player, shows as a wolf.
type SeerChecked struct {
	engine.EventMarker
	Seer   engine.PlayerID
	Target engine.PlayerID
	IsWolf bool
}

func init() { engine.Register(seer{}) }

func (seer) ID() engine.RoleID            { return Seer }
func (seer) BaseTeam() engine.Team        { return engine.TeamVillage }
func (seer) NightCall() *engine.NightCall { return &engine.NightCall{Order: 7} }

func (seer) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(SeerCheck)
	if !ok {
		return nil, false, nil
	}
	me, err := engine.NightActor(s, cmd.By, Seer)
	if err != nil {
		return nil, true, err
	}
	if me.Role != Seer {
		return nil, true, fmt.Errorf("%w: %s is not the Seer", engine.ErrNotYourTurn, cmd.By)
	}
	if s.Pending.SeerTarget != "" {
		return nil, true, engine.ErrAlreadyActed
	}
	target, err := engine.LivingTarget(s, cmd.Target)
	if err != nil {
		return nil, true, err
	}
	return []engine.Event{SeerChecked{
		Seer:   cmd.By,
		Target: cmd.Target,
		IsWolf: target.Team == engine.TeamWolves,
	}}, true, nil
}

func (seer) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(SeerChecked)
	if !ok {
		return false
	}
	s.Pending.SeerTarget = ev.Target
	return true
}
