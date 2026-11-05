package engine

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"masoi/internal/i18n"
)

// RoleViewStarted opens the role-view countdown right after the deal (rules
// §2 bước 5): every player looks at their own role before anything else.
type RoleViewStarted struct {
	EventMarker
	Duration time.Duration
}

// AnnouncementKey is read aloud to everybody.
func (RoleViewStarted) AnnouncementKey() string { return i18n.RoleViewStarted }

func decideDealRoles(s State, cmd DealRoles) ([]Event, error) {
	if s.Dealt {
		return nil, ErrAlreadyDealt // the deal is final: the host's powers end with it
	}
	if s.Phase != PhaseLobby {
		return nil, fmt.Errorf("%w: cannot deal during %s", ErrWrongPhase, s.Phase)
	}
	if len(s.Settings.Roles) == 0 {
		return nil, ErrRolesNotConfigured
	}
	online := s.Settings.Mode == Online
	if !online && len(s.Seats) != len(s.Players) {
		return nil, fmt.Errorf("%w: %d seats for %d players", ErrSeatsNotSet, len(s.Seats), len(s.Players))
	}
	// Players may have joined after ConfigureRoles; the list must still fit.
	if err := validateRoleConfig(len(s.Players), s.Settings.Roles); err != nil {
		return nil, err
	}
	seats, seed := s.Seats, cmd.Seed
	var evs []Event
	if online {
		seats, seed = drawSeats(cmd.Seed, s.Players)
		evs = append(evs, SeatsSet{Order: seats})
	}
	return append(evs,
		RolesDealt{Assignments: shuffleRoles(seed, seats, s.Settings.Roles)},
		RoleViewStarted{Duration: s.Settings.RoleView},
	), nil
}

// drawSeats draws an online room's circle (rules §1b). It also returns the
// seed for the role shuffle, taken from the same generator: shuffling seats
// and roles with one seed would apply one permutation to both, and each
// player's role would then follow from the order they joined in.
func drawSeats(seed int64, players map[PlayerID]*Player) ([]PlayerID, int64) {
	seats := make([]PlayerID, 0, len(players))
	for id := range players {
		seats = append(seats, id)
	}
	sort.Slice(seats, func(i, j int) bool { return seats[i] < seats[j] })
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(seats), func(i, j int) { seats[i], seats[j] = seats[j], seats[i] })
	return seats, rng.Int63()
}

// shuffleRoles deals roles to seats. The role list is expanded in ID order
// (map order is random) and shuffled with a generator owned by this call, so
// the result depends only on its arguments.
func shuffleRoles(seed int64, seats []PlayerID, roles map[RoleID]int) map[PlayerID]RoleID {
	ids := make([]RoleID, 0, len(roles))
	for id := range roles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	deck := make([]RoleID, 0, len(seats))
	for _, id := range ids {
		for i := 0; i < roles[id]; i++ {
			deck = append(deck, id)
		}
	}

	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	out := make(map[PlayerID]RoleID, len(seats))
	for i, id := range seats {
		out[id] = deck[i]
	}
	return out
}
