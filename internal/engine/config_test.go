package engine

import (
	"errors"
	"fmt"
	"testing"
)

// withVillagers fills the role list up to playerCount with villagers, so each
// test only states the roles it cares about.
func withVillagers(playerCount int, roles map[RoleID]int) map[RoleID]int {
	out := copyRoles(roles)
	if out == nil {
		out = map[RoleID]int{}
	}
	slots := 0
	for _, n := range out {
		slots += n
	}
	if rest := playerCount - slots; rest > 0 {
		out[roleVillager] += rest
	}
	return out
}

// configure runs ConfigureRoles on a lobby of playerCount players.
func configure(t *testing.T, playerCount int, roles map[RoleID]int) error {
	t.Helper()
	_, err := Decide(lobbyOf(t, playerCount), ConfigureRoles{Roles: roles})
	return err
}

func TestConfigureRoles_StoresTheRoleList(t *testing.T) {
	s := lobbyOf(t, 7)
	roles := withVillagers(7, map[RoleID]int{roleWerewolf: 2, roleSeer: 1})

	s = apply(t, s, ConfigureRoles{Roles: roles})

	if len(s.Settings.Roles) != 3 || s.Settings.Roles[roleWerewolf] != 2 || s.Settings.Roles[roleVillager] != 4 {
		t.Fatalf("roles = %v, want 2 wolves, 1 seer, 4 villagers", s.Settings.Roles)
	}

	// The event must own its map.
	roles[roleWerewolf] = 99
	if s.Settings.Roles[roleWerewolf] != 2 {
		t.Errorf("settings aliased the command map: %v", s.Settings.Roles)
	}
}

// TestConfigureRoles_WolfLimitFollowsRulesTable covers the rules §2 limit:
// total wolves must not exceed 30% of the players, rounded down.
func TestConfigureRoles_WolfLimitFollowsRulesTable(t *testing.T) {
	tests := []struct {
		players   int
		maxWolves int
	}{
		{5, 1},  // table row 5-6
		{6, 1},  // table row 5-6
		{7, 2},  // table row 7-9
		{9, 2},  // table row 7-9
		{10, 3}, // table row 10-13
		{13, 3}, // table row 10-13
		{14, 4}, // table row 14-16
		{16, 4}, // table row 14-16
		{17, 5}, // not listed in the table; 30% rounded down
		{20, 6}, // table row 20
		{30, 9}, // table row 30
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d players", tt.players), func(t *testing.T) {
			if got := MaxWolves(tt.players); got != tt.maxWolves {
				t.Fatalf("MaxWolves(%d) = %d, want %d", tt.players, got, tt.maxWolves)
			}

			atLimit := withVillagers(tt.players, map[RoleID]int{roleWerewolf: tt.maxWolves})
			if err := configure(t, tt.players, atLimit); err != nil {
				t.Errorf("%d wolves at %d players = %v, want accepted", tt.maxWolves, tt.players, err)
			}

			overLimit := withVillagers(tt.players, map[RoleID]int{roleWerewolf: tt.maxWolves + 1})
			if err := configure(t, tt.players, overLimit); !errors.Is(err, ErrTooManyWolves) {
				t.Errorf("%d wolves at %d players = %v, want %v",
					tt.maxWolves+1, tt.players, err, ErrTooManyWolves)
			}
		})
	}
}

// TestConfigureRoles_CursedWolfReservesAWolfSlot covers the rules §2 wording
// "tính cả người sẽ bị nguyền": the villager the cursed wolf will turn counts
// towards the limit, so the cursed option always deals one wolf fewer.
func TestConfigureRoles_CursedWolfReservesAWolfSlot(t *testing.T) {
	tests := []struct {
		name    string
		players int
		roles   map[RoleID]int
		want    error // nil = accepted
	}{
		{"5 players: cursed option does not fit", 5, map[RoleID]int{roleCursedWolf: 1}, ErrTooManyWolves},
		{"7 players: 1 cursed wolf", 7, map[RoleID]int{roleCursedWolf: 1}, nil},
		{"7 players: 1 wolf + 1 cursed wolf is too many", 7, map[RoleID]int{roleWerewolf: 1, roleCursedWolf: 1}, ErrTooManyWolves},
		{"10 players: 1 wolf + 1 cursed wolf", 10, map[RoleID]int{roleWerewolf: 1, roleCursedWolf: 1}, nil},
		{"10 players: 2 wolves + 1 cursed wolf is too many", 10, map[RoleID]int{roleWerewolf: 2, roleCursedWolf: 1}, ErrTooManyWolves},
		{"14 players: 2 wolves + 1 cursed wolf", 14, map[RoleID]int{roleWerewolf: 2, roleCursedWolf: 1}, nil},
		{"14 players: 3 wolves + 1 cursed wolf is too many", 14, map[RoleID]int{roleWerewolf: 3, roleCursedWolf: 1}, ErrTooManyWolves},
		{"20 players: 4 wolves + 1 cursed wolf", 20, map[RoleID]int{roleWerewolf: 4, roleCursedWolf: 1}, nil},
		{"20 players: 5 wolves + 1 cursed wolf is too many", 20, map[RoleID]int{roleWerewolf: 5, roleCursedWolf: 1}, ErrTooManyWolves},
		{"30 players: 7 wolves + 1 cursed wolf", 30, map[RoleID]int{roleWerewolf: 7, roleCursedWolf: 1}, nil},
		{"30 players: 8 wolves + 1 cursed wolf is too many", 30, map[RoleID]int{roleWerewolf: 8, roleCursedWolf: 1}, ErrTooManyWolves},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configure(t, tt.players, withVillagers(tt.players, tt.roles))
			if tt.want == nil {
				if err != nil {
					t.Fatalf("got %v, want accepted", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

// TestConfigureRoles_DetectiveNeedsMoreThanEightPlayers covers rules §2:
// "Thám tử — Ván trên 8 người".
func TestConfigureRoles_DetectiveNeedsMoreThanEightPlayers(t *testing.T) {
	for _, players := range []int{5, 8} {
		roles := withVillagers(players, map[RoleID]int{roleWerewolf: 1, roleDetective: 1})
		if err := configure(t, players, roles); !errors.Is(err, ErrRoleNeedsMorePlayers) {
			t.Errorf("detective at %d players = %v, want %v", players, err, ErrRoleNeedsMorePlayers)
		}
	}

	roles := withVillagers(9, map[RoleID]int{roleWerewolf: 2, roleDetective: 1})
	if err := configure(t, 9, roles); err != nil {
		t.Errorf("detective at 9 players = %v, want accepted", err)
	}
}

// TestConfigureRoles_CupidNeedsTenPlayers covers rules §2:
// "Cupid — Ván từ 10 người".
func TestConfigureRoles_CupidNeedsTenPlayers(t *testing.T) {
	for _, players := range []int{5, 9} {
		roles := withVillagers(players, map[RoleID]int{roleWerewolf: 1, roleCupid: 1})
		if err := configure(t, players, roles); !errors.Is(err, ErrRoleNeedsMorePlayers) {
			t.Errorf("cupid at %d players = %v, want %v", players, err, ErrRoleNeedsMorePlayers)
		}
	}

	roles := withVillagers(10, map[RoleID]int{roleWerewolf: 3, roleCupid: 1})
	if err := configure(t, 10, roles); err != nil {
		t.Errorf("cupid at 10 players = %v, want accepted", err)
	}
}

// TestConfigureRoles_OnlyVillagersAndWolvesCanRepeat covers rules §2 "Số lượng
// mỗi vai": villagers and plain wolves can be several, every other role is
// unique.
func TestConfigureRoles_OnlyVillagersAndWolvesCanRepeat(t *testing.T) {
	tests := []struct {
		name    string
		players int
		roles   map[RoleID]int
		want    error // nil = accepted
	}{
		{"several villagers", 10, map[RoleID]int{roleWerewolf: 1, roleVillager: 9}, nil},
		{"several wolves", 10, map[RoleID]int{roleWerewolf: 3, roleVillager: 7}, nil},
		{"two seers", 10, map[RoleID]int{roleWerewolf: 3, roleSeer: 2, roleVillager: 5}, ErrRoleMustBeUnique},
		{"one seer", 10, map[RoleID]int{roleWerewolf: 3, roleSeer: 1, roleVillager: 6}, nil},
		{"two cursed wolves", 20, map[RoleID]int{roleWerewolf: 3, roleCursedWolf: 2, roleVillager: 15}, ErrRoleMustBeUnique},
		{"two detectives", 10, map[RoleID]int{roleWerewolf: 3, roleDetective: 2, roleVillager: 5}, ErrRoleMustBeUnique},
		{"two cupids", 10, map[RoleID]int{roleWerewolf: 3, roleCupid: 2, roleVillager: 5}, ErrRoleMustBeUnique},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configure(t, tt.players, tt.roles)
			if tt.want == nil {
				if err != nil {
					t.Fatalf("got %v, want accepted", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

// TestConfigureRoles_PlayerCountLimits covers rules §2: "Một ván có 5-30
// người".
func TestConfigureRoles_PlayerCountLimits(t *testing.T) {
	tests := []struct {
		name    string
		players int
		want    error
	}{
		{"empty lobby", 0, ErrTooFewPlayers},
		{"four players", 4, ErrTooFewPlayers},
		{"five players", 5, nil},
		{"thirty players", MaxPlayers, nil},
		{"thirty-one players", MaxPlayers + 1, ErrTooManyPlayers},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A lobby cannot hold more than MaxPlayers players, so the
			// limit is checked against the validator directly.
			err := validateRoleConfig(tt.players, withVillagers(tt.players, map[RoleID]int{roleWerewolf: 1}))
			if tt.want == nil {
				if err != nil {
					t.Fatalf("got %v, want accepted", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestConfigureRoles_Rejects(t *testing.T) {
	tests := []struct {
		name    string
		players int
		roles   map[RoleID]int
		want    error
	}{
		{"unknown role", 7, map[RoleID]int{roleWerewolf: 1, "Werehamster": 6}, ErrUnknownRole},
		{"zero count", 7, map[RoleID]int{roleWerewolf: 1, roleSeer: 0, roleVillager: 6}, ErrInvalidRoleCount},
		{"negative count", 7, map[RoleID]int{roleWerewolf: 1, roleSeer: -1, roleVillager: 7}, ErrInvalidRoleCount},
		{"too few slots", 7, map[RoleID]int{roleWerewolf: 1, roleVillager: 4}, ErrRoleCountMismatch},
		{"too many slots", 7, map[RoleID]int{roleWerewolf: 1, roleVillager: 9}, ErrRoleCountMismatch},
		{"no roles at all", 7, nil, ErrRoleCountMismatch},
		{"no wolves", 7, map[RoleID]int{roleSeer: 1, roleVillager: 6}, ErrNoWolves},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := configure(t, tt.players, tt.roles); !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestConfigureRoles_RejectedOutsideLobby(t *testing.T) {
	roles := withVillagers(5, map[RoleID]int{roleWerewolf: 1})

	if _, err := Decide(nightState(t), ConfigureRoles{Roles: roles}); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("got %v, want %v", err, ErrWrongPhase)
	}
}
