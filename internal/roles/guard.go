package roles

import (
	"fmt"

	"masoi/internal/engine"
)

// Guard shields one player a night from the bite (rules §6 "Bảo vệ"). It can
// shield itself, but not the same player two nights in a row. A night without a
// choice breaks the streak.
type guard struct{}

// Guard is the ID of the guard.
const Guard engine.RoleID = "Guard"

// Flags kept on the Guard's player: the night of the last shield and whom it
// covered (seat index + 1, so 0 means nobody).
const (
	flagLastNight  = "guard.last_night"
	flagLastTarget = "guard.last_target"
)

// GuardProtect is the Guard's choice for tonight.
type GuardProtect struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// PlayerProtected records the Guard's choice.
type PlayerProtected struct {
	engine.EventMarker
	Guard  engine.PlayerID
	Target engine.PlayerID
	Night  int
}

func init() { engine.Register(guard{}) }

// Audience: only the Guard knows whom she protected.
func (e PlayerProtected) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Guard}}
}

func (guard) ID() engine.RoleID            { return Guard }
func (guard) BaseTeam() engine.Team        { return engine.TeamVillage }
func (guard) NightCall() *engine.NightCall { return &engine.NightCall{Order: 2} }

func (guard) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(GuardProtect)
	if !ok {
		return nil, false, nil
	}
	me, err := engine.NightActor(s, cmd.By, Guard)
	if err != nil {
		return nil, true, err
	}
	if s.Pending.Protected != "" {
		return nil, true, engine.ErrAlreadyActed
	}
	if _, err := engine.LivingTarget(s, cmd.Target); err != nil {
		return nil, true, err
	}
	if me.Flags[flagLastNight] == s.Night-1 && me.Flags[flagLastTarget] == seatNumber(s, cmd.Target) {
		return nil, true, fmt.Errorf("%w: %s", engine.ErrSameGuardTarget, cmd.Target)
	}
	return []engine.Event{PlayerProtected{Guard: cmd.By, Target: cmd.Target, Night: s.Night}}, true, nil
}

func (guard) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(PlayerProtected)
	if !ok {
		return false
	}
	s.Pending.Protected = ev.Target
	if p := s.Players[ev.Guard]; p != nil {
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		p.Flags[flagLastNight] = ev.Night
		p.Flags[flagLastTarget] = seatNumber(*s, ev.Target)
	}
	return true
}

func seatNumber(s engine.State, id engine.PlayerID) int {
	for i, seat := range s.Seats {
		if seat == id {
			return i + 1
		}
	}
	return 0
}
