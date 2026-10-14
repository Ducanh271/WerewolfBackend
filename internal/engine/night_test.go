package engine

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

// dealtTable returns a table of n players dealt with the given roles (Villager
// fills the rest), now in the role-view phase.
func dealtTable(t *testing.T, n int, roles map[RoleID]int) State {
	t.Helper()
	cfg := map[RoleID]int{}
	rest := n
	for r, c := range roles {
		cfg[r] = c
		rest -= c
	}
	cfg[roleVillager] = rest
	s := apply(t, lobbyOf(t, n), SetSeats{Order: seatOrder(n)})
	s = apply(t, s, ConfigureRoles{Roles: cfg})
	return apply(t, s, DealRoles{Seed: 1})
}

// smallTable is 7 players: 2 wolves, a Seer and villagers.
func smallTable(t *testing.T) State {
	return dealtTable(t, 7, map[RoleID]int{roleWerewolf: 2, roleSeer: 1})
}

// bigTable is 14 players with every role that has a night call, so the call
// order and the early-night limits can be checked.
func bigTable(t *testing.T) State {
	return dealtTable(t, 14, map[RoleID]int{
		roleWerewolf: 1, roleCursedWolf: 1, roleWolfSeer: 1, roleCupid: 1, roleSeer: 1, roleDetective: 1,
	})
}

// timeout sends the timeout the room would send for the current phase.
func timeout(t *testing.T, s State) (State, []Event) {
	t.Helper()
	evs, err := Decide(s, PhaseTimeout{Phase: s.Phase, Seq: s.Seq})
	if err != nil {
		t.Fatalf("PhaseTimeout rejected: %v", err)
	}
	return Fold(s, evs...), evs
}

// playNight runs a state from role view through night `night`'s end,
// returning every event the narrator sees. Earlier nights are skipped by
// setting the counter directly: day phases are not built yet (T7).
func playNight(t *testing.T, s State, night int) []Event {
	t.Helper()
	var all []Event
	if s.Phase == PhaseRoleView {
		var evs []Event
		s, evs = timeout(t, s)
		all = append(all, evs...)
	}
	if night != 1 {
		s.Night = night
		all = all[:0]
		all = append(all, NightStarted{Night: night}, nextNightEvent(s, night, 0))
		s = Fold(s, all...)
	}
	for s.Phase == PhaseNight {
		var evs []Event
		s, evs = timeout(t, s)
		all = append(all, evs...)
	}
	if s.Phase != PhaseDawn {
		t.Fatalf("night ended in %s, want Dawn", s.Phase)
	}
	return all
}

func calledRoles(events []Event) []RoleID {
	var roles []RoleID
	for _, e := range events {
		if c, ok := e.(NightStepCalled); ok {
			roles = append(roles, c.Role)
		}
	}
	return roles
}

func TestDealRoles_StartsRoleViewWithCountdown(t *testing.T) {
	s := smallTable(t)
	if s.Phase != PhaseRoleView {
		t.Fatalf("phase = %s, want RoleView", s.Phase)
	}
	if s.Settings.RoleView != 30*time.Second {
		t.Fatalf("role view = %v, want 30s default", s.Settings.RoleView)
	}
	if _, err := Decide(s, JoinRoom{Player: "late", Nickname: "Late"}); !errors.Is(err, ErrWrongPhase) {
		t.Fatalf("join after the deal: err = %v, want %v", err, ErrWrongPhase)
	}
}

func TestRoleView_TimeoutStartsNightOne(t *testing.T) {
	s, _ := timeout(t, smallTable(t))
	if s.Phase != PhaseNight || s.Night != 1 {
		t.Fatalf("after role view: phase %s night %d, want Night 1", s.Phase, s.Night)
	}
	if role, ok := CurrentNightRole(s); !ok || role != roleWerewolf {
		t.Fatalf("current step = %q, %v; want Werewolf, the first role in the game", role, ok)
	}
}

func TestNight_OrderFollowsRulesTable(t *testing.T) {
	got := calledRoles(playNight(t, bigTable(t), 1))
	want := []RoleID{roleCupid, roleCursedWolf, roleWolfSeer, roleWerewolf, roleSeer, roleDetective}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("night 1 calls %v, want %v", got, want)
	}
}

func TestNight_CursedWolfAndCupidOnlyEarlyNights(t *testing.T) {
	tests := []struct {
		night int
		want  []RoleID
	}{
		{1, []RoleID{roleCupid, roleCursedWolf, roleWolfSeer, roleWerewolf, roleSeer, roleDetective}},
		{3, []RoleID{roleCursedWolf, roleWolfSeer, roleWerewolf, roleSeer, roleDetective}},
		{4, []RoleID{roleWolfSeer, roleWerewolf, roleSeer, roleDetective}},
	}
	for _, tt := range tests {
		got := calledRoles(playNight(t, bigTable(t), tt.night))
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("night %d calls %v, want %v", tt.night, got, tt.want)
		}
	}
}

func TestNight_OnlyRolesInTheGameAreCalled(t *testing.T) {
	// The role list is public, so a role nobody holds is never called.
	got := calledRoles(playNight(t, smallTable(t), 1))
	want := []RoleID{roleWerewolf, roleSeer}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("night 1 calls %v, want %v", got, want)
	}
}

func TestNight_WolfPackAlwaysCalledEvenWithoutPlainWerewolf(t *testing.T) {
	tests := []struct {
		name  string
		wolfs map[RoleID]int
		want  []RoleID
	}{
		{"only a cursed wolf", map[RoleID]int{roleCursedWolf: 1}, []RoleID{roleCursedWolf, roleWerewolf}},
		{"only a wolf seer", map[RoleID]int{roleWolfSeer: 1}, []RoleID{roleWolfSeer, roleWerewolf}},
		{"all three, in the agreed order", map[RoleID]int{roleWerewolf: 1, roleCursedWolf: 1, roleWolfSeer: 1},
			[]RoleID{roleCursedWolf, roleWolfSeer, roleWerewolf}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calledRoles(playNight(t, dealtTable(t, 14, tt.wolfs), 1)); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("night 1 calls %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNight_EveryCallRunsFullTime(t *testing.T) {
	s := smallTable(t)
	s.Settings.NightStep = 45 * time.Second
	calls := 0
	for _, e := range playNight(t, s, 1) {
		if c, ok := e.(NightStepCalled); ok {
			calls++
			if c.Duration != 45*time.Second {
				t.Errorf("%s called for %v, want the full 45s", c.Role, c.Duration)
			}
		}
	}
	if calls == 0 {
		t.Fatal("no role was called")
	}
}

func TestNight_DeadRoleIsCalledLikeALivingOne(t *testing.T) {
	living := bigTable(t)
	dead := bigTable(t)
	for _, p := range dead.Players {
		if p.Role == roleSeer || p.Role == roleDetective {
			p.Alive = false
		}
	}
	if got, want := playNight(t, dead, 1), playNight(t, living, 1); !reflect.DeepEqual(got, want) {
		t.Fatalf("with dead holders the narrator sees %v, but with living ones %v", got, want)
	}
}

func TestNight_TimeoutMeansNoActionAndNightEndsInDawn(t *testing.T) {
	s := smallTable(t)
	events := playNight(t, s, 1)
	if _, ok := events[len(events)-1].(NightEnded); !ok {
		t.Fatalf("last event = %T, want NightEnded", events[len(events)-1])
	}
	for _, p := range Fold(s, events...).Players {
		if !p.Alive {
			t.Errorf("%s died although nobody acted", p.ID)
		}
	}
}

func TestPhaseTimeout_StaleOrWrongPhaseIsIgnored(t *testing.T) {
	roleView := smallTable(t)
	night, _ := timeout(t, roleView)
	tests := []struct {
		name  string
		state State
		cmd   PhaseTimeout
	}{
		{"old seq in role view", roleView, PhaseTimeout{Phase: PhaseRoleView, Seq: roleView.Seq - 1}},
		{"role view timeout arrives in night", night, PhaseTimeout{Phase: PhaseRoleView, Seq: roleView.Seq}},
		{"previous step arrives late", night, PhaseTimeout{Phase: PhaseNight, Seq: night.Seq - 1}},
		{"lobby has no deadline", lobbyOf(t, 5), PhaseTimeout{Phase: PhaseLobby}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evs, err := Decide(tt.state, tt.cmd)
			if err != nil || len(evs) != 0 {
				t.Fatalf("got %v, %v; want nothing", evs, err)
			}
		})
	}
}

func TestCurrentNightRole_FollowsTheStep(t *testing.T) {
	s := bigTable(t)
	if _, ok := CurrentNightRole(s); ok {
		t.Fatal("no night step is running during role view")
	}
	s, _ = timeout(t, s)
	var seen []RoleID
	for role, ok := CurrentNightRole(s); ok; role, ok = CurrentNightRole(s) {
		seen = append(seen, role)
		s, _ = timeout(t, s)
	}
	if want := []RoleID{roleCupid, roleCursedWolf, roleWolfSeer, roleWerewolf, roleSeer, roleDetective}; !reflect.DeepEqual(seen, want) {
		t.Fatalf("steps %v, want %v", seen, want)
	}
}
