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

// doomIndex says whether the player carries a death that falls in a night
// after the one that sentenced them, and where in tonight's steps it falls.
// Two things sentence a player: the Rusty knight's poison (FlagWolfPoisoned)
// and the grief of a lover whose partner died by day (FlagGrieving).
//
// index is the position, in tonight's steps, of the step where the player dies:
//   - the poison takes hold when the pack step starts;
//   - grief takes hold at the first step that calls the player: their own role's
//     step, or the pack step for anybody on the wolf team;
//   - a player no step calls (Villager, Hunter, Knight...) has index -1: they
//     die in the night resolution, at step 7 of rules §9.
//
// A player with both takes the earlier one.
func doomIndex(s State, p *Player) (index int, doomed bool) {
	poisoned := p.Flags[FlagWolfPoisoned] > 0 && p.Flags[FlagWolfPoisoned] < s.Night
	grieving := p.Flags[FlagGrieving] > 0 && p.Flags[FlagGrieving] < s.Night
	if !poisoned && !grieving {
		return 0, false
	}
	index = -1
	for i, step := range nightSteps(s, s.Night) {
		own := grieving && (step == p.Role || (p.Team == TeamWolves && isPackStep(step)))
		if own || (poisoned && isPackStep(step)) {
			index = i
			break
		}
	}
	return index, true
}

// Doomed reports whether the player is dead by the rules from the step now
// running: they cannot act or vote, and the resolution kills them. The view
// keeps showing them alive until the morning announcement, so nobody learns of
// it in the night.
func Doomed(s State, id PlayerID) bool {
	p := s.Players[id]
	if p == nil || !p.Alive || s.Phase != PhaseNight {
		return false
	}
	index, doomed := doomIndex(s, p)
	return doomed && index >= 0 && s.NightStep >= index+1
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

// stepDelayedDeaths (rules §9 step 6): the players whose death took hold at a
// step of this night (the poisoned wolf, a grieving lover with a night call)
// die. Of the law they died when that step started; the death is announced
// with the others at step 8.
func stepDelayedDeaths(r *Resolution) {
	for _, id := range r.State.Seats {
		if Doomed(r.State, id) {
			r.addDeath(id)
		}
	}
}

// stepGriefWithoutCall (rules §9 step 7): a grieving lover whom no night step
// calls dies now.
func stepGriefWithoutCall(r *Resolution) {
	for _, id := range r.State.Seats {
		if p := r.State.Players[id]; p != nil && p.Alive {
			if index, doomed := doomIndex(r.State, p); doomed && index < 0 {
				r.addDeath(id)
			}
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
