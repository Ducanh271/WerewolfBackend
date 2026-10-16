package roles_test

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// pickPlayerCount weights the table size toward 10-13 players, the common
// case, but still plays 5-9 and up to 20.
func pickPlayerCount(rng *rand.Rand) int {
	switch r := rng.IntN(100); {
	case r < 55:
		return 10 + rng.IntN(4)
	case r < 80:
		return 5 + rng.IntN(5)
	default:
		return 14 + rng.IntN(7)
	}
}

// randomRoleList builds a legal role list from the role registry and the
// limits of rules §2: roles registered later are picked up by themselves, as
// long as they respect their MinPlayers and their unique/multi-copy rule.
func randomRoleList(rng *rand.Rand, players int) map[engine.RoleID]int {
	ids := make([]engine.RoleID, 0, len(engine.Registry))
	for id := range engine.Registry {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] }) // the same seed gives the same list
	rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })

	list := map[engine.RoleID]int{}
	special, reserve := 0, 0
	for _, id := range ids {
		role := engine.Registry[id]
		if id == roles.Villager || role.BaseTeam() == engine.TeamWolves && multiCopy(role) {
			continue
		}
		if rule, ok := role.(engine.MinPlayersRule); ok && players < rule.MinPlayers() {
			continue
		}
		if rng.IntN(100) >= 75 { // not every role every game
			continue
		}
		if curse, ok := role.(engine.CurseSource); ok && curse.CursesVillager() {
			reserve = engine.CurseWolfReserve
		}
		list[id] = 1
		special++
	}

	wolves := 0
	for id := range list {
		if engine.Registry[id].BaseTeam() == engine.TeamWolves {
			wolves++
		}
	}
	// Drop wolf specials until the limit, curse reserve included, holds.
	for wolves > 0 && wolves+reserve > engine.MaxWolves(players) {
		for _, id := range ids {
			if _, in := list[id]; in && engine.Registry[id].BaseTeam() == engine.TeamWolves {
				delete(list, id)
				wolves--
				special--
				break
			}
		}
		reserve = 0
		for id := range list {
			if curse, ok := engine.Registry[id].(engine.CurseSource); ok && curse.CursesVillager() {
				reserve = engine.CurseWolfReserve
			}
		}
	}
	// A plain wolf pack: at least one wolf in the game, more when there is room.
	room := engine.MaxWolves(players) - wolves - reserve
	plain := 0
	if wolves == 0 {
		plain = 1
	}
	if room > 1 && rng.IntN(2) == 0 {
		plain += 1 + rng.IntN(room-plain)
	}
	if plain > 0 {
		list[roles.Werewolf] = plain
	}
	// Too many specials for a small table: drop the non-wolf ones.
	for special+plain > players {
		for _, id := range ids {
			if _, in := list[id]; in && id != roles.Werewolf && engine.Registry[id].BaseTeam() != engine.TeamWolves {
				delete(list, id)
				special--
				break
			}
		}
	}
	v := players - special - plain
	if v < 1 {
		// rules §2 "Dân làng: 1 trở lên" is never optional; free one slot for
		// it, a non-wolf special first, else one plain wolf.
		freed := false
		for _, id := range ids {
			if _, in := list[id]; in && id != roles.Werewolf && engine.Registry[id].BaseTeam() != engine.TeamWolves {
				delete(list, id)
				special--
				freed = true
				break
			}
		}
		if !freed && plain > 0 {
			plain--
			if plain == 0 {
				delete(list, roles.Werewolf)
			} else {
				list[roles.Werewolf] = plain
			}
		}
		v = players - special - plain
	}
	list[roles.Villager] = v
	return list
}

func multiCopy(r engine.Role) bool {
	m, ok := r.(engine.MultiCopyRole)
	return ok && m.AllowsMultipleCopies()
}

// setup seats the players in a random order and deals the roles, the
// way the room does, and returns the state at the start of the role view.
func (g *simGame) setup() error {
	count := pickPlayerCount(g.rng)
	list := randomRoleList(g.rng, count)
	g.desc = fmt.Sprintf("%d players, roles %v", count, list)
	names := make([]engine.PlayerID, count)
	for i := range names {
		names[i] = engine.PlayerID(fmt.Sprintf("P%d", i+1))
	}
	steps := []engine.Command{}
	for _, id := range names {
		steps = append(steps, engine.JoinRoom{Player: id, Nickname: string(id)})
	}
	order := append([]engine.PlayerID(nil), names...)
	g.rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
	steps = append(steps, engine.SetSeats{Order: order}, engine.ConfigureRoles{Roles: list},
		engine.DealRoles{Seed: g.rng.Int64()})
	for _, c := range steps {
		if err := g.do(c); err != nil {
			return fmt.Errorf("setup %T: %w", c, err)
		}
	}
	return nil
}
