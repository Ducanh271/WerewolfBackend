package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

func TestNight_SeerChecks_WolfOrNotWolf(t *testing.T) {
	tests := []struct {
		target string
		wolf   bool
	}{
		{"An", true},    // plain wolf
		{"Phuc", false}, // villager
		{"Dung", false}, // guard
		{"Em", false},   // herself: allowed
	}
	for _, tc := range tests {
		t.Run(tc.target, func(t *testing.T) {
			g := newTable(t)
			g.Night().Seer("Em").Checks(tc.target).ExpectSeerSaw("Em", tc.target, tc.wolf)
		})
	}
}

// A cursed villager reads as a wolf (rules §5). The curse event arrives with
// T15, so the scenario flips the team directly.
func TestNight_SeerChecksACursedPlayer_SeesAWolf(t *testing.T) {
	g := newTable(t)
	g.State().Players[g.Player("Phuc")].Team = engine.TeamWolves
	g.Night().Seer("Em").Checks("Phuc").ExpectSeerSaw("Em", "Phuc", true)
}

func TestNight_SeerChoiceIsRejectedWhenInvalid(t *testing.T) {
	tests := []struct {
		name string
		do   func(n *Night)
	}{
		{"not the seer", func(n *Night) { n.Seer("Phuc").CannotCheck("An", engine.ErrNotYourTurn) }},
		{"second check", func(n *Night) { n.Seer("Em").Checks("An").Seer("Em").CannotCheck("Binh", engine.ErrAlreadyActed) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { tc.do(newTable(t).Night()) })
	}
}

func TestNight_SeerCheckDoesNotChangeTheNight(t *testing.T) {
	g := newTable(t)
	g.Night().Seer("Em").Checks("An")
	g.Dawn().ExpectNobodyDied()
}
