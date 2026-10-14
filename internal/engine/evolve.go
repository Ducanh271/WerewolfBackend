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

	case SeatsSet:
		out := s.clone()
		out.Seats = append([]PlayerID(nil), ev.Order...)
		return out

	case RolesConfigured:
		out := s.clone()
		out.Settings.Roles = copyRoles(ev.Roles)
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

// Fold rebuilds a state by applying events in order.
func Fold(s State, events ...Event) State {
	for _, e := range events {
		s = Evolve(s, e)
	}
	return s
}
