package enginetest

import (
	"reflect"

	"masoi/internal/engine"
	"masoi/internal/i18n"
)

// CardText is the card of a role as rules §9b writes it (from i18n).
func CardText(role engine.RoleID) engine.RoleCard {
	c, _ := i18n.Card(string(role))
	return engine.RoleCard{Name: c.Name, Team: c.Team, Goal: c.Goal, Description: c.Description}
}

// OwnCard checks the viewer's own role card.
func (s *Seen) OwnCard(role engine.RoleID) *Seen {
	s.g.t.Helper()
	want := CardText(role)
	if s.view.Me == nil || s.view.Me.RoleCard == nil || *s.view.Me.RoleCard != want {
		s.g.t.Fatalf("%s should see the card of %s %+v, sees %+v", s.who, role, want, s.view.Me)
	}
	return s
}

// CardOf checks the card the viewer sees on name's seat; "" means no card.
func (s *Seen) CardOf(name string, role engine.RoleID) *Seen {
	s.g.t.Helper()
	got := s.seat(name).Card
	if role == "" {
		if got != nil {
			s.g.t.Fatalf("%s must not see the card of %s, sees %+v", s.who, name, *got)
		}
		return s
	}
	if want := CardText(role); got == nil || *got != want {
		s.g.t.Fatalf("%s should see %s's card %+v, sees %+v", s.who, name, want, got)
	}
	return s
}

// StatusCards checks the viewer's extra cards, by kind and in order.
func (s *Seen) StatusCards(kinds ...string) *Seen {
	s.g.t.Helper()
	var got []string
	if s.view.Me != nil {
		for _, c := range s.view.Me.StatusCards {
			got = append(got, c.Kind)
		}
	}
	if !reflect.DeepEqual(got, kinds) && (len(got) != 0 || len(kinds) != 0) {
		s.g.t.Fatalf("%s should hold the extra cards %v, holds %v", s.who, kinds, got)
	}
	return s
}
