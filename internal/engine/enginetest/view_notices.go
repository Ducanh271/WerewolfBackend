package enginetest

// The notice texts the server sends next to the flags (plan T30e).

// TurnNotice checks the text shown on the viewer's own turn for being
// blocked, poisoned or grieving; "" means no text.
func (s *Seen) TurnNotice(want string) *Seen {
	s.g.t.Helper()
	var got string
	if s.view.Me != nil && s.view.Me.Turn != nil {
		got = s.view.Me.Turn.Notice
	}
	if got != want {
		s.g.t.Fatalf("%s: turn notice %q, want %q", s.who, got, want)
	}
	return s
}

// PhoneNoticeText checks the two lines of the phone check notice; "" means
// the line is not sent at all.
func (s *Seen) PhoneNoticeText(cursed, love string) *Seen {
	s.g.t.Helper()
	if s.view.Me == nil || s.view.Me.PhoneCheck == nil {
		s.g.t.Fatalf("%s should see a phone check notice", s.who)
	}
	pc := s.view.Me.PhoneCheck
	if got := textOr(pc.CursedText); got != cursed {
		s.g.t.Fatalf("%s: curse line %q, want %q", s.who, got, cursed)
	}
	if got := textOr(pc.LoveText); got != love {
		s.g.t.Fatalf("%s: love line %q, want %q", s.who, got, love)
	}
	return s
}

// PackDoomedText checks the lines the pack is shown about wolves that can no
// longer vote, in order.
func (s *Seen) PackDoomedText(lines ...string) *Seen {
	s.g.t.Helper()
	var got []string
	if s.view.Pack != nil {
		got = s.view.Pack.DoomedNotices
	}
	if len(got) != len(lines) {
		s.g.t.Fatalf("%s: pack lines %q, want %q", s.who, got, lines)
	}
	for i := range got {
		if got[i] != lines[i] {
			s.g.t.Fatalf("%s: pack lines %q, want %q", s.who, got, lines)
		}
	}
	return s
}

func textOr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
