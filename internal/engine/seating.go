package engine

// NearestLiving returns the nearest living player from `from` in seating
// order, skipping the dead and never returning `from` itself. step is +1 or -1
// (the two sides of the circle). It returns "" when nobody else is alive, or
// when `from` is not seated. `from` may be dead: the Rusty knight starts from
// his own seat after the bite (rules §6).
func NearestLiving(s State, from PlayerID, step int) PlayerID {
	n := len(s.Seats)
	start := -1
	for i, id := range s.Seats {
		if id == from {
			start = i
			break
		}
	}
	if start < 0 || (step != 1 && step != -1) {
		return ""
	}
	for k := 1; k < n; k++ {
		id := s.Seats[((start+step*k)%n+n)%n]
		if p, ok := s.Players[id]; ok && p.Alive {
			return id
		}
	}
	return ""
}
