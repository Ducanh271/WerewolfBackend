package engine

import "masoi/internal/i18n"

// Outcome is how a finished game ended (rules §8).
type Outcome struct {
	Winner  Team // TeamNone when Draw
	Draw    bool
	Winners []PlayerID // the players who win, dead or alive, in seating order; empty on a draw
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
	out.Winners = append([]PlayerID(nil), ev.Outcome.Winners...)
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
	var winner Team
	switch {
	case alive == 0:
		return Outcome{Draw: true}, true
	case crossed && alive == 2:
		return Outcome{Winner: TeamLovers, Winners: inSeatOrder(s.Seats, []PlayerID{lovA, lovB})}, true
	case wolvesAll == 0:
		winner = TeamVillage
	case wolves >= 1 && wolves >= villagers:
		winner = TeamWolves
	default:
		return Outcome{}, false
	}
	return Outcome{Winner: winner, Winners: teamWinners(s, winner, crossed)}, true
}

// teamWinners is every player of the winning team, dead or alive, except the
// lovers of a cross-team couple: once on the third team, they only win as the
// couple. That holds while the couple lives and after it broke as the third
// team (LoversLink.ThirdTeam): the wolf lover never shares a wolf win (rules §8).
func teamWinners(s State, winner Team, crossed bool) []PlayerID {
	thirdTeam := s.Lovers != nil && (crossed || s.Lovers.ThirdTeam)
	var out []PlayerID
	for _, id := range s.Seats {
		p := s.Players[id]
		if p == nil || p.Team != winner || (thirdTeam && PartnerOf(s, id) != "") {
			continue
		}
		out = append(out, id)
	}
	return out
}
