package engine

import (
	"fmt"
	"sort"
)

// MaxWolfPercent is the share of players that may be wolves (rules §2:
// "không vượt 30% số người chơi, làm tròn xuống").
const MaxWolfPercent = 30

// MaxWolves returns the highest number of wolves allowed at a table of
// playerCount, counting the villager a cursed wolf will turn later.
//
// The 30% floor reproduces every row of the rules §2 table and also covers the
// table's gaps: 5-6 -> 1, 7-9 -> 2, 10-13 -> 3, 14-16 -> 4, 20 -> 6, 30 -> 9.
func MaxWolves(playerCount int) int {
	return playerCount * MaxWolfPercent / 100
}

// allowsMultipleCopies reports whether a role may be dealt to more than one
// player. Only Villager and the plain Werewolf opt in; every other role is
// unique (rules §2 "Số lượng mỗi vai").
func allowsMultipleCopies(role Role) bool {
	multi, ok := role.(MultiCopyRole)
	return ok && multi.AllowsMultipleCopies()
}

// validateRoleConfig checks a role list against rules §2.
//
// Roles are visited in ID order so a config with several problems always
// reports the same error.
func validateRoleConfig(playerCount int, roles map[RoleID]int) error {
	if playerCount < MinPlayers {
		return fmt.Errorf("%w: %d players, need at least %d", ErrTooFewPlayers, playerCount, MinPlayers)
	}
	if playerCount > MaxPlayers {
		return fmt.Errorf("%w: %d players, allowed at most %d", ErrTooManyPlayers, playerCount, MaxPlayers)
	}

	ids := make([]RoleID, 0, len(roles))
	for id := range roles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	slots, wolves, cursedReserve := 0, 0, 0
	for _, id := range ids {
		count := roles[id]
		if count < 1 {
			return fmt.Errorf("%w: %s has count %d", ErrInvalidRoleCount, id, count)
		}
		role, ok := Registry[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrUnknownRole, id)
		}
		if count > 1 && !allowsMultipleCopies(role) {
			return fmt.Errorf("%w: %s is unique, asked for %d", ErrRoleMustBeUnique, id, count)
		}
		if rule, ok := role.(MinPlayersRule); ok && playerCount < rule.MinPlayers() {
			return fmt.Errorf("%w: %s needs at least %d players, have %d",
				ErrRoleNeedsMorePlayers, id, rule.MinPlayers(), playerCount)
		}
		if curse, ok := role.(CurseSource); ok && curse.CursesVillager() {
			// However many curse sources are picked, the limit reserves
			// one extra wolf slot (rules §2 table).
			cursedReserve = 1
		}

		slots += count
		if role.BaseTeam() == TeamWolves {
			wolves += count
		}
	}

	if slots != playerCount {
		return fmt.Errorf("%w: %d slots for %d players", ErrRoleCountMismatch, slots, playerCount)
	}
	if wolves == 0 {
		return ErrNoWolves
	}
	if limit := MaxWolves(playerCount); wolves+cursedReserve > limit {
		return fmt.Errorf("%w: %d wolves (+%d to be cursed) at %d players, allowed at most %d",
			ErrTooManyWolves, wolves, cursedReserve, playerCount, limit)
	}
	return nil
}
