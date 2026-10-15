package engine

// FlagWolfPoisoned is set on a wolf poisoned by the Rusty knight. Its value is
// the number of the night of the bite; the wolf dies when the pack step of the
// next night starts (rules §6 "Hiệp sĩ kiếm gỉ").
const FlagWolfPoisoned = "wolf_poisoned_night"

// BiteAvenger is implemented by a role that reacts to being bitten by the pack
// (the Rusty knight). The pipeline calls it for the bitten player when the
// Guard did not shield them, also when the Witch heals them (rules §9 step 4).
type BiteAvenger interface {
	// AvengeBite returns the wolf to poison, or "" when there is none. s is the
	// state of the night being resolved; holder is the bitten player.
	AvengeBite(s State, holder PlayerID) PlayerID
}

// WolfPoisoned records the wolf poisoned by a bitten Rusty knight. Nobody sees
// it: the wolf learns of it when it dies (full views still see it).
type WolfPoisoned struct {
	EventMarker
	Wolf  PlayerID
	Night int
}

// Doomed reports whether the wolf was poisoned on an earlier night and is
// dead by the rules from the start of tonight's pack step: it cannot vote and
// the resolution kills it. The view keeps showing it alive until the
// morning announcement, so nobody learns of it in the night.
func Doomed(s State, id PlayerID) bool {
	p := s.Players[id]
	if p == nil || !p.Alive || s.Phase != PhaseNight {
		return false
	}
	night := p.Flags[FlagWolfPoisoned]
	return night > 0 && night < s.Night
}

// stepKnight (rules §9 step 4): a bitten player whose role is a BiteAvenger
// poisons a wolf. The Witch's heal does not stop it.
func stepKnight(r *Resolution) {
	if r.Bitten == "" || r.Protected {
		return
	}
	p := r.State.Players[r.Bitten]
	if p == nil {
		return
	}
	if avenger, ok := Registry[p.Role].(BiteAvenger); ok {
		r.KnightPoisoned = avenger.AvengeBite(r.State, r.Bitten)
	}
}

// stepDelayedWolfDeaths (rules §9 step 6): the wolves poisoned on an earlier
// night die.
func stepDelayedWolfDeaths(r *Resolution) {
	for _, id := range r.State.Seats {
		if Doomed(r.State, id) {
			r.addDeath(id)
		}
	}
}

func (r *Resolution) addDeath(id PlayerID) {
	for _, d := range r.Deaths {
		if d == id {
			return
		}
	}
	r.Deaths = append(r.Deaths, id)
}

// NearestLivingWolf walks from `from` in the direction of step (+1 is the
// left: the next seat, rules §6) and returns the first living wolf-team player
// who is not already doomed, never `from` itself. "" when there is none.
func NearestLivingWolf(s State, from PlayerID, step int) PlayerID {
	for id := NearestLiving(s, from, step); id != "" && id != from; id = NearestLiving(s, id, step) {
		if p := s.Players[id]; p.Team == TeamWolves && !Doomed(s, id) {
			return id
		}
	}
	return ""
}

func evolveWolfPoisoned(s State, e Event) (State, bool) {
	ev, ok := e.(WolfPoisoned)
	if !ok {
		return s, false
	}
	out := s.clone()
	if p := out.Players[ev.Wolf]; p != nil {
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		p.Flags[FlagWolfPoisoned] = ev.Night
	}
	return out, true
}
