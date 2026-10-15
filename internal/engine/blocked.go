package engine

// Blocked reports whether the player lost their night ability tonight, whoever
// blocked them (rules §3). Blocking is a night-only state: the day abilities
// (the Hunter's shot) never look at it.
func Blocked(s State, id PlayerID) bool {
	_, blocked := s.Pending.Blocked[id]
	return blocked
}

// BlockedBy reports whether the player is blocked tonight by that specific
// role. For roles of the engine only: it is never shown to any player.
func BlockedBy(s State, id PlayerID, source RoleID) bool {
	by, blocked := s.Pending.Blocked[id]
	return blocked && by == source
}

// Block records, in Evolve of a blocking role, that target loses their night
// ability tonight. The night start clears it (rules §9 step 1).
func Block(s *State, target PlayerID, source RoleID) {
	if s.Pending.Blocked == nil {
		s.Pending.Blocked = map[PlayerID]RoleID{}
	}
	s.Pending.Blocked[target] = source
}
