package engine

import (
	"fmt"
	"math/rand"
	"sort"
)

func decideDealRoles(s State, cmd DealRoles) ([]Event, error) {
	if s.Phase != PhaseLobby {
		return nil, fmt.Errorf("%w: cannot deal during %s", ErrWrongPhase, s.Phase)
	}
	if s.Dealt {
		return nil, ErrAlreadyDealt
	}
	if len(s.Settings.Roles) == 0 {
		return nil, ErrRolesNotConfigured
	}
	if len(s.Seats) != len(s.Players) {
		return nil, fmt.Errorf("%w: %d seats for %d players", ErrSeatsNotSet, len(s.Seats), len(s.Players))
	}
	// Players may have joined after ConfigureRoles; the list must still fit.
	if err := validateRoleConfig(len(s.Players), s.Settings.Roles); err != nil {
		return nil, err
	}
	return []Event{RolesDealt{Assignments: shuffleRoles(cmd.Seed, s.Seats, s.Settings.Roles)}}, nil
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
