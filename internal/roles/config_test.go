package roles_test

import (
	"errors"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// TestConfigureRoles_NeedsAtLeastOneVillager covers rules §2 "Dân làng: 1 trở
// lên": a table with every unique village role, plus wolves at the limit,
// and no Villager at all must still be rejected, even though every other
// rule (wolf cap, uniqueness, slot count, minimum table sizes) holds.
func TestConfigureRoles_NeedsAtLeastOneVillager(t *testing.T) {
	// 11 players: the 8 unique village roles (Detective needs more than 8
	// players, Cupid needs at least 10, both satisfied) plus 3 wolves, the
	// limit for 11 players (rules §2 table, 10-13 players row) — every seat
	// filled, nothing left over for a Villager.
	noVillager := map[engine.RoleID]int{
		roles.Werewolf: 3, roles.Seer: 1, roles.Guard: 1, roles.Witch: 1, roles.Hunter: 1,
		roles.Visitor: 1, roles.Detective: 1, roles.RustyKnight: 1, roles.Cupid: 1,
	}
	s := lobbyOf(t, 11)
	if _, err := engine.Decide(s, engine.ConfigureRoles{Roles: noVillager}); !errors.Is(err, engine.ErrRoleRequired) {
		t.Fatalf("no villager = %v, want %v", err, engine.ErrRoleRequired)
	}

	// Trade one wolf for one Villager: still 11 slots, still within every
	// other limit, now accepted.
	withVillager := map[engine.RoleID]int{}
	for id, n := range noVillager {
		withVillager[id] = n
	}
	withVillager[roles.Werewolf] = 2
	withVillager[roles.Villager] = 1
	if _, err := engine.Decide(s, engine.ConfigureRoles{Roles: withVillager}); err != nil {
		t.Fatalf("with a villager = %v, want accepted", err)
	}
}
