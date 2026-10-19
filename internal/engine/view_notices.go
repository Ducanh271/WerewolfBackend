package engine

import "masoi/internal/i18n"

// The notices a phone shows as text on its own screen (rules §3), worded by
// i18n so no client writes its own (plan T30e). Each sits next to the flag it
// spells out; the flags stay as they were.

// turnNotice is the text for a TurnView's flags, "" when none is set. Grief
// and the Rusty knight's poison both set Doomed; Grief tells them apart. A
// doomed player cannot act anyway, so the death wins over a block.
func turnNotice(t *TurnView) string {
	switch {
	case t.Grief:
		return i18n.Vi(i18n.NightGrief)
	case t.Doomed:
		return i18n.Vi(i18n.NightWolfDoomed)
	case t.Blocked:
		return i18n.Vi(i18n.NightBlocked)
	}
	return ""
}

// phoneCheckView is the phone check notice with its two lines of text. Each
// line is set only when the public role list has the role it is about (a
// Cursed wolf for the curse line, Cupid for the love line), so every phone
// gets the same keys and only the values differ.
func phoneCheckView(s State, p *Player) *PhoneCheckView {
	out := &PhoneCheckView{Cursed: IsCursed(p), Love: loveNotice(s, p, VisibleTeam)}
	if roleListHas(s, func(r Role) bool { c, ok := r.(CurseSource); return ok && c.CursesVillager() }) {
		key := i18n.NoticeNotCursed
		if out.Cursed {
			key = i18n.NoticeCursed
		}
		out.CursedText = textOf(i18n.Vi(key))
	}
	if roleListHas(s, func(r Role) bool { l, ok := r.(LoveSource); return ok && l.CreatesLovers() }) {
		text := i18n.Vi(i18n.NoticeNoLove)
		if out.Love.InLove {
			text = i18n.Vi(i18n.NoticeInLove, nicknameOf(s, out.Love.Partner), roleName(out.Love.PartnerRole))
		}
		out.LoveText = textOf(text)
	}
	return out
}

// packDoomedNotices is one line per wolf in doomed, in the same order.
func packDoomedNotices(s State, doomed []PlayerID) []string {
	var out []string
	for _, id := range doomed {
		out = append(out, i18n.Vi(i18n.NightPackDoomed, nicknameOf(s, id)))
	}
	return out
}

// roleListHas reports whether a role of the game's public role list matches.
func roleListHas(s State, match func(Role) bool) bool {
	for id, n := range s.Settings.Roles {
		if r := Registry[id]; n > 0 && r != nil && match(r) {
			return true
		}
	}
	return false
}

func nicknameOf(s State, id PlayerID) string {
	if p := s.Players[id]; p != nil {
		return p.Nickname
	}
	return string(id)
}

// roleName is the role's name as its card prints it.
func roleName(role RoleID) string {
	if c, ok := i18n.Card(string(role)); ok {
		return c.Name
	}
	return string(role)
}

func textOf(s string) *string { return &s }
