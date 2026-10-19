package engine

// TeamChanged moves a player to another team, e.g. a villager cursed into the
// pack (rules §5 "Sói nguyền"). Only that player is told; the pack sees its
// new member in the view. A team changed during the night shows on the phones
// only at the phone check (FlagTeamHidden), and the win check (rules §8) runs
// by day, so a team change itself ends nothing.
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
		if s.Phase == PhaseNight {
			SetFlag(p, FlagTeamHidden, 1)
		}
	}
	return out, true
}
