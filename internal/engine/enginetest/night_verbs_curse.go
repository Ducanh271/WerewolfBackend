package enginetest

import (
	"masoi/internal/engine"
	"masoi/internal/roles"
)

// CursedWolf is the Cursed wolf: CursedWolf("An").Curses("Binh").
func (n *Night) CursedWolf(name string) *CursedWolfActor {
	return &CursedWolfActor{n: n, id: n.Player(name)}
}

// CursedWolfActor is the Cursed wolf acting.
type CursedWolfActor struct {
	n  *Night
	id engine.PlayerID
}

// Curses turns the player into a wolf at once. Their phone and the pack's
// change only at the phone check: see PhoneCheck.
func (a *CursedWolfActor) Curses(target string) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.CursedWolf)
	return a.n.Do(roles.CursedWolfCurse{By: a.id, Target: a.n.Player(target)})
}

// CannotCurse curses and expects the engine to refuse with want.
func (a *CursedWolfActor) CannotCurse(target string, want error) *Night {
	a.n.g.t.Helper()
	a.n.advanceTo(roles.CursedWolf)
	return a.n.DoExpectErr(roles.CursedWolfCurse{By: a.id, Target: a.n.Player(target)}, want)
}

// PhoneCheck lets the night run to the phone check step, where everybody looks
// at their phone (rules §3), for checking what the screens show.
func (n *Night) PhoneCheck() *Night {
	n.g.t.Helper()
	n.advanceTo(engine.StepPhoneCheck)
	return n
}

// PhoneNotice checks the notice the viewer sees at the phone check: cursed or
// not.
func (s *Seen) PhoneNotice(cursed bool) *Seen {
	s.g.t.Helper()
	if s.view.Me == nil || s.view.Me.PhoneCheck == nil || s.view.Me.PhoneCheck.Cursed != cursed {
		s.g.t.Fatalf("%s should see the phone check notice cursed=%v, sees %+v", s.who, cursed, s.view.Me)
	}
	return s
}

// NoPhoneNotice checks that the viewer sees no phone check notice.
func (s *Seen) NoPhoneNotice() *Seen {
	s.g.t.Helper()
	if s.view.Me != nil && s.view.Me.PhoneCheck != nil {
		s.g.t.Fatalf("%s must not see a phone check notice now, sees %+v", s.who, s.view.Me.PhoneCheck)
	}
	return s
}

// PhoneNoticesInFeed checks how many phone check notices the viewer knows of.
func (s *Seen) PhoneNoticesInFeed(n int) *Seen {
	s.g.t.Helper()
	if got := count[engine.PhoneChecked](s.feed); got != n {
		s.g.t.Fatalf("%s should know of %d phone check notices, knows of %d", s.who, n, got)
	}
	return s
}
