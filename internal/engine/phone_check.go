package engine

import (
	"time"

	"masoi/internal/i18n"
)

// StepPhoneCheck is the night step in which everybody looks at their phone
// (rules §3 "Kiểm tra điện thoại"). It belongs to no role: it follows Cupid's
// step on night 1 and the curse step on nights 1-3, on a fixed schedule (see
// engineSlots), so it tells nothing. A night may have it twice.
const StepPhoneCheck RoleID = "PhoneCheck"

// FlagTeamHidden is set on a player whose team changed during the night, until
// the phone check shows them the change. While it holds, no view treats the
// player as being on the new team (VisibleTeam).
const FlagTeamHidden = "team.hidden"

// PhoneCheckAnnounced is the voice line that opens the step, read aloud by the
// speaker phone. Public: it is the same on every phone.
type PhoneCheckAnnounced struct {
	EventMarker
}

// AnnouncementKey is read aloud to everybody.
func (PhoneCheckAnnounced) AnnouncementKey() string { return i18n.NightPhoneCheck }

// PhoneChecked is the notice one living player gets at the phone check. It
// combines the curse status and the love status, and every notice has this
// exact shape whatever the answer: only the values differ. A client shows the
// curse part when the (public) role list has a Cursed wolf and the love part
// when it has Cupid.
type PhoneChecked struct {
	EventMarker
	Player PlayerID
	Cursed bool
	Love   LoveNotice
}

// LoveNotice is the love part of the notice. Outside the pair InLove is false
// and the other fields are empty.
type LoveNotice struct {
	InLove      bool
	Partner     PlayerID
	PartnerRole RoleID
	Colour      LinkColour
}

// PhoneCheckView is that notice as the phone shows it, while the step runs.
type PhoneCheckView struct {
	Cursed bool
	Love   LoveNotice
}

// loveNotice is what the player is told of the pair, with the colour worked
// out from the teams team returns (the real ones at the phone check, the ones
// the views show afterwards).
func loveNotice(s State, p *Player, team func(*Player) Team) LoveNotice {
	partner := s.Players[PartnerOf(s, p.ID)]
	if partner == nil {
		return LoveNotice{}
	}
	return LoveNotice{InLove: true, Partner: partner.ID, PartnerRole: partner.Role, Colour: coupleColour(team(p), team(partner))}
}

func realTeam(p *Player) Team { return p.Team }

// IsCursed reports whether the player is on the wolf team without holding a
// wolf role.
func IsCursed(p *Player) bool {
	role := Registry[p.Role]
	return role != nil && p.Team == TeamWolves && role.BaseTeam() != TeamWolves
}

// VisibleTeam is the team a viewer other than the engine may act on: the team
// of the player's card, until the phone check has told them of a change.
func VisibleTeam(p *Player) Team {
	if p.Flags[FlagTeamHidden] != 0 {
		if role := Registry[p.Role]; role != nil {
			return role.BaseTeam()
		}
	}
	return p.Team
}

// phoneCheckNotices is one notice per living player, in seating order.
func phoneCheckNotices(s State) []Event {
	var evs []Event
	for _, id := range s.Seats {
		if p := s.Players[id]; p != nil && p.Alive {
			evs = append(evs, PhoneChecked{Player: id, Cursed: IsCursed(p), Love: loveNotice(s, p, realTeam)})
		}
	}
	return evs
}

// engineSlots are the steps of this night that belong to no role, each placed
// right after the role step it follows (rules §3):
//   - after a CurseSource's step (the Cursed wolf, nights 1-3): the phone check;
//   - after a LoveSource's step (Cupid, night 1): the phone check, then the
//     lovers' step (StepLoversMeet).
//
// They depend on the role list and the night only, so they run whether or not
// the curse was used, the Cursed wolf is dead or blocked: nothing leaks.
func engineSlots(s State, night int) []nightSlot {
	var out []nightSlot
	for id, role := range Registry {
		call := role.NightCall()
		if call == nil || s.Settings.Roles[id] < 1 || !call.CalledOn(night) {
			continue
		}
		if curse, ok := role.(CurseSource); ok && curse.CursesVillager() {
			out = append(out, nightSlot{StepPhoneCheck, call.Order, 1})
		}
		if love, ok := role.(LoveSource); ok && love.CreatesLovers() {
			out = append(out, nightSlot{StepPhoneCheck, call.Order, 1}, nightSlot{StepLoversMeet, call.Order, 2})
		}
	}
	return out
}

// NightStepDuration is how long the step being called lasts.
func NightStepDuration(s State, step RoleID) time.Duration {
	switch step {
	case StepPhoneCheck:
		return s.Settings.PhoneCheck
	case StepLoversMeet:
		return s.Settings.LoversMeet
	}
	return s.Settings.NightStep
}

func evolvePhoneCheck(s State, e Event) (State, bool) {
	ev, ok := e.(PhoneChecked)
	if !ok {
		return s, false
	}
	out := s.clone()
	if p := out.Players[ev.Player]; p != nil {
		delete(p.Flags, FlagTeamHidden)
		if ev.Love.InLove {
			SetFlag(p, FlagLoveKnown, 1)
		}
	}
	return out, true
}
