package enginetest

import (
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// Pack checks the wolf pack the viewer sees, in seating order.
func (s *Seen) Pack(names ...string) *Seen {
	s.g.t.Helper()
	if s.view.Pack == nil {
		s.g.t.Fatalf("%s should see the pack %v, sees no pack", s.who, names)
	}
	got := make([]string, len(s.view.Pack.Members))
	for i, id := range s.view.Pack.Members {
		got[i] = string(id)
	}
	if !sameNames(got, names) {
		s.g.t.Fatalf("%s should see the pack %v, sees %v", s.who, names, got)
	}
	return s
}

// NoPack checks that the viewer does not see the pack or its votes.
func (s *Seen) NoPack() *Seen {
	s.g.t.Helper()
	if s.view.Pack != nil {
		s.g.t.Fatalf("%s must not see the pack, sees %v", s.who, s.view.Pack.Members)
	}
	return s
}

// WolfVote checks a live bite vote on the pack screen.
func (s *Seen) WolfVote(wolf, target string) *Seen {
	s.g.t.Helper()
	if s.view.Pack == nil || s.view.Pack.Votes[engine.PlayerID(wolf)] != engine.PlayerID(target) {
		s.g.t.Fatalf("%s should see %s vote to bite %s, sees %+v", s.who, wolf, target, s.view.Pack)
	}
	return s
}

// WolfVotesInFeed checks how many bite votes the viewer knows of.
func (s *Seen) WolfVotesInFeed(n int) *Seen {
	s.g.t.Helper()
	if got := count[roles.WolfVoted](s.feed); got != n {
		s.g.t.Fatalf("%s should know of %d bite votes, knows of %d", s.who, n, got)
	}
	return s
}

// SeerResults checks how many Seer results the viewer knows of.
func (s *Seen) SeerResults(n int) *Seen {
	s.g.t.Helper()
	if got := count[roles.SeerChecked](s.feed); got != n {
		s.g.t.Fatalf("%s should know of %d Seer results, knows of %d", s.who, n, got)
	}
	return s
}

// Event checks that the viewer knows of an event equal to e.
func (s *Seen) Event(e engine.Event) *Seen {
	s.g.t.Helper()
	for _, got := range s.feed {
		if equalEvents(got, e) {
			return s
		}
	}
	s.g.t.Fatalf("%s should know of %T %+v, but does not", s.who, e, e)
	return s
}

// NoEvent checks that the viewer does not know of an event equal to e.
func (s *Seen) NoEvent(e engine.Event) *Seen {
	s.g.t.Helper()
	for _, got := range s.feed {
		if equalEvents(got, e) {
			s.g.t.Fatalf("%s must not know of %T %+v", s.who, e, e)
		}
	}
	return s
}

// Turn checks that the viewer is being called for the given step.
func (s *Seen) Turn(step engine.RoleID) *Seen {
	s.g.t.Helper()
	if s.view.Me == nil || s.view.Me.Turn == nil || s.view.Me.Turn.Step != step {
		s.g.t.Fatalf("%s should be called for %s now", s.who, step)
	}
	return s
}

// NoTurn checks that the viewer is not being called.
func (s *Seen) NoTurn() *Seen {
	s.g.t.Helper()
	if s.view.Me != nil && s.view.Me.Turn != nil {
		s.g.t.Fatalf("%s should not be called now, is called for %s", s.who, s.view.Me.Turn.Step)
	}
	return s
}

// Blocked checks the generic "cannot use your ability tonight" notice on the
// viewer's turn.
func (s *Seen) Blocked(want bool) *Seen {
	s.g.t.Helper()
	got := s.view.Me != nil && s.view.Me.Turn != nil && s.view.Me.Turn.Blocked
	if got != want {
		s.g.t.Fatalf("%s: blocked notice shown=%v, want %v", s.who, got, want)
	}
	return s
}

// Bitten checks whom the viewer sees as bitten tonight; "" means nobody.
func (s *Seen) Bitten(name string) *Seen {
	s.g.t.Helper()
	var got engine.PlayerID
	if s.view.Me != nil && s.view.Me.Turn != nil {
		got = s.view.Me.Turn.Bitten
	}
	if got != engine.PlayerID(name) {
		s.g.t.Fatalf("%s should see %q as bitten, sees %q", s.who, name, got)
	}
	return s
}

// BiteShown checks whether the viewer is shown the bite information on their
// turn; with Bitten("") it means "nobody was bitten".
func (s *Seen) BiteShown(want bool) *Seen {
	s.g.t.Helper()
	got := s.view.Me != nil && s.view.Me.Turn != nil && s.view.Me.Turn.BiteShown
	if got != want {
		s.g.t.Fatalf("%s: bite information shown=%v, want %v", s.who, got, want)
	}
	return s
}

// Cursed checks the viewer's curse notice.
func (s *Seen) Cursed(want bool) *Seen {
	s.g.t.Helper()
	if got := s.view.Me != nil && s.view.Me.Cursed; got != want {
		s.g.t.Fatalf("%s: cursed shown=%v, want %v", s.who, got, want)
	}
	return s
}

// Lover checks the lover's screen: partner, partner's role, link colour.
func (s *Seen) Lover(partner string, role engine.RoleID, colour engine.LinkColour) *Seen {
	s.g.t.Helper()
	want := engine.LoverView{Partner: engine.PlayerID(partner), PartnerRole: role, Colour: colour}
	if s.view.Lover == nil || *s.view.Lover != want {
		s.g.t.Fatalf("%s should see the lover link %+v, sees %+v", s.who, want, s.view.Lover)
	}
	return s
}

// NoLover checks that the viewer sees nothing about a pair.
func (s *Seen) NoLover() *Seen {
	s.g.t.Helper()
	if s.view.Lover != nil {
		s.g.t.Fatalf("%s must not see a lover link, sees %+v", s.who, *s.view.Lover)
	}
	return s
}

// NightCall checks the night cue: only a role and a duration.
func (s *Seen) NightCall(role engine.RoleID, d time.Duration) *Seen {
	s.g.t.Helper()
	want := engine.NightCallView{Role: role, Duration: d}
	if s.view.NightCall == nil || *s.view.NightCall != want {
		s.g.t.Fatalf("%s should see the call %+v, sees %+v", s.who, want, s.view.NightCall)
	}
	return s
}

// NoSecretsOfOthers is the leak check for a living viewer: no
// other player's role, no hidden state, no cause of death, nobody's bite,
// protection or check unless it is the viewer's own.
func (s *Seen) NoSecretsOfOthers() *Seen {
	s.g.t.Helper()
	if s.view.Full || s.view.Secrets != nil {
		s.g.t.Fatalf("%s must not see the hidden state", s.who)
	}
	var me engine.PlayerID
	if s.view.Me != nil {
		me = s.view.Me.ID
	}
	wolf := me != "" && s.view.Me.Alive && s.view.Me.Team == engine.TeamWolves
	if s.view.Pack != nil && !wolf {
		s.g.t.Fatalf("%s is not a wolf and must not see the pack", s.who)
	}
	for _, p := range s.view.Players {
		lover := s.view.Lover != nil && p.ID == s.view.Lover.Partner
		if p.ID != me && !lover && (p.Role != "" || p.Team != engine.TeamNone) {
			s.g.t.Fatalf("%s must not see the role of %s", s.who, p.ID)
		}
	}
	for _, e := range s.feed {
		if leak := leakedBy(e, me, wolf); leak != "" {
			s.g.t.Fatalf("%s must not know %s (%T %+v)", s.who, leak, e, e)
		}
	}
	return s
}

func leakedBy(e engine.Event, me engine.PlayerID, wolf bool) string {
	switch ev := e.(type) {
	case engine.PlayerDied:
		return "the order of the deaths (who was bitten, who was poisoned)"
	case engine.RolesDealt:
		return "the deal"
	case roles.WolfVoted:
		if !wolf {
			return "the pack's votes"
		}
	case roles.PlayerProtected:
		if ev.Guard != me {
			return "whom the Guard protected"
		}
	case roles.SeerChecked:
		if ev.Seer != me {
			return "the Seer's result"
		}
	case engine.TeamChanged:
		return "a team change: the phone check is the only place it shows"
	case roles.PlayerCursed:
		if ev.Wolf != me {
			return "the curse"
		}
	}
	return ""
}

func count[T engine.Event](feed []engine.Event) int {
	n := 0
	for _, e := range feed {
		if _, ok := e.(T); ok {
			n++
		}
	}
	return n
}
