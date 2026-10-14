package engine

import (
	"fmt"
	"sort"
)

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
			// extra wolf slots (rules §2 table).
			cursedReserve = CurseWolfReserve
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
