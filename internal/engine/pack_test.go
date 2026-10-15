package engine

import "testing"

func packTable(alive map[PlayerID]bool, votes map[PlayerID]PlayerID) State {
	s := NewState()
	s.Seats = []PlayerID{"A", "B", "C", "D"}
	for _, id := range s.Seats {
		team := TeamWolves
		if id == "D" {
			team = TeamVillage
		}
		s.Players[id] = &Player{ID: id, Team: team, Alive: alive[id] || len(alive) == 0}
	}
	s.Pending.WolfVotes = votes
	return s
}

func TestWolfVoters_AreTheLivingWolves(t *testing.T) {
	s := packTable(map[PlayerID]bool{"A": true, "B": false, "C": true, "D": true}, nil)
	got := WolfVoters(s)
	if len(got) != 2 || got[0] != "A" || got[1] != "C" {
		t.Fatalf("voters = %v, want [A C]", got)
	}
}

func TestBiteTarget_CountsOnlyTheVoters(t *testing.T) {
	votes := map[PlayerID]PlayerID{"A": "D", "B": "D", "C": "A"}
	tests := []struct {
		name   string
		voters []PlayerID
		want   PlayerID
		bitten bool
	}{
		{"all three wolves, two agree", []PlayerID{"A", "B", "C"}, "D", true},
		// A blocked wolf (T14) is just left out of the voters.
		{"B blocked: A for D, C for A", []PlayerID{"A", "C"}, "", false},
		{"B and C blocked: A alone decides", []PlayerID{"A"}, "D", true},
		{"every wolf blocked", nil, "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := BiteTarget(packTable(nil, votes), tc.voters)
			if got != tc.want || ok != tc.bitten {
				t.Fatalf("BiteTarget = (%q, %v), want (%q, %v)", got, ok, tc.want, tc.bitten)
			}
		})
	}
}
