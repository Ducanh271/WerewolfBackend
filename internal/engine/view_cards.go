package engine

import (
	"sort"

	"masoi/internal/i18n"
)

// RoleCard is a role's card as a phone shows it (rules §9b): the text comes
// from i18n, the one source of truth. Every card has these four fields, all
// set, so no card can be told from another by its shape.
type RoleCard struct {
	Name        string
	Team        string
	Goal        string
	Description string
}

// RoleListCard is one role of the game's public role list, with its name and
// short description.
type RoleListCard struct {
	Role        RoleID
	Name        string
	Description string
}

// StatusCard is an extra card a player holds while the state applies (rules
// §9b "Thẻ phụ"): cursed, in love (same team or cross-team) or Chief.
type StatusCard struct {
	Kind string
	Name string
	Text string
}

// roleCard is the card of a role. A role with no text gets its ID as every
// field, so the shape stays the same.
func roleCard(role RoleID) *RoleCard {
	if role == "" {
		return nil
	}
	c, ok := i18n.Card(string(role))
	if !ok {
		id := string(role)
		return &RoleCard{Name: id, Team: id, Goal: id, Description: id}
	}
	return &RoleCard{Name: c.Name, Team: c.Team, Goal: c.Goal, Description: c.Description}
}

// roleListCards is the public role list of the game, in the order of the
// rules §9b table (roles without text last, by ID).
func roleListCards(s State) []RoleListCard {
	rank := map[RoleID]int{}
	for i, id := range i18n.CardOrder() {
		rank[RoleID(id)] = i + 1
	}
	var ids []RoleID
	for id, n := range s.Settings.Roles {
		if n > 0 {
			ids = append(ids, id)
		}
	}
	order := func(id RoleID) int {
		if r, ok := rank[id]; ok {
			return r
		}
		return len(rank) + 1
	}
	sort.Slice(ids, func(i, j int) bool {
		if oi, oj := order(ids[i]), order(ids[j]); oi != oj {
			return oi < oj
		}
		return ids[i] < ids[j]
	})
	out := make([]RoleListCard, 0, len(ids))
	for _, id := range ids {
		c := roleCard(id)
		out = append(out, RoleListCard{Role: id, Name: c.Name, Description: c.Description})
	}
	return out
}

// statusCards are the extra cards of the viewer, each only once the viewer may
// know it: the curse and the pair from the phone check on, the title while
// they hold it. lover is the viewer's LoverView (nil outside the pair).
func statusCards(s State, me *Player, cursed bool, lover *LoverView) []StatusCard {
	var out []StatusCard
	add := func(kind, partner string) {
		name, text := i18n.StatusCard(kind, partner)
		out = append(out, StatusCard{Kind: kind, Name: name, Text: text})
	}
	if cursed {
		add(i18n.StatusCursed, "")
	}
	if lover != nil {
		partner := string(lover.Partner)
		if p := s.Players[lover.Partner]; p != nil {
			partner = p.Nickname
		}
		if lover.Colour == LinkStriped || (s.Lovers != nil && s.Lovers.ThirdTeam) {
			add(i18n.StatusLoversCross, partner)
		} else {
			add(i18n.StatusLovers, partner)
		}
	}
	if s.Chief != "" && s.Chief == me.ID {
		add(i18n.StatusChief, "")
	}
	return out
}
