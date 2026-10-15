package enginetest

import (
	"fmt"

	"masoi/internal/engine"
)

// Option configures NewGame.
type Option func(*setup)

type setup struct {
	seats  []string
	cast   map[string]engine.RoleID
	roles  map[engine.RoleID]int
	seed   int64
	chief  string
	driver Driver
	errs   []string
}

// Seats lists the players clockwise. A name is both the player's ID and
// nickname.
func Seats(names ...string) Option {
	return func(s *setup) { s.seats = append([]string(nil), names...) }
}

// Cast gives each named player a role: Cast("An", Guard, "Binh", Seer, ...).
// Every seated player must appear once. The cast goes through the same rules §2
// check as ConfigureRoles (player count, wolf limit, unique roles, Detective
// and Cupid minimums), so a scenario cannot build an illegal table.
func Cast(pairs ...any) Option {
	return func(s *setup) {
		if len(pairs)%2 != 0 {
			s.errs = append(s.errs, "Cast needs name, role pairs; got an odd number of arguments")
			return
		}
		s.cast = map[string]engine.RoleID{}
		for i := 0; i < len(pairs); i += 2 {
			name, okName := pairs[i].(string)
			role, okRole := pairs[i+1].(engine.RoleID)
			if !okName || !okRole {
				s.errs = append(s.errs, fmt.Sprintf("Cast argument %d: want name then role, got %T then %T", i/2+1, pairs[i], pairs[i+1]))
				return
			}
			if _, dup := s.cast[name]; dup {
				s.errs = append(s.errs, fmt.Sprintf("Cast gives %s two roles", name))
				return
			}
			s.cast[name] = role
		}
	}
}

// Roles picks the role list and lets DealRoles shuffle it with Seed. Use it
// for dealing tests; use Cast when a scenario needs to know who is what.
func Roles(counts map[engine.RoleID]int) Option {
	return func(s *setup) { s.roles = counts }
}

// Seed is the DealRoles seed used with Roles.
func Seed(seed int64) Option {
	return func(s *setup) { s.seed = seed }
}

// Chief makes the named player the Chief: in the election that the scenario
// does not play itself, everybody votes for them.
func Chief(name string) Option {
	return func(s *setup) { s.chief = name }
}

// WithDriver replaces the night/day driver.
func WithDriver(d Driver) Option {
	return func(s *setup) { s.driver = d }
}
