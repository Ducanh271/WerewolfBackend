package engine

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

// onlineTable is a lobby of n players switched to online, with roles chosen
// and nobody seated: what an online host has before pressing deal.
func onlineTable(t *testing.T, n int) State {
	t.Helper()
	s := apply(t, lobbyOf(t, n), ConfigureMode{Mode: Online})
	return apply(t, s, ConfigureRoles{Roles: map[RoleID]int{roleWerewolf: 1, roleSeer: 1, roleVillager: n - 2}})
}

func TestMode_DefaultsToAroundTheTable(t *testing.T) {
	if got := lobbyOf(t, 5).Settings.Mode; got != InPerson {
		t.Fatalf("default mode = %v, want InPerson", got)
	}
}

func TestMode_SwitchPutsSoundBackToTheModesDefault_HostMayChooseAgain(t *testing.T) {
	s := apply(t, lobbyOf(t, 5), ConfigureMode{Mode: Online})
	if s.Settings.Mode != Online || s.Settings.Sound != AllPhonesSpeak {
		t.Fatalf("online: mode %v, sound %v", s.Settings.Mode, s.Settings.Sound)
	}
	s = apply(t, s, ConfigureSound{Mode: OneSpeaker})
	if s.Settings.Sound != OneSpeaker {
		t.Fatalf("the host could not choose again: %v", s.Settings.Sound)
	}
	s = apply(t, s, ConfigureSound{Mode: AllPhonesSpeak})
	s = apply(t, s, ConfigureMode{Mode: InPerson})
	if s.Settings.Mode != InPerson || s.Settings.Sound != OneSpeaker {
		t.Fatalf("back in person: mode %v, sound %v", s.Settings.Mode, s.Settings.Sound)
	}
}

func TestMode_PickingTheSameModeChangesNothing(t *testing.T) {
	s := apply(t, lobbyOf(t, 5), ConfigureSound{Mode: AllPhonesSpeak})
	s = apply(t, s, SetSeats{Order: seatOrder(5)})
	evs, err := Decide(s, ConfigureMode{Mode: InPerson})
	if err != nil || len(evs) != 0 {
		t.Fatalf("same mode = %v, %v; want no events", evs, err)
	}
}

func TestMode_SwitchDropsTheSeats(t *testing.T) {
	s := apply(t, lobbyOf(t, 5), SetSeats{Order: seatOrder(5)})
	s = apply(t, s, ConfigureMode{Mode: Online})
	if len(s.Seats) != 0 {
		t.Fatalf("seats after the switch = %v", s.Seats)
	}
}

func TestMode_UnknownIsRefused(t *testing.T) {
	if _, err := Decide(lobbyOf(t, 5), ConfigureMode{Mode: Mode(99)}); !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("err = %v, want ErrInvalidMode", err)
	}
}

func TestMode_CannotChangeAfterTheDeal(t *testing.T) {
	s := apply(t, onlineTable(t, 5), DealRoles{Seed: 1})
	if _, err := Decide(s, ConfigureMode{Mode: InPerson}); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("err = %v, want ErrWrongPhase", err)
	}
}

func TestMode_ChangesAreToldToEverybody(t *testing.T) {
	log := []Event{ModeConfigured{Mode: Online}, SoundConfigured{Mode: AllPhonesSpeak}}
	got := Visible(NewState(), log, Viewer{Player: "p1"})
	if !reflect.DeepEqual(got, log) {
		t.Fatalf("visible = %v, want %v", got, log)
	}
}

func TestMode_InTheViewInEveryPhase(t *testing.T) {
	s := onlineTable(t, 5)
	if got := ViewFor(s, nil, Viewer{Player: "p1"}).Mode; got != Online {
		t.Errorf("lobby: mode = %v", got)
	}
	s = apply(t, s, DealRoles{Seed: 1})
	if got := ViewFor(s, nil, Viewer{Player: "p1"}).Mode; got != Online {
		t.Errorf("role view: mode = %v", got)
	}
}

func TestOnline_NoSeatingBeforeTheDeal(t *testing.T) {
	s := onlineTable(t, 5)
	if _, err := Decide(s, SetSeats{Order: seatOrder(5)}); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("err = %v, want ErrWrongPhase", err)
	}
}

func TestOnline_TheDealDrawsTheCircle(t *testing.T) {
	s := onlineTable(t, 8)
	evs, err := Decide(s, DealRoles{Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	seated, ok := evs[0].(SeatsSet)
	if !ok {
		t.Fatalf("first event = %T, want SeatsSet", evs[0])
	}
	got := slices.Clone(seated.Order)
	slices.Sort(got)
	want := seatOrder(8)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("circle = %v, want everybody once", seated.Order)
	}
	if after := Fold(s, evs...); !slices.Equal(after.Seats, seated.Order) || !after.Dealt {
		t.Fatalf("after the deal: seats %v, dealt %v", after.Seats, after.Dealt)
	}

	differs := false
	for seed := int64(1); seed < 20 && !differs; seed++ {
		differs = !slices.Equal(apply(t, s, DealRoles{Seed: seed}).Seats, apply(t, s, DealRoles{Seed: seed + 1}).Seats)
	}
	if !differs {
		t.Fatal("every seed draws the same circle")
	}
}

// One seed for both shuffles would apply one permutation to players and roles,
// so a player's role would follow from the order they joined in.
func TestOnline_RoleDoesNotFollowFromJoinOrder(t *testing.T) {
	s := onlineTable(t, 8)
	roles := map[RoleID]bool{}
	for seed := int64(1); seed <= 100; seed++ {
		roles[apply(t, s, DealRoles{Seed: seed}).Players["p1"].Role] = true
	}
	if len(roles) != 3 {
		t.Fatalf("over 100 deals the first to join only ever got %v", roles)
	}
}

func TestOnline_NewGameKeepsTheModeAndDrawsTheCircleAgain(t *testing.T) {
	s := apply(t, onlineTable(t, 5), ConfigureSound{Mode: OneSpeaker})
	s = apply(t, s, DealRoles{Seed: 1})
	s = Fold(s, GameEnded{Outcome: Outcome{Winner: TeamVillage}})
	s = apply(t, s, NewGame{})
	if s.Settings.Mode != Online || s.Settings.Sound != OneSpeaker {
		t.Fatalf("new game: mode %v, sound %v", s.Settings.Mode, s.Settings.Sound)
	}
	if len(s.Seats) != 0 {
		t.Fatalf("new game: seats %v, want none until the deal", s.Seats)
	}
	if s = apply(t, s, DealRoles{Seed: 2}); len(s.Seats) != 5 {
		t.Fatalf("second deal: seats %v", s.Seats)
	}
}

func TestInPerson_NewGameKeepsTheModeAndTheSeats(t *testing.T) {
	s := tableOf(t, 5, 1, false)
	s = apply(t, s, DealRoles{Seed: 1})
	s = Fold(s, GameEnded{Outcome: Outcome{Winner: TeamVillage}})
	s = apply(t, s, NewGame{})
	if s.Settings.Mode != InPerson || !slices.Equal(s.Seats, seatOrder(5)) {
		t.Fatalf("new game: mode %v, seats %v", s.Settings.Mode, s.Seats)
	}
}
