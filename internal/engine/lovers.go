package engine

// FlagLoveKnown is set on a lover once the phone check has told them of the
// pair; until then their view shows nothing about it (rules §3, §7).
const FlagLoveKnown = "love.known"

// FlagGrieving is set on the surviving lover of a pair broken by day. Its value
// is the number of the night the day belongs to. The survivor lives on through
// the day, then dies in the next night (rules §7, §9); see doomIndex.
const FlagGrieving = "love.grieving_night"

// LoveSource is implemented by the role that creates the pair (Cupid). The
// phone check runs on the nights it is called, so the lovers can learn of it.
type LoveSource interface {
	CreatesLovers() bool
}

// LoverGrieving records that the partner of a lover who died by day will die
// in the next night (rules §7). Nobody is told: the survivor already knows their lover is
// dead, and the others see them alive until the morning announcement.
type LoverGrieving struct {
	EventMarker
	Player PlayerID
	Night  int
}

// PartnerOf returns the other lover of the pair, "" when id is not in it.
func PartnerOf(s State, id PlayerID) PlayerID {
	switch {
	case s.Lovers == nil:
		return ""
	case id == s.Lovers.A:
		return s.Lovers.B
	case id == s.Lovers.B:
		return s.Lovers.A
	}
	return ""
}

// griefMarks is what the death of a lover by day does (rules §7): the
// surviving partner is marked to die in the next night. dead is everybody who died in the chain;
// after is the state once the chain is over, so a partner who died in it too
// is not marked.
func griefMarks(after State, dead []PlayerID) []Event {
	var evs []Event
	marked := map[PlayerID]bool{}
	for _, id := range dead {
		partner := PartnerOf(after, id)
		p := after.Players[partner]
		if partner == "" || p == nil || !p.Alive || marked[partner] || p.Flags[FlagGrieving] != 0 {
			continue
		}
		marked[partner] = true
		evs = append(evs, LoverGrieving{Player: partner, Night: after.Night})
	}
	return evs
}

// stepLoverDeaths (rules §9 step 7): the partner of everybody who died in the
// night dies too, whatever the cause.
func stepLoverDeaths(r *Resolution) {
	for _, id := range append([]PlayerID(nil), r.Deaths...) {
		if partner := PartnerOf(r.State, id); partner != "" && r.State.Players[partner].Alive {
			r.addDeath(partner)
		}
	}
}

// coupleColour is the colour of the link for a pair on those teams.
func coupleColour(a, b Team) LinkColour {
	switch {
	case a == TeamLovers || b == TeamLovers || a != b:
		return LinkStriped
	case a == TeamWolves:
		return LinkBlack
	default:
		return LinkWhite
	}
}

func evolveLovers(s State, e Event) (State, bool) {
	ev, ok := e.(LoverGrieving)
	if !ok {
		return s, false
	}
	out := s.clone()
	if p := out.Players[ev.Player]; p != nil {
		SetFlag(p, FlagGrieving, ev.Night)
	}
	return out, true
}
