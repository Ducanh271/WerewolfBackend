package engine

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// tableOf returns a seated, configured lobby of n players: the wolves, the
// optional cursed wolf, one Seer when it fits, and villagers for the rest.
func tableOf(t *testing.T, n, wolves int, cursed bool) State {
	t.Helper()
	roles := map[RoleID]int{}
	used := 0
	if wolves > 0 {
		roles[roleWerewolf] = wolves
		used += wolves
	}
	if cursed {
		roles[roleCursedWolf] = 1
		used++
	}
	roles[roleSeer] = 1
	used++
	roles[roleVillager] = n - used

	s := lobbyOf(t, n)
	s = apply(t, s, SetSeats{Order: seatOrder(n)})
	return apply(t, s, ConfigureRoles{Roles: roles})
}

func dealtRoles(s State) map[PlayerID]RoleID {
	out := map[PlayerID]RoleID{}
	for id, p := range s.Players {
		out[id] = p.Role
	}
	return out
}

func TestDealRoles_SameSeedGivesSameDeal(t *testing.T) {
	base := tableOf(t, 10, 1, true)

	a := dealtRoles(apply(t, base, DealRoles{Seed: 42}))
	b := dealtRoles(apply(t, base, DealRoles{Seed: 42}))
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed gave different deals:\n%v\n%v", a, b)
	}

	different := false
	for seed := int64(1); seed <= 20; seed++ {
		if !reflect.DeepEqual(a, dealtRoles(apply(t, base, DealRoles{Seed: 1000 + seed}))) {
			different = true
			break
		}
	}
	if !different {
		t.Fatal("20 other seeds all gave the same deal; shuffle is not using the seed")
	}
}

func TestDealRoles_WolfCountMatchesRulesTable(t *testing.T) {
	tests := []struct {
		players     int
		plainWolves int // "phương án sói thường": all wolves are Werewolf cards
		cursedPlain int // Werewolf cards next to the CursedWolf in the cursed option
		cursedTotal int // all wolf-team cards dealt in the cursed option
	}{
		// Tổng sói tối đa is 1, 2, 3, 4, 6, 9; the cursed option deals one
		// fewer wolf card because the villager to be cursed fills the last slot.
		{5, 1, -1, -1}, // cursed option does not apply at 5-6
		{7, 2, 0, 1},   // CursedWolf only
		{10, 3, 1, 2},  // 1 Werewolf + 1 CursedWolf
		{14, 4, 2, 3},
		{20, 6, 4, 5},
		{30, 9, 7, 8},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("plain_%d_players", tt.players), func(t *testing.T) {
			s := apply(t, tableOf(t, tt.players, tt.plainWolves, false), DealRoles{Seed: 7})
			if got := countWolves(s); got != tt.plainWolves {
				t.Fatalf("wolves = %d, want %d", got, tt.plainWolves)
			}
		})
		if tt.cursedPlain < 0 {
			continue
		}
		t.Run(fmt.Sprintf("cursed_%d_players", tt.players), func(t *testing.T) {
			s := apply(t, tableOf(t, tt.players, tt.cursedPlain, true), DealRoles{Seed: 7})
			if got := countWolves(s); got != tt.cursedTotal {
				t.Fatalf("wolf cards = %d, want %d", got, tt.cursedTotal)
			}
			cursed := 0
			for _, p := range s.Players {
				if p.Role == roleCursedWolf {
					cursed++
				}
			}
			if cursed != 1 {
				t.Fatalf("CursedWolf cards = %d, want 1", cursed)
			}
		})
	}
}

func countWolves(s State) int {
	n := 0
	for _, p := range s.Players {
		if p.Team == TeamWolves {
			n++
		}
	}
	return n
}

func TestDealRoles_EveryRoleDealtAsConfiguredWithTeams(t *testing.T) {
	s := apply(t, tableOf(t, 10, 1, true), DealRoles{Seed: 3})

	got := map[RoleID]int{}
	for _, p := range s.Players {
		got[p.Role]++
		if want := Registry[p.Role].BaseTeam(); p.Team != want {
			t.Errorf("%s (%s) team = %s, want %s", p.ID, p.Role, p.Team, want)
		}
	}
	if !reflect.DeepEqual(got, s.Settings.Roles) {
		t.Fatalf("dealt %v, configured %v", got, s.Settings.Roles)
	}
	if s.Phase != PhaseLobby {
		t.Fatalf("phase = %s, want Lobby", s.Phase)
	}
}

func TestDealRoles_RejectedWhenNotReady(t *testing.T) {
	unseated := lobbyOf(t, 5)
	unseated = apply(t, unseated, ConfigureRoles{Roles: map[RoleID]int{roleWerewolf: 1, roleVillager: 4}})

	unconfigured := apply(t, lobbyOf(t, 5), SetSeats{Order: seatOrder(5)})

	lateJoiner := tableOf(t, 5, 1, false)
	lateJoiner = apply(t, lateJoiner, JoinRoom{Player: "late", Nickname: "Late"})

	night := tableOf(t, 5, 1, false)
	night.Phase = PhaseNight

	tests := []struct {
		name  string
		state State
		want  error
	}{
		{"NoRolesConfigured", unconfigured, ErrRolesNotConfigured},
		{"SeatsNotSet", unseated, ErrSeatsNotSet},
		{"PlayerJoinedAfterConfigure", lateJoiner, ErrSeatsNotSet},
		{"NotInLobby", night, ErrWrongPhase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decide(tt.state, DealRoles{Seed: 1}); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestDealRoles_RolesNoLongerFitPlayerCount(t *testing.T) {
	s := tableOf(t, 5, 1, false)
	// Seats refreshed but roles not reconfigured.
	s = apply(t, s, JoinRoom{Player: "late", Nickname: "Late"})
	s = apply(t, s, SetSeats{Order: append(seatOrder(5), "late")})

	if _, err := Decide(s, DealRoles{Seed: 1}); !errors.Is(err, ErrRoleCountMismatch) {
		t.Fatalf("err = %v, want ErrRoleCountMismatch", err)
	}
}

func TestDealRoles_SecondDealRejectedUnlessPlayerListChanged(t *testing.T) {
	s := apply(t, tableOf(t, 5, 1, false), DealRoles{Seed: 1})

	if _, err := Decide(s, DealRoles{Seed: 2}); !errors.Is(err, ErrAlreadyDealt) {
		t.Fatalf("re-deal err = %v, want ErrAlreadyDealt", err)
	}

	// A sixth player joins: seats and roles are redone, then dealing is open again.
	s = apply(t, s, JoinRoom{Player: "p6", Nickname: "P6"})
	s = apply(t, s, SetSeats{Order: seatOrder(6)})
	s = apply(t, s, ConfigureRoles{Roles: map[RoleID]int{roleWerewolf: 1, roleVillager: 5}})
	s = apply(t, s, DealRoles{Seed: 2})

	if !s.Dealt || s.Players["p6"].Role == "" {
		t.Fatalf("second deal not applied: dealt=%v role=%q", s.Dealt, s.Players["p6"].Role)
	}
}

func TestDealRoles_DecideDoesNotMutateState(t *testing.T) {
	s := tableOf(t, 7, 2, false)
	before := dealtRoles(s)

	if _, err := Decide(s, DealRoles{Seed: 9}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, dealtRoles(s)) || s.Dealt {
		t.Fatal("Decide changed the state")
	}
}
