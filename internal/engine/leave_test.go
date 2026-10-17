package engine

import (
	"errors"
	"reflect"
	"testing"
)

func TestLobby_LeaveRoom_DropsOnlyThatSeatKeepsRoles(t *testing.T) {
	s := tableOf(t, 6, 1, false)
	wantRoles := copyRoles(s.Settings.Roles)

	s = apply(t, s, LeaveRoom{Player: "p3"})

	if _, still := s.Players["p3"]; still || len(s.Players) != 5 {
		t.Fatalf("p3 should be gone, players = %v", s.Players)
	}
	wantSeats := []PlayerID{"p1", "p2", "p4", "p5", "p6"}
	if !reflect.DeepEqual(s.Seats, wantSeats) {
		t.Fatalf("seats = %v, want %v (p3 dropped, the rest keep their order)", s.Seats, wantSeats)
	}
	if !reflect.DeepEqual(s.Settings.Roles, wantRoles) {
		t.Fatalf("role list = %v, want %v untouched", s.Settings.Roles, wantRoles)
	}

	// The role list no longer fits 5 players (it still has 6 slots): Deal is
	// blocked and says why, instead of the lobby silently losing its setup
	// (audit K2).
	if _, err := Decide(s, DealRoles{Seed: 1}); !errors.Is(err, ErrRoleCountMismatch) {
		t.Fatalf("Decide(DealRoles) = %v, want %v", err, ErrRoleCountMismatch)
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

func TestLobby_KickPlayer_DropsOnlyThatSeatKeepsRoles(t *testing.T) {
	s := tableOf(t, 6, 1, false)
	wantRoles := copyRoles(s.Settings.Roles)

	s = apply(t, s, KickPlayer{Target: "p3"})

	if _, still := s.Players["p3"]; still || len(s.Players) != 5 {
		t.Fatalf("p3 should be gone, players = %v", s.Players)
	}
	wantSeats := []PlayerID{"p1", "p2", "p4", "p5", "p6"}
	if !reflect.DeepEqual(s.Seats, wantSeats) {
		t.Fatalf("seats = %v, want %v (p3 dropped, the rest keep their order)", s.Seats, wantSeats)
	}
	if !reflect.DeepEqual(s.Settings.Roles, wantRoles) {
		t.Fatalf("role list = %v, want %v untouched", s.Settings.Roles, wantRoles)
	}
}

func TestLobby_KickPlayer_Rejected(t *testing.T) {
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
			if _, err := Decide(tc.state, KickPlayer{Target: tc.who}); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}
