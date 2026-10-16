package room

import "masoi/internal/engine"

// Table is what a SpeakerPolicy needs to know about the room.
type Table struct {
	Host      engine.PlayerID
	Order     []engine.PlayerID // seat order once seats are set, join order before
	Connected func(engine.PlayerID) bool
	// Current is who is speaking right now, "" if nobody is yet (rules §1
	// "Phòng ghi nhớ ai đang làm loa"): the room remembers, so a policy only
	// has to decide whether to keep it or hand it on.
	Current engine.PlayerID
}

// SpeakerPolicy chooses the phones that read announcements aloud. The others
// only show text. Playing in person uses one phone; the online mode (and the
// "Mọi máy cùng phát" room option, rules §1) uses AllPhones.
type SpeakerPolicy interface {
	Speakers(t Table) []engine.PlayerID
}

// InPersonSpeaker is the v1 policy: one phone speaks at a time. Whoever is
// currently reading (Table.Current) keeps it until that phone disconnects;
// only then does the next connected player after the host in Order take
// over. The host coming back mid-announcement does not take it back (rules
// §1).
type InPersonSpeaker struct{}

// Speakers implements SpeakerPolicy.
func (InPersonSpeaker) Speakers(t Table) []engine.PlayerID {
	if t.Current != "" && t.Connected(t.Current) {
		return []engine.PlayerID{t.Current}
	}
	start := 0
	for i, p := range t.Order {
		if p == t.Host {
			start = i
			break
		}
	}
	for k := 0; k < len(t.Order); k++ {
		p := t.Order[(start+k)%len(t.Order)]
		if t.Connected(p) {
			return []engine.PlayerID{p}
		}
	}
	return nil
}

// setSoundPolicy switches the room's SpeakerPolicy to match the host's
// ConfigureSound choice (rules §1).
func (r *Room) setSoundPolicy(mode engine.SoundMode) {
	if mode == engine.AllPhonesSpeak {
		r.pol = AllPhones{}
	} else {
		r.pol = InPersonSpeaker{}
	}
	r.speaker = ""
}

// AllPhones is the "Mọi máy cùng phát" option (rules §1) and what the online
// mode will use: every phone reads the announcements aloud, so every view
// carries speaker: true.
type AllPhones struct{}

// Speakers implements SpeakerPolicy.
func (AllPhones) Speakers(t Table) []engine.PlayerID {
	return t.Order
}
