package engine

import (
	"errors"
	"testing"
)

func TestSound_DefaultIsOneSpeaker(t *testing.T) {
	if got := NewState().Settings.Sound; got != OneSpeaker {
		t.Fatalf("default sound mode = %v, want OneSpeaker", got)
	}
}

func TestSound_ConfigureStoresTheChoice(t *testing.T) {
	s := apply(t, lobbyOf(t, 5), ConfigureSound{Mode: AllPhonesSpeak})
	if s.Settings.Sound != AllPhonesSpeak {
		t.Fatalf("sound mode = %v, want AllPhonesSpeak", s.Settings.Sound)
	}
}

func TestSound_RejectsUnknownMode(t *testing.T) {
	_, err := Decide(lobbyOf(t, 5), ConfigureSound{Mode: SoundMode(99)})
	if !errors.Is(err, ErrInvalidSoundMode) {
		t.Fatalf("err = %v, want ErrInvalidSoundMode", err)
	}
}

func TestSound_RejectedAfterTheDeal(t *testing.T) {
	s := tableOf(t, 5, 1, false)
	s = apply(t, s, DealRoles{Seed: 1})
	if _, err := Decide(s, ConfigureSound{Mode: AllPhonesSpeak}); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("err = %v, want ErrWrongPhase", err)
	}
}
