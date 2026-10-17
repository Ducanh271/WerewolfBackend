package engine

import (
	"errors"
	"testing"
)

func TestNewGame_OnlyAllowedOnceTheGameEnded(t *testing.T) {
	for _, phase := range []Phase{PhaseLobby, PhaseRoleView, PhaseNight, PhaseDay} {
		s := NewState()
		s.Phase = phase
		if _, err := Decide(s, NewGame{}); !errors.Is(err, ErrWrongPhase) {
			t.Errorf("phase %s: err = %v, want ErrWrongPhase", phase, err)
		}
	}
}

// TestNewGame_KeepsPlayersSeatsAndRoles_ResetsTheMatch covers audit K3: the
// table does not have to recreate the room for a second game.
func TestNewGame_KeepsPlayersSeatsAndRoles_ResetsTheMatch(t *testing.T) {
	s := lobbyOf(t, 5)
	order := seatOrder(5)
	s = apply(t, s, SetSeats{Order: order})
	roles := map[RoleID]int{"Villager": 4, "Werewolf": 1}
	s = apply(t, s, ConfigureRoles{Roles: roles})
	s = apply(t, s, DealRoles{Seed: 1})
	// End the game however it got there (a win, a draw, an abort — NewGame
	// does not care which).
	winner := TeamVillage
	s = Fold(s, GameEnded{Outcome: Outcome{Winner: winner, Winners: order}})
	if s.Phase != PhaseEnded {
		t.Fatalf("setup broken: phase = %s, want Ended", s.Phase)
	}

	s = apply(t, s, NewGame{})

	if s.Phase != PhaseLobby {
		t.Fatalf("phase = %s, want Lobby", s.Phase)
	}
	if s.Dealt {
		t.Fatalf("dealt = true, want false: the host must deal again")
	}
	if len(s.Seats) != 5 {
		t.Fatalf("seats = %v, want the 5 kept from before", s.Seats)
	}
	for i, id := range order {
		if s.Seats[i] != id {
			t.Fatalf("seats = %v, want %v unchanged", s.Seats, order)
		}
	}
	if len(s.Settings.Roles) != 2 || s.Settings.Roles["Villager"] != 4 || s.Settings.Roles["Werewolf"] != 1 {
		t.Fatalf("role list = %v, want kept as %v", s.Settings.Roles, roles)
	}
	if s.Winner != nil || len(s.Winners) != 0 {
		t.Fatalf("winner/winners = %v/%v, want cleared", s.Winner, s.Winners)
	}
	if s.Night != 0 || s.NightStep != 0 {
		t.Fatalf("night/step = %d/%d, want 0/0", s.Night, s.NightStep)
	}
	if s.Chief != "" {
		t.Fatalf("chief = %s, want cleared", s.Chief)
	}
	for _, id := range order {
		p := s.Players[id]
		if p == nil {
			t.Fatalf("player %s missing after NewGame", id)
		}
		if p.Role != "" || p.Team != TeamNone || !p.Alive || len(p.Flags) != 0 {
			t.Errorf("player %s = %+v, want a fresh lobby player", id, p)
		}
	}
}

// TestNewGame_HostCanDealAgainRightAway: no SetSeats/ConfigureRoles needed
// for the second game, since NewGame kept them.
func TestNewGame_HostCanDealAgainRightAway(t *testing.T) {
	s := lobbyOf(t, 5)
	s = apply(t, s, SetSeats{Order: seatOrder(5)})
	s = apply(t, s, ConfigureRoles{Roles: map[RoleID]int{"Villager": 4, "Werewolf": 1}})
	s = apply(t, s, DealRoles{Seed: 1})
	s = Fold(s, GameEnded{Outcome: Outcome{Draw: true}})
	s = apply(t, s, NewGame{})

	if _, err := Decide(s, DealRoles{Seed: 2}); err != nil {
		t.Fatalf("DealRoles after NewGame = %v, want no error", err)
	}
}
