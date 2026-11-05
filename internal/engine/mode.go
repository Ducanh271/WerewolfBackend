package engine

import "fmt"

// Mode is how a room plays (rules §1b). The whole room follows one mode. It
// changes no game rule: it picks the default sound, whether seats are drawn
// at the deal, and how the app looks and sounds.
type Mode int

// Modes.
const (
	InPerson Mode = iota
	Online
)

func (m Mode) String() string {
	if m == Online {
		return "Online"
	}
	return "InPerson"
}

// defaultSound is the sound a room falls back to on entering this mode.
func (m Mode) defaultSound() SoundMode {
	if m == Online {
		return AllPhonesSpeak
	}
	return OneSpeaker
}

// ConfigureMode picks the mode before the deal (rules §1b); the room checks
// that only the host may send it.
type ConfigureMode struct {
	CommandMarker
	Mode Mode
}

// ModeConfigured records the choice. Any seats set so far are gone: the two
// modes seat people differently (rules §1b).
type ModeConfigured struct {
	EventMarker
	Mode Mode
}

func decideConfigureMode(s State, cmd ConfigureMode) ([]Event, error) {
	if s.Phase != PhaseLobby || s.Dealt {
		return nil, fmt.Errorf("%w: cannot change the mode during %s", ErrWrongPhase, s.Phase)
	}
	if cmd.Mode != InPerson && cmd.Mode != Online {
		return nil, fmt.Errorf("%w: %d", ErrInvalidMode, cmd.Mode)
	}
	if cmd.Mode == s.Settings.Mode {
		return nil, nil // not a change: the host's sound choice stays
	}
	return []Event{
		ModeConfigured{Mode: cmd.Mode},
		SoundConfigured{Mode: cmd.Mode.defaultSound()},
	}, nil
}
