package engine

import "masoi/internal/i18n"

// StepLoversMeet is the night-1 step in which the two lovers open their eyes
// and look at each other (rules §3 "0c", §7). It belongs to no role: it follows
// the phone check that comes right after Cupid's step, on the nights a
// LoveSource is called, so it runs on a fixed schedule and tells nothing.
// Nobody acts in it; only the lovers' phones show it as their turn.
const StepLoversMeet RoleID = "LoversMeet"

// LoversMeetAnnounced is the voice line that opens the step, read aloud by the
// speaker phone. Public: it is the same on every phone.
type LoversMeetAnnounced struct {
	EventMarker
}

// AnnouncementKey is read aloud to everybody.
func (LoversMeetAnnounced) AnnouncementKey() string { return i18n.NightLoversMeet }

// loversMeetTurn is the turn the lovers' phones show while the step runs; nil
// for anybody else.
func loversMeetTurn(s State, p *Player) *TurnView {
	if PartnerOf(s, p.ID) == "" {
		return nil
	}
	return &TurnView{Step: StepLoversMeet}
}
