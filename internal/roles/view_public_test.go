package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

// The usual table: wolves An, Binh, Chi; Guard Dung; Seer Em; villagers Phuc,
// Giang, Hoa, Khoa, Lan.
var cast = map[string]engine.RoleID{
	"An": Werewolf, "Binh": Werewolf, "Chi": Werewolf, "Dung": Guard, "Em": Seer,
	"Phuc": Villager, "Giang": Villager, "Hoa": Villager, "Khoa": Villager, "Lan": Villager,
}

func TestView_PublicPart_IsTheSameOnEveryPhone(t *testing.T) {
	step := engine.DefaultSettings().NightStep
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"a night call is a role and a duration, on every phone", func(t *testing.T) {
			g := newTable(t)
			g.Night().TurnOf(Guard)
			for name := range cast {
				g.As(name).NightCall(Guard, step)
			}
		}},
		{"a fake call looks exactly like a real one", func(t *testing.T) {
			g := newTable(t)
			g.Night().TurnOf(Guard)
			real := g.As("Hoa").View().NightCall
			g.Dawn()
			g.Day().Hanged("Dung") // the Guard is dead: her call is fake now
			g.Night().TurnOf(Guard)
			for _, name := range []string{"An", "Em", "Hoa", "Phuc"} {
				g.As(name).NightCall(real.Role, real.Duration)
			}
		}},
		{"the called role differs from the others only in its own turn card", func(t *testing.T) {
			g := newTable(t)
			g.Night().TurnOf(Guard)
			if g.As("Dung").View().Me.Turn == nil || g.As("Hoa").View().Me.Turn != nil {
				t.Fatal("only the Guard has a turn card")
			}
			if *g.As("Dung").View().NightCall != *g.As("Hoa").View().NightCall {
				t.Fatal("the public night call must be the same on every phone")
			}
		}},
		{"what is read aloud, never the order of deaths", func(t *testing.T) {
			g := newTable(t)
			g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
			g.Dawn()
			g.As("Hoa").
				Event(engine.DawnAnnounced{Kind: engine.DawnDeaths, Dead: []engine.PlayerID{"Phuc"}}).
				NoEvent(engine.PlayerDied{Player: "Phuc"}).
				Dead("Phuc")
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestView_EveryPlayer_SeesOwnRoleAndThePublicTable(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"own role", func(t *testing.T) {
			g := newTable(t)
			for name, role := range cast {
				g.As(name).OwnRole(role).NoRoleExceptOwn()
			}
		}},
		{"role list, seating, who is alive, the Chief", func(t *testing.T) {
			g := chiefTable(t, "Dung")
			g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
			g.Dawn()
			for _, name := range []string{"An", "Dung", "Em", "Hoa"} {
				g.As(name).PublicTable().Dead("Phuc")
			}
		}},
		{"Chief election votes", func(t *testing.T) {
			g := newTable(t)
			g.Election().Vote("An", "Dung").Vote("Hoa", "Hoa")
			g.As("Khoa").ChiefVote("An", "Dung").ChiefVote("Hoa", "Hoa")
		}},
		{"nominations", func(t *testing.T) {
			g, d := day(t)
			d.Nominate("Giang", "An").NominateNobody("Hoa")
			g.As("Khoa").Nomination("Giang", "An").Nomination("Hoa", "")
		}},
		{"kill or spare votes", func(t *testing.T) {
			g, d := day(t)
			toVerdict(d, "Giang")
			d.Kill("An").Spare("Em")
			g.As("Khoa").Verdict("An", true).Verdict("Em", false)
		}},
		{"past days' votes stay in the history", func(t *testing.T) {
			g, d := day(t)
			toVerdict(d, "Giang")
			d.Kill("An")
			d.TimeOut()
			g.As("Khoa").
				Event(engine.NominationCast{Voter: "Hoa", Target: "Giang"}).
				Event(engine.VerdictCast{Voter: "An", Kill: true})
		}},
		{"own turn at night, nobody else's", func(t *testing.T) {
			g := newTable(t)
			g.Night().TurnOf(Seer)
			g.As("Em").Turn(Seer)
			g.As("Hoa").NoTurn()
			g.As("An").NoTurn()
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

// A full night and day; at every moment, no living player learns a role, a cause of death, or who was bitten, protected or
// checked, unless it is their own.
func TestView_NobodyAlive_SeesRolesCausesOrNightTargets(t *testing.T) {
	g := chiefTable(t, "Dung")
	check := func(moment string) {
		t.Run(moment, func(t *testing.T) {
			for _, p := range g.State().Players {
				if p.Alive {
					g.As(string(p.ID)).NoSecretsOfOthers()
				}
			}
		})
	}
	n := g.Night().Guard("Dung").Protects("Em").Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	check("pack step")
	n.Seer("Em").Checks("An")
	check("seer step")
	g.Dawn().ExpectDead("Phuc")
	check("dawn")
	g.Day().Hanged("Giang")
	check("dusk")
	g.Night().TurnOf(Seer)
	check("second night")

	g.As("Hoa").NoEvent(engine.PlayerDied{Player: "Phuc"}).NoEvent(engine.PlayerDied{Player: "Giang"}).SeerResults(0)
	g.As("Em").SeerResults(1)
}

func TestView_GameOver_EverybodySeesEverything(t *testing.T) {
	g := NewGame(t,
		Seats("An", "Binh", "Chi", "Dung", "Em"),
		Cast("An", Werewolf, "Binh", Guard, "Chi", Seer, "Dung", Villager, "Em", Villager))
	g.Night()
	g.Dawn()
	g.Day().Hanged("An")
	g.ExpectPhase(engine.PhaseEnded)
	for _, name := range []string{"Binh", "Chi", "Dung", "Em"} {
		g.As(name).Full()
	}
}
