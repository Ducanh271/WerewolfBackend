package engine

import (
	"errors"
	"testing"
)

// Blocked behaviour of the roles that are not built yet (CursedWolf T15,
// WolfSeer T16, Detective T17), on hand-built state and the fake roles: those
// roles only use the shared hooks tested here (NightActor, WolfVoters,
// NightCall.Nights) and keep their charges in their own Flags.

// nightAt is a night where the given role is being called.
func nightAt(t *testing.T, night int, step RoleID, roles map[RoleID]int) State {
	t.Helper()
	s := NewState()
	s.Phase = PhaseNight
	s.Night = night
	s.Settings.Roles = roles
	s.Seats = []PlayerID{"W", "S", "C", "D", "V"}
	teams := map[PlayerID]Team{"W": TeamWolves, "S": TeamWolves, "C": TeamWolves, "D": TeamVillage, "V": TeamVillage}
	holds := map[PlayerID]RoleID{"W": roleWerewolf, "S": roleWolfSeer, "C": roleCursedWolf, "D": roleDetective, "V": roleVillager}
	for _, id := range s.Seats {
		s.Players[id] = &Player{ID: id, Role: holds[id], Team: teams[id], Alive: true, Flags: map[string]int{}}
	}
	for i, id := range nightSteps(s, night) {
		if id == step {
			s.NightStep = i + 1
			return s
		}
	}
	t.Fatalf("%s is not called on night %d", step, night)
	return s
}

var allNightRoles = map[RoleID]int{roleWerewolf: 1, roleWolfSeer: 1, roleCursedWolf: 1, roleDetective: 1, roleVillager: 1}

func TestBlocked_AnyBlockerMakesEveryRoleRefuseWithErrBlocked(t *testing.T) {
	tests := []struct {
		name   string
		role   RoleID
		holder PlayerID
	}{
		{"wolf seer: loses the scan", roleWolfSeer, "S"},
		{"cursed wolf", roleCursedWolf, "C"},
		{"detective", roleDetective, "D"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := nightAt(t, 1, tc.role, allNightRoles)
			if _, err := NightActor(s, tc.holder, tc.role); err != nil {
				t.Fatalf("not blocked yet: %v", err)
			}
			Block(&s, tc.holder, "SomeFutureBlocker") // any role, not only the Visitor
			if _, err := NightActor(s, tc.holder, tc.role); !errors.Is(err, ErrBlocked) {
				t.Fatalf("err = %v, want ErrBlocked", err)
			}
		})
	}
}

func TestBlocked_WolfSeerLosesTheBiteVoteToo(t *testing.T) {
	s := nightAt(t, 1, roleWerewolf, allNightRoles)
	Block(&s, "S", roleSeer)
	voters := WolfVoters(s)
	if len(voters) != 2 || voters[0] != "W" || voters[1] != "C" {
		t.Fatalf("voters = %v, want [W C]", voters)
	}
	if _, err := NightActor(s, "S", roleWerewolf); !errors.Is(err, ErrBlocked) {
		t.Fatalf("a blocked wolf seer at the pack step: err = %v, want ErrBlocked", err)
	}
}

func TestBlocked_NightEndsOnTheNextNightStart(t *testing.T) {
	s := nightAt(t, 1, roleDetective, allNightRoles)
	Block(&s, "D", roleSeer)
	s = Fold(s, NightStarted{Night: 2})
	if Blocked(s, "D") {
		t.Fatal("the block must not carry over to the next night")
	}
}

func TestBlocked_DetectiveKeepsTheChargeAndIsCalledAgain(t *testing.T) {
	s := nightAt(t, 2, roleDetective, allNightRoles)
	Block(&s, "D", roleSeer)
	if _, err := NightActor(s, "D", roleDetective); !errors.Is(err, ErrBlocked) {
		t.Fatalf("err = %v, want ErrBlocked", err)
	}
	if s.Players["D"].Flags["detective.used"] != 0 {
		t.Fatal("a refused action must not use up the charge")
	}
	// Called "every night until used": the next night he is called again.
	if got := nightSteps(s, 3); got[len(got)-1] != roleDetective {
		t.Fatalf("night 3 steps %v: the Detective must still be called", got)
	}
}

func TestBlocked_CursedWolfBlockedOnNight1Or2_CalledAgain_OnNight3_NeverCalledAgain(t *testing.T) {
	for _, night := range []int{1, 2} {
		s := nightAt(t, night, roleCursedWolf, allNightRoles)
		Block(&s, "C", roleSeer)
		if _, err := NightActor(s, "C", roleCursedWolf); !errors.Is(err, ErrBlocked) {
			t.Fatalf("night %d: err = %v, want ErrBlocked", night, err)
		}
		if !contains(nightSteps(s, night+1), roleCursedWolf) {
			t.Fatalf("blocked on night %d: the curse must be called again on night %d", night, night+1)
		}
	}
	// Night 3 is the last call (NightCall.Nights = 1..3): a block that night
	// loses the curse for good, without ever calling a fourth night.
	s := nightAt(t, 3, roleCursedWolf, allNightRoles)
	Block(&s, "C", roleSeer)
	if _, err := NightActor(s, "C", roleCursedWolf); !errors.Is(err, ErrBlocked) {
		t.Fatalf("night 3: err = %v, want ErrBlocked", err)
	}
	for night := 4; night <= 6; night++ {
		if contains(nightSteps(s, night), roleCursedWolf) {
			t.Fatalf("the cursed wolf is called on night %d", night)
		}
	}
}

func TestBlocked_ViewNeverShowsTheBlockerInSecrets(t *testing.T) {
	s := nightAt(t, 1, roleDetective, allNightRoles)
	Block(&s, "D", roleSeer)
	if got := secrets(s); got.Pending.Blocked != nil {
		t.Fatalf("secrets leak who blocked whom: %v", got.Pending.Blocked)
	}
}

func contains(ids []RoleID, want RoleID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
