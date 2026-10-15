package enginetest

import (
	"maps"
	"reflect"

	"masoi/internal/engine"
)

// As looks at the game through one player's phone: g.As("An").OwnRole(Seer).
// It reads both what the screen shows now (engine.ViewFor) and the events the
// player may know about (engine.Visible).
func (g *Game) As(name string) *Seen {
	g.t.Helper()
	return g.seen(name, engine.PlayerViewer(g.Player(name)))
}

func (g *Game) seen(who string, v engine.Viewer) *Seen {
	return &Seen{
		g:    g,
		who:  who,
		view: engine.ViewFor(g.state, v),
		feed: engine.Visible(g.state, g.events, v),
	}
}

// HandBuild edits the state directly. It is only for roles that are not built
// yet (Witch, Visitor, Cupid...), to test what their screens would show. The
// edit is not an event, so it shows in views but not in event feeds.
func (g *Game) HandBuild(edit func(s *engine.State)) *Game {
	s := g.state
	edit(&s)
	g.state = s
	return g
}

// Inject records events no built role can produce yet, e.g. a curse.
func (g *Game) Inject(evs ...engine.Event) *Game {
	g.record(evs...)
	return g
}

// Seen is one screen at one moment.
type Seen struct {
	g    *Game
	who  string
	view engine.View
	feed []engine.Event
}

// View returns the raw projection.
func (s *Seen) View() engine.View { return s.view }

// Feed returns the raw visible events.
func (s *Seen) Feed() []engine.Event { return s.feed }

func (s *Seen) seat(name string) engine.PlayerView {
	s.g.t.Helper()
	for _, p := range s.view.Players {
		if p.ID == engine.PlayerID(name) {
			return p
		}
	}
	s.g.t.Fatalf("%s does not see %s at the table", s.who, name)
	return engine.PlayerView{}
}

// OwnRole checks the viewer's own card.
func (s *Seen) OwnRole(role engine.RoleID) *Seen {
	s.g.t.Helper()
	if s.view.Me == nil || s.view.Me.Role != role {
		s.g.t.Fatalf("%s should see their own role %s, got %+v", s.who, role, s.view.Me)
	}
	return s
}

// RoleOf checks that the viewer sees name's role.
func (s *Seen) RoleOf(name string, role engine.RoleID) *Seen {
	s.g.t.Helper()
	if got := s.seat(name).Role; got != role {
		s.g.t.Fatalf("%s should see %s as %s, sees %q", s.who, name, role, got)
	}
	return s
}

// NoRoleExceptOwn checks that no other seat shows a role or a team.
func (s *Seen) NoRoleExceptOwn() *Seen {
	s.g.t.Helper()
	for _, p := range s.view.Players {
		if s.view.Me != nil && p.ID == s.view.Me.ID {
			continue
		}
		if p.Role != "" || p.Team != engine.TeamNone || p.Card != nil {
			s.g.t.Fatalf("%s must not see the role of %s, sees %s/%s %+v", s.who, p.ID, p.Role, p.Team, p.Card)
		}
	}
	return s
}

// PublicTable checks what every screen shows: the role list, seating, alive
// flags and the Chief, all matching the game.
func (s *Seen) PublicTable() *Seen {
	s.g.t.Helper()
	st := s.g.state
	if !maps.Equal(s.view.Roles, st.Settings.Roles) {
		s.g.t.Fatalf("%s should see the role list %v, sees %v", s.who, st.Settings.Roles, s.view.Roles)
	}
	if len(s.view.Players) != len(st.Seats) {
		s.g.t.Fatalf("%s should see %d seats, sees %d", s.who, len(st.Seats), len(s.view.Players))
	}
	for i, p := range s.view.Players {
		if p.ID != st.Seats[i] || p.Seat != i+1 || p.Nickname != st.Players[p.ID].Nickname {
			s.g.t.Fatalf("%s sees seat %d as %+v, want %s", s.who, i+1, p, st.Seats[i])
		}
		if p.Alive != st.Players[p.ID].Alive {
			s.g.t.Fatalf("%s sees %s alive=%v, want %v", s.who, p.ID, p.Alive, !p.Alive)
		}
	}
	if s.view.Chief != st.Chief {
		s.g.t.Fatalf("%s sees Chief %q, want %q", s.who, s.view.Chief, st.Chief)
	}
	return s
}

// Dead checks that the viewer sees these players as dead.
func (s *Seen) Dead(names ...string) *Seen {
	s.g.t.Helper()
	for _, name := range names {
		if s.seat(name).Alive {
			s.g.t.Fatalf("%s should see %s as dead", s.who, name)
		}
	}
	return s
}

// Nomination checks a public nomination: voter put target up ("" is no one).
func (s *Seen) Nomination(voter, target string) *Seen {
	s.g.t.Helper()
	got, ok := s.view.Votes.Nominations[engine.PlayerID(voter)]
	if !ok || got != engine.PlayerID(target) {
		s.g.t.Fatalf("%s should see %s nominate %q, sees %q (voted=%v)", s.who, voter, target, got, ok)
	}
	return s
}

// ChiefVote checks a public vote of the Chief election.
func (s *Seen) ChiefVote(voter, candidate string) *Seen {
	s.g.t.Helper()
	if got := s.view.Votes.Chief[engine.PlayerID(voter)]; got != engine.PlayerID(candidate) {
		s.g.t.Fatalf("%s should see %s vote %s for Chief, sees %q", s.who, voter, candidate, got)
	}
	return s
}

// Verdict checks a public kill (true) or spare vote.
func (s *Seen) Verdict(voter string, kill bool) *Seen {
	s.g.t.Helper()
	got, ok := s.view.Votes.Verdicts[engine.PlayerID(voter)]
	if !ok || got != kill {
		s.g.t.Fatalf("%s should see %s vote kill=%v, sees kill=%v (voted=%v)", s.who, voter, kill, got, ok)
	}
	return s
}

// Full checks that the viewer sees everything (dead, or the game is over).
func (s *Seen) Full() *Seen {
	s.g.t.Helper()
	if !s.view.Full || s.view.Secrets == nil {
		s.g.t.Fatalf("%s should see everything, but the view is not full", s.who)
	}
	for _, p := range s.view.Players {
		if want := s.g.state.Players[p.ID].Role; p.Role != want {
			s.g.t.Fatalf("%s should see %s as %s, sees %q", s.who, p.ID, want, p.Role)
		}
	}
	if len(s.feed) != len(s.g.events) {
		s.g.t.Fatalf("%s should see all %d events, sees %d", s.who, len(s.g.events), len(s.feed))
	}
	return s
}

func equalEvents(a, b engine.Event) bool { return reflect.DeepEqual(a, b) }

// TurnOf lets the night run until the given role is called, for checking what
// a screen shows during that step.
func (n *Night) TurnOf(role engine.RoleID) *Night {
	n.g.t.Helper()
	n.advanceTo(role)
	return n
}
