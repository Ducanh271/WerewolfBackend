package engine

// SetFlag sets one of the player's flags, creating the map on first use.
func SetFlag(p *Player, key string, value int) {
	if p.Flags == nil {
		p.Flags = map[string]int{}
	}
	p.Flags[key] = value
}

// LastTarget keeps, in two of a player's flags, the night of their last choice
// and whom it was (seat number, see SeatNumber). It carries the rule "not the
// same player two nights in a row" (rules §6 "Bảo vệ", "Nguyệt Nữ"); a night
// without a choice breaks the streak.
type LastTarget struct {
	Night  string // flag holding the night of the last choice
	Target string // flag holding the seat number of the last target
}

// Is reports whether target is the one p chose the night before tonight.
func (l LastTarget) Is(s State, p *Player, target PlayerID) bool {
	return p.Flags[l.Night] == s.Night-1 && p.Flags[l.Target] == SeatNumber(s, target)
}

// Record stores p's choice of target on that night.
func (l LastTarget) Record(s State, p *Player, night int, target PlayerID) {
	SetFlag(p, l.Night, night)
	SetFlag(p, l.Target, SeatNumber(s, target))
}
