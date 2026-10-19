package engine

import "sort"

// CardRole is implemented by a role printed in the rules §9b table. Its rank
// there orders the public role list of a game, lower first. Ranks are spaced
// (10, 20, …) so a new role can slot in without renumbering the others.
type CardRole interface {
	CardOrder() int
}

// CardOrder lists the registered roles in the order of the rules §9b table:
// roles with a CardOrder by rank, then the others by ID.
func CardOrder() []RoleID {
	ids := make([]RoleID, 0, len(Registry))
	for id := range Registry {
		ids = append(ids, id)
	}
	sortByCard(ids)
	return ids
}

// sortByCard sorts role IDs in the order of CardOrder.
func sortByCard(ids []RoleID) {
	rank := func(id RoleID) (int, bool) {
		if r, ok := Registry[id].(CardRole); ok {
			return r.CardOrder(), true
		}
		return 0, false
	}
	sort.Slice(ids, func(i, j int) bool {
		ri, oki := rank(ids[i])
		rj, okj := rank(ids[j])
		switch {
		case oki != okj:
			return oki
		case ri != rj:
			return ri < rj
		}
		return ids[i] < ids[j]
	})
}
