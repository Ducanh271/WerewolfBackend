package engine

import (
	"errors"
	"testing"
)

func TestLobby_LeaveRoom_RemovesPlayerAndClearsSeatsAndRoles(t *testing.T) {
	s := tableOf(t, 6, 1, false)
	s = apply(t, s, LeaveRoom{Player: "p3"})
	if _, still := s.Players["p3"]; still || len(s.Players) != 5 {
		t.Fatalf("p3 should be gone, players = %v", s.Players)
	}
	if len(s.Seats) != 0 || len(s.Settings.Roles) != 0 {
		t.Fatalf("seats and roles were made for the old table and must be cleared: %v %v", s.Seats, s.Settings.Roles)
	}
}

func TestLobby_LeaveRoom_Rejected(t *testing.T) {
	dealt := apply(t, tableOf(t, 6, 1, false), DealRoles{Seed: 1})
	tests := []struct {
		name  string
		state State
		who   PlayerID
		want  error
	}{
		{"after the deal", dealt, "p1", ErrWrongPhase},
		{"a stranger", lobbyOf(t, 5), "p9", ErrUnknownPlayer},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Decide(tc.state, LeaveRoom{Player: tc.who}); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
