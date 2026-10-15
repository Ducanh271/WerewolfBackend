package engine

import "fmt"

// NightActor checks the parts of a night action that every role shares: it is
// the night, it is the given role's step, the actor is a living player and was
// not blocked tonight (ErrBlocked). A new blocking role only has to record the
// block in Pending.Blocked and every role stops acting.
// Roles then check that the actor really holds their role.
func NightActor(s State, by PlayerID, step RoleID) (*Player, error) {
	if cur, ok := CurrentNightRole(s); !ok || cur != step {
		return nil, fmt.Errorf("%w: %s is not being called", ErrNotYourTurn, step)
	}
	p, ok := s.Players[by]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownPlayer, by)
	}
	if !p.Alive {
		return nil, fmt.Errorf("%w: %s", ErrNotAlive, by)
	}
	if Blocked(s, by) {
		return nil, fmt.Errorf("%w: %s", ErrBlocked, by)
	}
	if Doomed(s, by) {
		return nil, fmt.Errorf("%w: %s is dead by the rules from this step", ErrNotYourTurn, by)
	}
	return p, nil
}

// LivingTarget returns the chosen player, who must be seated and alive.
func LivingTarget(s State, id PlayerID) (*Player, error) {
	p, ok := s.Players[id]
	if !ok || !p.Alive {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTarget, id)
	}
	return p, nil
}
