package engine

// SeatNumber is the player's seat index + 1, so 0 means not seated. Flags keep
// players by seat number since they hold ints.
func SeatNumber(s State, id PlayerID) int {
	for i, seat := range s.Seats {
		if seat == id {
			return i + 1
		}
	}
	return 0
}

// HolderOf is the seated player holding the role, alive or dead; nil when
// nobody does. For roles held by one player only.
func HolderOf(s State, role RoleID) *Player {
	for _, id := range s.Seats {
		if p := s.Players[id]; p != nil && p.Role == role {
			return p
		}
	}
	return nil
}

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
