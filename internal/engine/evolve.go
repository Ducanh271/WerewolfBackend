package engine

// Evolve applies one event and returns the resulting state. It never fails:
// an event nobody claims leaves the state unchanged.
func Evolve(s State, e Event) State {
	switch ev := e.(type) {
	case PlayerJoined:
		out := s.clone()
		out.Players[ev.Player] = &Player{
			ID:       ev.Player,
			Nickname: ev.Nickname,
			Team:     TeamNone,
			Alive:    true,
			Flags:    map[string]int{},
		}
		return out

	case PlayerLeft:
		return evolvePlayerRemoved(s, ev.Player)

	case PlayerKicked:
		return evolvePlayerRemoved(s, ev.Player)

	case TimersConfigured:
		out := s.clone()
		out.Settings = applyTimers(out.Settings, ev.Timers)
		return out

	case SeatsSet:
		out := s.clone()
		out.Seats = append([]PlayerID(nil), ev.Order...)
		return out

	case RolesConfigured:
		out := s.clone()
		out.Settings.Roles = copyRoles(ev.Roles)
		return out

	case SoundConfigured:
		out := s.clone()
		out.Settings.Sound = ev.Mode
		return out

	case RolesDealt:
		out := s.clone()
		for id, role := range ev.Assignments {
			p, ok := out.Players[id]
			if !ok {
				continue
			}
			p.Role = role
			p.Team = Registry[role].BaseTeam()
		}
		out.Dealt = true
		out.Phase = PhaseRoleView
		out.Seq++
		return out
	}

	if out, ok := evolvePhoneCheck(s, e); ok {
		return out
	}
	if out, ok := evolveLovers(s, e); ok {
		return out
	}
	if out, ok := evolveTeam(s, e); ok {
		return out
	}
	if out, ok := evolveWin(s, e); ok {
		return out
	}
	if out, ok := evolveChief(s, e); ok {
		return out
	}
	if out, ok := evolveHunter(s, e); ok {
		return out
	}
	if out, ok := evolveDay(s, e); ok {
		return out
	}
	if out, ok := evolveNewGame(s, e); ok {
		return out
	}
	if out, ok := evolveNight(s, e); ok {
		return out
	}

	// Role events. An event type belongs to exactly one role, so the random
	// map order does not matter.
	out := s.clone()
	for _, role := range Registry {
		if role.Evolve(&out, e) {
			return out
		}
	}
	return s
}

// evolvePlayerRemoved takes a player out of the lobby: only their own seat is
// freed, the rest keep theirs in order, and the chosen role list is left
// alone (PlayerLeft, PlayerKicked).
func evolvePlayerRemoved(s State, id PlayerID) State {
	out := s.clone()
	delete(out.Players, id)
	for i, p := range out.Seats {
		if p == id {
			out.Seats = append(out.Seats[:i], out.Seats[i+1:]...)
			break
		}
	}
	return out
}

// Fold rebuilds a state by applying events in order.
func Fold(s State, events ...Event) State {
	for _, e := range events {
		s = Evolve(s, e)
	}
	return s
}
