package engine

import (
	"fmt"
	"strings"
)

// Decide validates a command against the current state and returns the events
// it produces. It has no side effects and never panics on player input.
func Decide(s State, c Command) ([]Event, error) {
	switch cmd := c.(type) {
	case JoinRoom:
		return decideJoinRoom(s, cmd)
	case LeaveRoom:
		return decideLeaveRoom(s, cmd)
	case ConfigureTimers:
		return decideConfigureTimers(s, cmd)
	case SetSeats:
		return decideSetSeats(s, cmd)
	case ConfigureRoles:
		return decideConfigureRoles(s, cmd)
	case DealRoles:
		return decideDealRoles(s, cmd)
	case VoteChief:
		return decideVoteChief(s, cmd)
	case NameSuccessor:
		return decideNameSuccessor(s, cmd)
	case Nominate:
		return decideNominate(s, cmd)
	case SkipDay:
		return decideSkipDay(s, cmd)
	case CastVerdict:
		return decideCastVerdict(s, cmd)
	case PhaseTimeout:
		return decidePhaseTimeout(s, cmd)
	}

	// Role commands. A command type belongs to exactly one role, so the
	// random map order does not matter.
	for _, role := range Registry {
		evs, handled, err := role.Decide(s, c)
		if handled {
			return evs, err
		}
	}
	return nil, fmt.Errorf("%w: %T", ErrUnknownCommand, c)
}

func decideJoinRoom(s State, cmd JoinRoom) ([]Event, error) {
	if s.Phase != PhaseLobby {
		return nil, fmt.Errorf("%w: cannot join during %s", ErrWrongPhase, s.Phase)
	}
	if cmd.Player == "" {
		return nil, ErrInvalidPlayerID
	}
	nickname := strings.TrimSpace(cmd.Nickname)
	if nickname == "" {
		return nil, fmt.Errorf("%w: nickname is empty", ErrInvalidNickname)
	}
	if _, joined := s.Players[cmd.Player]; joined {
		return nil, fmt.Errorf("%w: %s", ErrAlreadyJoined, cmd.Player)
	}
	if len(s.Players) >= MaxPlayers {
		return nil, fmt.Errorf("%w: %d players", ErrRoomFull, len(s.Players))
	}
	for _, p := range s.Players {
		if strings.EqualFold(p.Nickname, nickname) {
			return nil, fmt.Errorf("%w: %q", ErrNicknameTaken, nickname)
		}
	}
	return []Event{PlayerJoined{Player: cmd.Player, Nickname: nickname}}, nil
}

func decideLeaveRoom(s State, cmd LeaveRoom) ([]Event, error) {
	if s.Phase != PhaseLobby || s.Dealt {
		return nil, fmt.Errorf("%w: cannot leave during %s", ErrWrongPhase, s.Phase)
	}
	if _, joined := s.Players[cmd.Player]; !joined {
		return nil, fmt.Errorf("%w: %s", ErrUnknownPlayer, cmd.Player)
	}
	return []Event{PlayerLeft{Player: cmd.Player}}, nil
}

func decideSetSeats(s State, cmd SetSeats) ([]Event, error) {
	if s.Phase != PhaseLobby {
		return nil, fmt.Errorf("%w: cannot seat players during %s", ErrWrongPhase, s.Phase)
	}
	if len(cmd.Order) != len(s.Players) {
		return nil, fmt.Errorf("%w: %d seats for %d players", ErrSeatsMismatch, len(cmd.Order), len(s.Players))
	}
	seen := make(map[PlayerID]bool, len(cmd.Order))
	for _, id := range cmd.Order {
		if _, joined := s.Players[id]; !joined {
			return nil, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
		}
		if seen[id] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateSeat, id)
		}
		seen[id] = true
	}
	return []Event{SeatsSet{Order: append([]PlayerID(nil), cmd.Order...)}}, nil
}

func decideConfigureRoles(s State, cmd ConfigureRoles) ([]Event, error) {
	if s.Phase != PhaseLobby {
		return nil, fmt.Errorf("%w: cannot configure roles during %s", ErrWrongPhase, s.Phase)
	}
	if err := validateRoleConfig(len(s.Players), cmd.Roles); err != nil {
		return nil, err
	}
	return []Event{RolesConfigured{Roles: copyRoles(cmd.Roles)}}, nil
}
