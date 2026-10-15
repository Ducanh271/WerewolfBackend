package engine

import "fmt"

// NightActor checks the parts of a night action that every role shares, in
// this order:
//
//  1. the actor holds the step's role (for the pack step: is on the wolf team).
//     Anybody else gets the bare ErrNotYourTurn, the same error with the same
//     text whatever their state, so a probing player cannot learn whether they
//     were blocked, are dead by the rules or the role is alive (rules §3);
//  2. it is the night and that step is being called;
//  3. the holder is alive, not blocked tonight (ErrBlocked) and not dead by
//     the rules from this step (Doomed). The holder sees all of that on their
//     own screen already.
//
// A new blocking role only has to record the block in Pending.Blocked and
// every role stops acting.
func NightActor(s State, by PlayerID, step RoleID) (*Player, error) {
	p, ok := s.Players[by]
	if !ok || !holdsStep(p, step) {
		return nil, ErrNotYourTurn
	}
	if cur, ok := CurrentNightRole(s); !ok || cur != step {
		return nil, fmt.Errorf("%w: %s is not being called", ErrNotYourTurn, step)
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

// holdsStep: the player is the one a step calls. Every wolf-team player holds
// the pack step, a cursed player too.
func holdsStep(p *Player, step RoleID) bool {
	if isPackStep(step) {
		return p.Team == TeamWolves
	}
	return p.Role == step
}

// LivingTarget returns the chosen player, who must be seated and alive.
func LivingTarget(s State, id PlayerID) (*Player, error) {
	p, ok := s.Players[id]
	if !ok || !p.Alive {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTarget, id)
	}
	return p, nil
}
