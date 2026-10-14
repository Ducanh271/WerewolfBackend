package engine

import (
	"errors"
	"fmt"
	"testing"
)

// apply decides a command and folds the resulting events, failing the test if
// the command is rejected.
func apply(t *testing.T, s State, c Command) State {
	t.Helper()
	evs, err := Decide(s, c)
	if err != nil {
		t.Fatalf("Decide(%T) = %v, want no error", c, err)
	}
	return Fold(s, evs...)
}

// lobbyWith returns a lobby where the given nicknames have joined, with player
// IDs p1, p2, ... in that order.
func lobbyWith(t *testing.T, nicknames ...string) State {
	t.Helper()
	s := NewState()
	for i, nickname := range nicknames {
		s = apply(t, s, JoinRoom{Player: playerID(i), Nickname: nickname})
	}
	return s
}

// lobbyOf returns a lobby of n joined players: p1 "P1", p2 "P2", ...
func lobbyOf(t *testing.T, n int) State {
	t.Helper()
	nicknames := make([]string, n)
	for i := range nicknames {
		nicknames[i] = fmt.Sprintf("P%d", i+1)
	}
	return lobbyWith(t, nicknames...)
}

func playerID(index int) PlayerID {
	return PlayerID(fmt.Sprintf("p%d", index+1))
}

func seatOrder(n int) []PlayerID {
	order := make([]PlayerID, n)
	for i := range order {
		order[i] = playerID(i)
	}
	return order
}

func TestJoinRoom_AddsPlayerToLobby(t *testing.T) {
	s := lobbyWith(t, "An", "  Bình  ")

	if got := len(s.Players); got != 2 {
		t.Fatalf("players = %d, want 2", got)
	}
	an := s.Players["p1"]
	if an.Nickname != "An" || !an.Alive || an.Team != TeamNone || an.Role != "" {
		t.Errorf("p1 = %+v, want nickname An, alive, no team, no role", an)
	}
	if got := s.Players["p2"].Nickname; got != "Bình" {
		t.Errorf("p2 nickname = %q, want trimmed %q", got, "Bình")
	}
	if s.Phase != PhaseLobby {
		t.Errorf("phase = %s, want Lobby", s.Phase)
	}
}

func TestJoinRoom_Rejects(t *testing.T) {
	joined := lobbyWith(t, "An", "Bình")

	tests := []struct {
		name string
		base State
		cmd  JoinRoom
		want error
	}{
		{"empty player id", joined, JoinRoom{Player: "", Nickname: "Chi"}, ErrInvalidPlayerID},
		{"empty nickname", joined, JoinRoom{Player: "p3", Nickname: ""}, ErrInvalidNickname},
		{"blank nickname", joined, JoinRoom{Player: "p3", Nickname: "   "}, ErrInvalidNickname},
		{"same player twice", joined, JoinRoom{Player: "p1", Nickname: "Khác"}, ErrAlreadyJoined},
		{"nickname taken", joined, JoinRoom{Player: "p3", Nickname: "An"}, ErrNicknameTaken},
		{"nickname taken, other case", joined, JoinRoom{Player: "p3", Nickname: " aN "}, ErrNicknameTaken},
		{"room full", lobbyOf(t, MaxPlayers), JoinRoom{Player: "p31", Nickname: "Muộn"}, ErrRoomFull},
		{"not in lobby", nightState(t), JoinRoom{Player: "p9", Nickname: "Muộn"}, ErrWrongPhase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decide(tt.base, tt.cmd)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Decide(JoinRoom) = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestSetSeats_StoresTheSeatingCircle(t *testing.T) {
	s := lobbyWith(t, "An", "Bình", "Chi")
	order := []PlayerID{"p3", "p1", "p2"}

	s = apply(t, s, SetSeats{Order: order})

	if len(s.Seats) != 3 || s.Seats[0] != "p3" || s.Seats[1] != "p1" || s.Seats[2] != "p2" {
		t.Fatalf("seats = %v, want %v", s.Seats, order)
	}

	// The event must own its slice: mutating the command afterwards cannot
	// reorder the seating.
	order[0] = "p2"
	if s.Seats[0] != "p3" {
		t.Errorf("seats aliased the command slice: %v", s.Seats)
	}
}

func TestSetSeats_Rejects(t *testing.T) {
	joined := lobbyWith(t, "An", "Bình", "Chi")

	tests := []struct {
		name string
		base State
		cmd  SetSeats
		want error
	}{
		{"missing a player", joined, SetSeats{Order: []PlayerID{"p1", "p2"}}, ErrSeatsMismatch},
		{"extra seat", joined, SetSeats{Order: []PlayerID{"p1", "p2", "p3", "p4"}}, ErrSeatsMismatch},
		{"stranger seated", joined, SetSeats{Order: []PlayerID{"p1", "p2", "p9"}}, ErrUnknownPlayer},
		{"player seated twice", joined, SetSeats{Order: []PlayerID{"p1", "p2", "p2"}}, ErrDuplicateSeat},
		{"not in lobby", nightState(t), SetSeats{Order: seatOrder(5)}, ErrWrongPhase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decide(tt.base, tt.cmd)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Decide(SetSeats) = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestDecide_UnknownCommandIsRejected(t *testing.T) {
	type unclaimedCommand struct{ CommandMarker }

	if _, err := Decide(NewState(), unclaimedCommand{}); !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("Decide(unclaimed) = %v, want %v", err, ErrUnknownCommand)
	}
}

func TestEvolve_UnknownEventLeavesStateUnchanged(t *testing.T) {
	type unclaimedEvent struct{ EventMarker }

	before := lobbyWith(t, "An")
	after := Evolve(before, unclaimedEvent{})

	if len(after.Players) != len(before.Players) || after.Players["p1"].Nickname != "An" {
		t.Fatalf("state changed: %+v", after)
	}
}

func TestEvolve_DoesNotMutateTheInputState(t *testing.T) {
	before := lobbyWith(t, "An")

	after := Fold(before,
		PlayerJoined{Player: "p2", Nickname: "Bình"},
		SeatsSet{Order: []PlayerID{"p1", "p2"}},
		RolesConfigured{Roles: map[RoleID]int{roleWerewolf: 1}},
	)

	if len(before.Players) != 1 {
		t.Errorf("input players = %d, want 1", len(before.Players))
	}
	if before.Seats != nil {
		t.Errorf("input seats = %v, want nil", before.Seats)
	}
	if before.Settings.Roles != nil {
		t.Errorf("input roles = %v, want nil", before.Settings.Roles)
	}
	if len(after.Players) != 2 || len(after.Seats) != 2 || len(after.Settings.Roles) != 1 {
		t.Errorf("folded state = %+v, want 2 players, 2 seats, 1 role", after)
	}
}

func TestNewState_HasDefaultTimers(t *testing.T) {
	s := NewState()

	if s.Settings.NightStep.Seconds() != 30 {
		t.Errorf("night step = %v, want 30s", s.Settings.NightStep)
	}
	if s.Settings.Day.Minutes() != 7 {
		t.Errorf("day = %v, want 7m", s.Settings.Day)
	}
	if s.Settings.Defense.Minutes() != 1 {
		t.Errorf("defense = %v, want 1m", s.Settings.Defense)
	}
	if s.Settings.Verdict.Seconds() != 30 {
		t.Errorf("verdict = %v, want 30s", s.Settings.Verdict)
	}
	if s.Settings.HunterShot.Seconds() != 15 {
		t.Errorf("hunter shot = %v, want 15s", s.Settings.HunterShot)
	}
}

// nightState returns a lobby of 5 players moved to the night phase, for
// "wrong phase" rows. Phase changes get their own command in a later task.
func nightState(t *testing.T) State {
	t.Helper()
	s := lobbyOf(t, 5)
	s.Phase = PhaseNight
	return s
}
