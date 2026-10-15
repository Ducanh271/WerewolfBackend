package engine

// TeamChanged moves a player to another team, e.g. a villager cursed into the
// pack (rules §5 "Sói nguyền"). Only that player is told; the pack sees its
// new member in the view. The win check after a team change (rules §8) is the
// job of whoever emits it.
type TeamChanged struct {
	EventMarker
	Player PlayerID
	Team   Team
}

func evolveTeam(s State, e Event) (State, bool) {
	ev, ok := e.(TeamChanged)
	if !ok {
		return s, false
	}
	out := s.clone()
	if p := out.Players[ev.Player]; p != nil {
		p.Team = ev.Team
	}
	return out, true
}
