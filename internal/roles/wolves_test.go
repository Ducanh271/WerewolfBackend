package roles_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	"masoi/internal/roles"
)

func TestNight_TwoOfThreeWolvesBite_TargetDiesAndNamesAreAnnounced(t *testing.T) {
	g := newTable(t)
	g.Night().
		Wolf("An").Bites("Phuc").
		Wolf("Binh").Bites("Phuc").
		Wolf("Chi").Bites("Giang")
	g.Dawn().ExpectDead("Phuc").ExpectAnnouncement(i18n.DawnSomeoneDied)
}

func TestNight_WolvesNeedMoreThanHalf_TableOfOutcomes(t *testing.T) {
	tests := []struct {
		name  string
		votes [][2]string // wolf, target
		dead  []string
	}{
		{"all three agree", [][2]string{{"An", "Phuc"}, {"Binh", "Phuc"}, {"Chi", "Phuc"}}, []string{"Phuc"}},
		{"three-way split", [][2]string{{"An", "Phuc"}, {"Binh", "Giang"}, {"Chi", "Hoa"}}, nil},
		{"one vote of three", [][2]string{{"An", "Phuc"}}, nil},
		{"two vote, one abstains", [][2]string{{"An", "Phuc"}, {"Binh", "Phuc"}}, []string{"Phuc"}},
		{"nobody votes", nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			n := g.Night()
			for _, v := range tc.votes {
				n.Wolf(v[0]).Bites(v[1])
			}
			d := g.Dawn()
			if tc.dead == nil {
				d.ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
				return
			}
			d.ExpectDead(tc.dead...)
		})
	}
}

func TestNight_WolfChangesVote_LastVoteCounts(t *testing.T) {
	g := newTable(t)
	g.Night().
		Wolf("An").Bites("Phuc").
		Wolf("Binh").Bites("Phuc").
		Wolf("An").Bites("Giang").
		Wolf("Chi").Bites("Giang")
	g.Dawn().ExpectDead("Giang")
}

func TestNight_WolvesCannotBiteAWolf(t *testing.T) {
	g := newTable(t)
	g.Night().
		Wolf("An").CannotBite("Binh", engine.ErrTargetIsWolf).
		Wolf("An").CannotBite("An", engine.ErrTargetIsWolf)
}

// A villager cursed into the pack is a wolf the pack cannot bite (rules §5). The
// curse event arrives with T15, so the scenario flips the team directly.
func TestNight_WolvesCannotBiteTheNewlyCursed(t *testing.T) {
	g := newTable(t)
	g.State().Players[g.Player("Phuc")].Team = engine.TeamWolves
	g.Night().Wolf("An").CannotBite("Phuc", engine.ErrTargetIsWolf)
}

func TestNight_WolfVoteIsRejectedWhenNotAllowed(t *testing.T) {
	g := newTable(t)
	n := g.Night()
	// Night 1 starts with the Guard's turn, not the pack's.
	n.DoExpectErr(roles.WolfVote{By: g.Player("An"), Target: g.Player("Phuc")}, engine.ErrNotYourTurn)
}

func TestNight_AVillagerCannotVoteWithThePack(t *testing.T) {
	g := newTable(t)
	g.Night().Wolf("Phuc").CannotBite("Giang", engine.ErrNotYourTurn)
}

func TestNight_WolvesActingAfterTheirTurn_FailsTheScenario(t *testing.T) {
	g := newTable(t)
	n := g.Night()
	n.Seer("Em").Checks("An")
	n.DoExpectErr(roles.WolfVote{By: g.Player("An"), Target: g.Player("Phuc")}, engine.ErrNotYourTurn)
}
