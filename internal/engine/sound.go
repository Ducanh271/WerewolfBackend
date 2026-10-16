package engine

import "fmt"

// SoundMode picks who reads the announcements aloud (rules §1 "Âm thanh").
// Public announcements always go out as text to every phone; this only
// decides which phone's screen also says speaker: true.
type SoundMode int

const (
	// OneSpeaker (default): one phone speaks at a time, in person.
	OneSpeaker SoundMode = iota
	// AllPhonesSpeak: every phone speaks; what the online mode will use too.
	AllPhonesSpeak
)

func (m SoundMode) String() string {
	if m == AllPhonesSpeak {
		return "AllPhonesSpeak"
	}
	return "OneSpeaker"
}

// ConfigureSound picks the sound mode before the deal (rules §1, §2); the
// room checks that only the host may send it.
type ConfigureSound struct {
	CommandMarker
	Mode SoundMode
}

// SoundConfigured records the choice.
type SoundConfigured struct {
	EventMarker
	Mode SoundMode
}

func decideConfigureSound(s State, cmd ConfigureSound) ([]Event, error) {
	if s.Phase != PhaseLobby || s.Dealt {
		return nil, fmt.Errorf("%w: cannot change the sound mode during %s", ErrWrongPhase, s.Phase)
	}
	if cmd.Mode != OneSpeaker && cmd.Mode != AllPhonesSpeak {
		return nil, fmt.Errorf("%w: %d", ErrInvalidSoundMode, cmd.Mode)
	}
	return []Event{SoundConfigured{Mode: cmd.Mode}}, nil
}
