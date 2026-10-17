package room

import (
	"fmt"

	"masoi/internal/engine"
)

// PlayerCommand builds a command for the player it is told about. The caller
// (the transport) never puts an id of its own in: the room fills it in from
// the connection.
type PlayerCommand func(by engine.PlayerID) engine.Command

// Submit runs a command as this client's player.
func (c *Client) Submit(build PlayerCommand) error {
	r := c.room
	return r.do(func() error {
		if !c.live() {
			return ErrDisconnected
		}
		cmd := build(c.player)
		if actor, ok := actorOf(cmd); ok && actor != c.player {
			return fmt.Errorf("%w: %T acts as %s", ErrForbidden, cmd, actor)
		}
		switch cmd := cmd.(type) {
		case engine.PhaseTimeout, engine.JoinRoom, engine.LeaveRoom:
			// Timeouts belong to the room; joining and leaving have their own calls.
			return fmt.Errorf("%w: %T", ErrForbidden, cmd)
		case engine.DealRoles:
			if err := r.hostOnly(c.player); err != nil {
				return err
			}
			cmd.Seed = r.rnd.Seed() // whatever the client sent is dropped
			return r.execute(cmd)
		case engine.SetSeats, engine.ConfigureRoles, engine.ConfigureTimers, engine.ConfigureSound:
			if err := r.hostOnly(c.player); err != nil {
				return err
			}
		case engine.NewGame:
			// Unlike hostOnly (gated to before the deal), NewGame only runs once
			// the game has ended, i.e. after the deal, so it checks the host
			// directly (audit K3).
			if c.player != r.host {
				return ErrNotHost
			}
		}
		return r.execute(cmd)
	})
}

// hostOnly guards the setup commands: only the host, and only before the deal.
func (r *Room) hostOnly(p engine.PlayerID) error {
	if r.state.Dealt {
		return fmt.Errorf("%w: setup is over once roles are dealt", engine.ErrWrongPhase)
	}
	if p != r.host {
		return ErrNotHost
	}
	return nil
}
