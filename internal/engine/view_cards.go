package engine

import (
	"sort"

	"masoi/internal/i18n"
)

// RoleCard is a player's card as a phone shows it (rules §9b): the role's name
// and description, with the team and goal of the player's CURRENT side (a
// cursed Seer shows the wolves, a lover of a cross-team couple the third team).
// The text comes from i18n, the one source of truth. Every card has these four
// fields, all set, so no card can be told from another by its shape.
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
	Team        string // the role's own team, public from the rules
	Description string
}

// StatusCard is an extra card a player holds while the state applies (rules
// §9b "Thẻ phụ"): cursed, in love (same team or cross-team) or Chief.
type StatusCard struct {
	Kind string
	Name string
	Text string
}

// roleCard is the card of a role on a side (i18n.SideVillage...); side ""
// keeps the role's own side, as printed. A role with no text gets its ID as
// every field, so the shape stays the same.
func roleCard(role RoleID, side string) *RoleCard {
	if role == "" {
		return nil
	}
	c, ok := i18n.Card(string(role))
	if !ok {
		id := string(role)
		return &RoleCard{Name: id, Team: id, Goal: id, Description: id}
	}
	out := &RoleCard{Name: c.Name, Team: c.Team, Goal: c.Goal, Description: c.Description}
	if side != "" {
		out.Team, out.Goal = i18n.Side(side)
	}
	return out
}

func sideOfTeam(t Team) string {
	if t == TeamWolves {
		return i18n.SideWolves
	}
	return i18n.SideVillage
}

// fillCards sets the card of every seat whose role the viewer sees, and the
// viewer's own card. The side shown is what the viewer may know:
//   - full views: the real current side, the third team for both lovers of a
//     cross-team couple (alive, or broken as the third team);
//   - the viewer and their partner: the side as the viewer's screen shows it
//     (VisibleTeam; the third team from the phone check that told the pair);
//   - a revealed Hunter seen by the living: his card as printed, since only
//     his role is public, never his side (rules §1, §6).
func fillCards(s State, out *View, me *Player, full bool) {
	for i := range out.Players {
		pv := &out.Players[i]
		p := s.Players[pv.ID]
		if pv.Role == "" || p == nil {
			continue
		}
		side := ""
		switch {
		case full:
			side = sideOfTeam(p.Team)
			if PartnerOf(s, p.ID) != "" && s.Lovers != nil && (s.Lovers.ThirdTeam || crossTeamNow(s)) {
				side = i18n.SideThird
			}
		case me != nil && (p.ID == me.ID || (out.Lover != nil && p.ID == out.Lover.Partner)):
			side = sideOfTeam(VisibleTeam(p))
			if out.Lover != nil && (out.Lover.Colour == LinkStriped || s.Lovers.ThirdTeam) {
				side = i18n.SideThird
			}
		}
		pv.Card = roleCard(pv.Role, side)
		if me != nil && p.ID == me.ID && out.Me != nil {
			card := *pv.Card
			out.Me.RoleCard = &card
		}
	}
}

func crossTeamNow(s State) bool {
	_, _, crossed := crossTeamLovers(s)
	return crossed
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
		c := roleCard(id, "")
		out = append(out, RoleListCard{Role: id, Name: c.Name, Team: c.Team, Description: c.Description})
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
