package room

import "masoi/internal/engine"

// Table is what a SpeakerPolicy needs to know about the room.
type Table struct {
	Host      engine.PlayerID
	Order     []engine.PlayerID // seat order once seats are set, join order before
	Connected func(engine.PlayerID) bool
}

// SpeakerPolicy chooses the phones that read announcements aloud. The others
// only show text. Playing in person uses one phone; a future online mode adds
// a policy where every phone speaks.
type SpeakerPolicy interface {
	Speakers(t Table) []engine.PlayerID
}

// InPersonSpeaker is the v1 policy: the host's phone speaks. If it is not
// connected, the next connected player after the host in Order takes over.
type InPersonSpeaker struct{}

// Speakers implements SpeakerPolicy.
func (InPersonSpeaker) Speakers(t Table) []engine.PlayerID {
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
