package engine

import "masoi/internal/i18n"

// Outcome is how a finished game ended (rules §8).
type Outcome struct {
	Winner Team // TeamNone when Draw
	Draw   bool
}

// GameEnded closes the game. A draw has no winner.
type GameEnded struct {
	EventMarker
	Outcome Outcome
}

// AnnouncementKey is read aloud to everybody.
func (e GameEnded) AnnouncementKey() string {
	switch {
	case e.Outcome.Draw:
		return i18n.WinDraw
	case e.Outcome.Winner == TeamLovers:
		return i18n.WinLovers
	case e.Outcome.Winner == TeamVillage:
		return i18n.WinVillage
	default:
		return i18n.WinWolves
	}
}

func evolveWin(s State, e Event) (State, bool) {
	ev, ok := e.(GameEnded)
	if !ok {
		return s, false
	}
	out := s.clone()
	if !ev.Outcome.Draw {
		winner := ev.Outcome.Winner
		out.Winner = &winner
	}
	out.Phase = PhaseEnded
	out.Seq++
	return out, true
}

// crossTeamLovers returns the two lovers while they are a couple of one wolf
// and one villager: the third team. Both must be alive.
func crossTeamLovers(s State) (a, b PlayerID, ok bool) {
	if s.Lovers == nil {
		return "", "", false
	}
	pa, pb := s.Players[s.Lovers.A], s.Players[s.Lovers.B]
	if pa == nil || pb == nil || !pa.Alive || !pb.Alive {
		return "", "", false
	}
	if (pa.Team == TeamWolves) == (pb.Team == TeamWolves) {
		return "", "", false
	}
	return pa.ID, pb.ID, true
}

// CheckWin applies rules §8. Priority when several hold: draw, then the
// cross-team lovers, then the village, then the wolves.
func CheckWin(s State) (Outcome, bool) {
	lovA, lovB, crossed := crossTeamLovers(s)
	alive, wolvesAll, wolves, villagers := 0, 0, 0, 0
	for id, p := range s.Players {
		if !p.Alive {
			continue
		}
		alive++
		if p.Team == TeamWolves {
			wolvesAll++
		}
		if crossed && (id == lovA || id == lovB) {
			continue // the couple counts on neither side
		}
		switch p.Team {
		case TeamWolves:
			wolves++
		case TeamVillage:
			villagers++
		}
	}
	switch {
	case alive == 0:
		return Outcome{Draw: true}, true
	case crossed && alive == 2:
		return Outcome{Winner: TeamLovers}, true
	case wolvesAll == 0:
		return Outcome{Winner: TeamVillage}, true
	case wolves >= 1 && wolves >= villagers:
		return Outcome{Winner: TeamWolves}, true
	}
	return Outcome{}, false
}
