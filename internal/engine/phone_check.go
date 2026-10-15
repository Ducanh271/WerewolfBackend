package engine

import (
	"time"

	"masoi/internal/i18n"
)

// StepPhoneCheck is the night step after the curse step (rules §3 "Kiểm tra
// điện thoại"): everybody looks at their phone. It belongs to no role, and it
// is called on every early night of a game whose role list has a CurseSource,
// whether or not the curse is still to come, so it tells nothing.
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

// PhoneChecked is the notice one living player gets at the phone check. Both
// answers have this exact shape; only Cursed differs.
type PhoneChecked struct {
	EventMarker
	Player PlayerID
	Cursed bool
}

// PhoneCheckView is that notice as the phone shows it, while the step runs.
type PhoneCheckView struct {
	Cursed bool
}

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
			evs = append(evs, PhoneChecked{Player: id, Cursed: IsCursed(p)})
		}
	}
	return evs
}

// phoneCheckAfter returns the order of the curse step the phone check follows,
// when the game has one on this night.
func phoneCheckAfter(s State, night int) (order int, ok bool) {
	for id, role := range Registry {
		curse, isCurse := role.(CurseSource)
		call := role.NightCall()
		if !isCurse || !curse.CursesVillager() || call == nil || s.Settings.Roles[id] < 1 || !call.CalledOn(night) {
			continue
		}
		return call.Order, true
	}
	return 0, false
}

// NightStepDuration is how long the step being called lasts.
func NightStepDuration(s State, step RoleID) time.Duration {
	if step == StepPhoneCheck {
		return s.Settings.PhoneCheck
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
	}
	return out, true
}
