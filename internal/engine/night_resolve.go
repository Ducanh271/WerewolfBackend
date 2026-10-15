package engine

import "masoi/internal/i18n"

// Resolution is the night being worked out at dawn (rules §9). Each pipeline
// step reads the choices in State and adds to it. Roles of later tasks hook in
// through new fields and new steps, never by reordering the existing ones.
type Resolution struct {
	State     State
	Bitten    PlayerID // the pack's target, empty if nobody
	Protected bool     // the Guard shielded the bitten player
	Deaths    []PlayerID
}

// nightPipeline is the ordered list of rules §9. Only step 3 exists so far; the
// rest are added by their tasks in the numbered place:
// 1 Visitor, 2 Curse, 4 RustyKnight, 5 Witch poison, 6 delayed wolf deaths,
// 7 lover deaths. Step 8 (announce) is resolveNight below, step 9 (win check)
// arrives with T8.
var nightPipeline = []func(*Resolution){
	stepBite, // 3
}

// stepBite: the pack's bite kills unless the Guard shields the bitten player.
func stepBite(r *Resolution) {
	target, ok := BiteTarget(r.State, WolfVoters(r.State))
	if !ok {
		return
	}
	r.Bitten = target
	if r.State.Pending.Protected == target {
		r.Protected = true
		return
	}
	r.Deaths = append(r.Deaths, target)
}

// resolveNight turns the night's choices into deaths and the morning
// announcement. s is the state when the last step runs out.
func resolveNight(s State) []Event {
	r := &Resolution{State: s}
	for _, step := range nightPipeline {
		step(r)
	}
	var evs []Event
	for _, id := range r.Deaths {
		evs = append(evs, PlayerDied{Player: id})
	}
	return append(evs, announceDawn(r)) // step 8
}

func announceDawn(r *Resolution) Event {
	switch {
	case len(r.Deaths) > 0:
		return DawnAnnounced{Kind: DawnDeaths, Dead: inSeatOrder(r.State.Seats, r.Deaths)}
	case r.Protected:
		return DawnAnnounced{Kind: DawnProtected}
	default:
		return DawnAnnounced{Kind: DawnNobodyDied}
	}
}

// PlayerDied records a death. The role is not part of the announcement.
type PlayerDied struct {
	EventMarker
	Player PlayerID
}

// DawnKind is which of the three morning announcements is read (rules §4).
type DawnKind int

// The three morning announcements.
const (
	DawnDeaths     DawnKind = iota // names of the dead, no roles
	DawnProtected                  // the Guard shielded the bitten player
	DawnNobodyDied                 // nobody died for any other reason
)

// DawnAnnounced is the morning announcement for the narrator.
type DawnAnnounced struct {
	EventMarker
	Kind DawnKind
	Dead []PlayerID
}

// AnnouncementKey is the i18n key of the message to read aloud.
func (e DawnAnnounced) AnnouncementKey() string {
	switch e.Kind {
	case DawnDeaths:
		return i18n.DawnSomeoneDied
	case DawnProtected:
		return i18n.DawnProtected
	default:
		return i18n.DawnNobodyDied
	}
}

func evolveResolution(s State, e Event) (State, bool) {
	if ev, ok := e.(PlayerDied); ok {
		out := s.clone()
		if p := out.Players[ev.Player]; p != nil {
			p.Alive = false
		}
		return out, true
	}
	return s, false
}

// inSeatOrder returns ids in seating order, so the narrator never reveals the
// order in which the pipeline killed them (rules §4). The input is untouched.
func inSeatOrder(seats, ids []PlayerID) []PlayerID {
	want := make(map[PlayerID]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	out := make([]PlayerID, 0, len(ids))
	for _, id := range seats {
		if want[id] {
			out = append(out, id)
		}
	}
	return out
}
