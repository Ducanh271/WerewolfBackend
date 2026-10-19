package engine

import (
	"errors"
	"testing"
)

// Rules §2 "Nickname, rời phòng và đá người": two names that look the same
// are the same name, however the phone typed the accents or the spaces.
const (
	binhNFC = "Bình"  // "ì" as one character
	binhNFD = "Bình" // "i" followed by a combining grave accent
)

func TestJoinRoom_NicknamesThatLookTheSame_AreTaken(t *testing.T) {
	tests := []struct {
		name    string
		already string
		joining string
	}{
		{"accents typed as one character, then as two", binhNFC, binhNFD},
		{"accents typed as two characters, then as one", binhNFD, binhNFC},
		{"double space inside", "An Bình", "An  Bình"},
		{"single space after a double one", "An  Bình", "An Bình"},
		{"accents and spaces together", "An " + binhNFC, " an   " + binhNFD + " "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := lobbyWith(t, tt.already)
			_, err := Decide(s, JoinRoom{Player: "p2", Nickname: tt.joining})
			if !errors.Is(err, ErrNicknameTaken) {
				t.Fatalf("Decide(JoinRoom{%q}) after %q = %v, want %v", tt.joining, tt.already, err, ErrNicknameTaken)
			}
		})
	}
}

func TestJoinRoom_NicknameIsKeptInItsNormalForm(t *testing.T) {
	tests := []struct {
		typed string
		kept  string
	}{
		{binhNFD, binhNFC},
		{"  An   Bình  ", "An Bình"},
		{"Nguyễn  Văn   An", "Nguyễn Văn An"},
	}
	for _, tt := range tests {
		t.Run(tt.typed, func(t *testing.T) {
			s := lobbyWith(t, tt.typed)
			if got := s.Players["p1"].Nickname; got != tt.kept {
				t.Errorf("nickname = %q, want %q", got, tt.kept)
			}
		})
	}
}

func TestJoinRoom_OtherBlanksAreStillRefused(t *testing.T) {
	for _, name := range []string{"An\tBình", "An Bình", "An\nBình"} {
		t.Run(name, func(t *testing.T) {
			if _, err := Decide(NewState(), JoinRoom{Player: "p1", Nickname: name}); !errors.Is(err, ErrInvalidNickname) {
				t.Fatalf("Decide(JoinRoom{%q}) = %v, want %v", name, err, ErrInvalidNickname)
			}
		})
	}
}
