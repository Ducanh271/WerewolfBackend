package roles_test

import (
	"testing"

	"masoi/internal/engine"
)

func TestVerdict_KillMustBeatSpare_TieSpares(t *testing.T) {
	tests := []struct {
		name        string
		kill, spare []string
		hanged      bool
	}{
		{"more kill", []string{"An", "Binh", "Chi"}, []string{"Dung", "Em"}, true},
		{"tie", []string{"An", "Binh"}, []string{"Dung", "Em"}, false},
		{"more spare", []string{"An"}, []string{"Dung", "Em"}, false},
		{"nobody voted", nil, nil, false},
		{"only kill votes", []string{"An"}, nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, d := day(t)
			toVerdict(d, "Giang")
			for _, v := range tc.kill {
				d.Kill(v)
			}
			for _, v := range tc.spare {
				d.Spare(v)
			}
			d.TimeOut()
			if tc.hanged {
				g.ExpectDead("Giang").ExpectPhase(engine.PhaseDusk)
			} else {
				g.ExpectAlive("Giang").ExpectPhase(engine.PhaseDay)
			}
		})
	}
}

func TestVerdict_ChiefsVoteCountsDouble_ForKillAndForSpare(t *testing.T) {
	tests := []struct {
		name        string
		kill, spare []string
		hanged      bool
	}{
		{"chief kills, tied heads", []string{"Phuc"}, []string{"An", "Binh"}, false}, // 2 v 2
		{"chief kills and wins", []string{"Phuc", "Chi"}, []string{"An", "Binh"}, true},
		{"chief kills alone", []string{"Phuc"}, []string{"An"}, true}, // 2 v 1
		{"chief spares and wins", []string{"An", "Binh"}, []string{"Phuc"}, false},
		{"chief spares, one more kill", []string{"An", "Binh", "Chi"}, []string{"Phuc"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, d := chiefDay(t)
			toVerdict(d, "Giang")
			for _, v := range tc.kill {
				d.Kill(v)
			}
			for _, v := range tc.spare {
				d.Spare(v)
			}
			d.TimeOut()
			if tc.hanged {
				g.ExpectDead("Giang")
			} else {
				g.ExpectAlive("Giang")
			}
		})
	}
}

func TestVerdict_PlayerOnTrialCannotVote(t *testing.T) {
	_, d := day(t)
	toVerdict(d, "Giang")
	d.VerdictExpectErr("Giang", false, engine.ErrOnTrial)
	d.VerdictExpectErr("Giang", true, engine.ErrOnTrial)
}

func TestVerdict_ChiefOnTrialCannotVote(t *testing.T) {
	_, d := chiefDay(t)
	toVerdict(d, "Phuc")
	d.VerdictExpectErr("Phuc", true, engine.ErrOnTrial)
}

func TestVerdict_VoteMustBeInTheVerdictPhase(t *testing.T) {
	_, d := day(t)
	d.VerdictExpectErr("An", true, engine.ErrWrongPhase)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock).TimeOut().ExpectPhase(engine.PhaseTrial).VerdictExpectErr("An", true, engine.ErrWrongPhase) // still defending
}

func TestVerdict_CanChangeVoteUntilItCloses(t *testing.T) {
	g, d := day(t)
	toVerdict(d, "Giang")
	d.Kill("An").Spare("An").Spare("Binh").Kill("Chi").TimeOut() // 2 spare, 1 kill
	g.ExpectAlive("Giang")
}

func TestVerdict_ClosesWhenEveryoneElseVoted(t *testing.T) {
	g, d := day(t)
	toVerdict(d, "Giang")
	for _, name := range all {
		if name != "Giang" {
			d.Kill(name)
		}
	}
	g.ExpectDead("Giang").ExpectPhase(engine.PhaseDusk) // no timeout needed
}
