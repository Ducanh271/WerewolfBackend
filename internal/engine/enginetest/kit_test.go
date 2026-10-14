package enginetest

import (
	"testing"

	"masoi/internal/engine"
)

var fiveSeats = Seats("An", "Binh", "Chi", "Dung", "Em")

func fiveCast() Option {
	return Cast("An", Seer, "Binh", Werewolf, "Chi", Villager, "Dung", Villager, "Em", Villager)
}

func TestCast_GivesEachPlayerTheNamedRole(t *testing.T) {
	g := NewGame(t, fiveSeats, fiveCast())
	g.ExpectRole("An", Seer).ExpectRole("Binh", Werewolf).ExpectRole("Em", Villager)
	g.ExpectAlive("An", "Binh", "Chi", "Dung", "Em")
	if got := g.State().Players["Binh"].Team; got != engine.TeamWolves {
		t.Fatalf("Binh team = %v, want wolves", got)
	}
}

func TestCast_BreakingSetupRulesFailsTheScenario(t *testing.T) {
	cases := []struct {
		name  string
		want  string
		seats Option
		cast  Option
	}{
		{
			name:  "fewer than 5 players",
			want:  "too few players",
			seats: Seats("An", "Binh", "Chi", "Dung"),
			cast:  Cast("An", Seer, "Binh", Werewolf, "Chi", Villager, "Dung", Villager),
		},
		{
			name:  "two wolves at a 5-player table (limit is 1)",
			want:  "too many wolves",
			seats: fiveSeats,
			cast:  Cast("An", Seer, "Binh", Werewolf, "Chi", Werewolf, "Dung", Villager, "Em", Villager),
		},
		{
			name:  "no wolf at all",
			want:  "at least one wolf",
			seats: fiveSeats,
			cast:  Cast("An", Seer, "Binh", Guard, "Chi", Villager, "Dung", Villager, "Em", Villager),
		},
		{
			name:  "two Seers",
			want:  "only be dealt to one player",
			seats: fiveSeats,
			cast:  Cast("An", Seer, "Binh", Seer, "Chi", Werewolf, "Dung", Villager, "Em", Villager),
		},
		{
			name:  "Detective at an 8-player table",
			want:  "needs at least 9 players",
			seats: Seats("A", "B", "C", "D", "E", "F", "G", "H"),
			cast: Cast("A", Detective, "B", Werewolf, "C", Villager, "D", Villager,
				"E", Villager, "F", Villager, "G", Villager, "H", Villager),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			failsWith(t, tc.want, func(tb testing.TB) { NewGame(tb, tc.seats, tc.cast) })
		})
	}
}

func TestCast_MissingOrUnseatedPlayerFailsTheScenario(t *testing.T) {
	failsWith(t, "has no role", func(tb testing.TB) {
		NewGame(tb, fiveSeats, Cast("An", Seer, "Binh", Werewolf, "Chi", Villager, "Dung", Villager))
	})
	failsWith(t, "not in Seats", func(tb testing.TB) {
		NewGame(tb, fiveSeats, Cast("An", Seer, "Binh", Werewolf, "Chi", Villager, "Dung", Villager, "Zed", Villager))
	})
	failsWith(t, "two roles", func(tb testing.TB) {
		NewGame(tb, fiveSeats, Cast("An", Seer, "An", Werewolf))
	})
}

func TestRoles_DealsWithTheSeed(t *testing.T) {
	roles := Roles(map[engine.RoleID]int{Werewolf: 1, Seer: 1, Villager: 3})
	a := NewGame(t, fiveSeats, roles, Seed(7))
	b := NewGame(t, fiveSeats, roles, Seed(7))
	for _, name := range []string{"An", "Binh", "Chi", "Dung", "Em"} {
		if a.State().Players[a.Player(name)].Role != b.State().Players[b.Player(name)].Role {
			t.Fatalf("same seed dealt %s differently", name)
		}
	}
}

func TestDoExpectErr_RejectsWrongErrorAndAcceptedCommands(t *testing.T) {
	g := NewGame(t, fiveSeats, fiveCast())
	g.DoExpectErr(engine.DealRoles{Seed: 1}, engine.ErrAlreadyDealt)
	failsWith(t, "was accepted", func(tb testing.TB) {
		NewGame(tb, fiveSeats, fiveCast()).DoExpectErr(engine.JoinRoom{Player: "Zed", Nickname: "Zed"}, engine.ErrAlreadyDealt)
	})
	failsWith(t, "was rejected with", func(tb testing.TB) {
		NewGame(tb, fiveSeats, fiveCast()).DoExpectErr(engine.DealRoles{Seed: 1}, engine.ErrWrongPhase)
	})
}

func TestNight_TwoNightsInARowFailsTheScenario(t *testing.T) {
	failsWith(t, "between nights there is a day", func(tb testing.TB) {
		g := NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{}))
		g.Night()
		g.Dawn()
		g.Night()
	})
}

func TestNight_BeforeTheDayEndsFailsTheScenario(t *testing.T) {
	failsWith(t, "before the day ended", func(tb testing.TB) {
		g := NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{}))
		g.Night()
		g.Dawn()
		g.Day()
		g.Night()
	})
}

func TestNight_WithoutRolesOrWhileRunningFailsTheScenario(t *testing.T) {
	failsWith(t, "needs roles", func(tb testing.TB) { NewGame(tb, fiveSeats).Night() })
	failsWith(t, "a night is running", func(tb testing.TB) {
		g := NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{}))
		g.Night()
		g.Night()
	})
}

func TestDawnAndDay_OutOfOrderFailsTheScenario(t *testing.T) {
	failsWith(t, "no night is running", func(tb testing.TB) {
		NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{})).Dawn()
	})
	failsWith(t, "no night has passed", func(tb testing.TB) {
		NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{})).Day()
	})
}

func TestNightDayNight_AlternatingPlaysThrough(t *testing.T) {
	g := NewGame(t, fiveSeats, fiveCast(), WithDriver(fakeDriver{bitten: "Em"}))

	g.Night()
	g.Dawn().ExpectDead("Em").ExpectAnnouncement("dawn.someone_died")
	g.ExpectDead("Em").ExpectAlive("An", "Binh", "Chi", "Dung")
	g.Day().Hanged("Chi")
	g.ExpectDead("Chi")
	g.Night() // night 2 starts: the day ended
}

func TestDawn_ExpectationsFailWithReadableMessages(t *testing.T) {
	failsWith(t, "expected nobody to die in the night, but died: [Em]", func(tb testing.TB) {
		g := NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{bitten: "Em"}))
		g.Night()
		g.Dawn().ExpectNobodyDied()
	})
	failsWith(t, "expected announcement \"dawn.protected\"", func(tb testing.TB) {
		g := NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{bitten: "Em"}))
		g.Night()
		g.Dawn().ExpectAnnouncement("dawn.protected")
	})
	failsWith(t, "expected Em alive, but Em is dead", func(tb testing.TB) {
		g := NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{bitten: "Em"}))
		g.Night()
		g.Dawn()
		g.ExpectAlive("Em")
	})
}

func TestDefaultDriver_SaysPhasesAreNotBuiltYet(t *testing.T) {
	failsWith(t, "not built yet", func(tb testing.TB) { NewGame(tb, fiveSeats, fiveCast()).Night() })
}

func TestNightDo_SendsCommandsAndReportsRejections(t *testing.T) {
	g := NewGame(t, fiveSeats, fiveCast(), WithDriver(fakeDriver{}))
	g.Night().DoExpectErr(engine.DealRoles{Seed: 1}, engine.ErrAlreadyDealt)
	failsWith(t, "was rejected", func(tb testing.TB) {
		NewGame(tb, fiveSeats, fiveCast(), WithDriver(fakeDriver{})).Night().
			Do(engine.DealRoles{Seed: 1})
	})
}
