package engine

// onlineNotices is what an online room says at once instead of at a phone
// check (rules §1b): one notice for each living player who has something new
// to learn. s is the state after the command. A lover is told when the pair is
// made and again when either of them is cursed, since the colour changes.
//
// Only those players get one. Around a table everybody needs the same notice
// because neighbours can see it; online nobody can.
func onlineNotices(s State) []Event {
	var evs []Event
	for _, id := range s.Seats {
		p := s.Players[id]
		if p == nil || !p.Alive {
			continue
		}
		partner := s.Players[PartnerOf(s, id)]
		news := p.Flags[FlagTeamHidden] != 0 ||
			partner != nil && (p.Flags[FlagLoveKnown] == 0 || partner.Flags[FlagTeamHidden] != 0)
		if news {
			evs = append(evs, PhoneChecked{Player: id, Cursed: IsCursed(p), Love: loveNotice(s, p, realTeam)})
		}
	}
	return evs
}
