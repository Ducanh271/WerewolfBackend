package engine

import "testing"

// handBuilt makes a state from seats, for scenarios no role can reach yet.
type seat struct {
	id    PlayerID
	team  Team
	alive bool
}

func alive(id PlayerID, team Team) seat { return seat{id, team, true} }
func dead(id PlayerID, team Team) seat  { return seat{id, team, false} }

func handBuilt(lovers *LoversLink, seats ...seat) State {
	s := State{Players: map[PlayerID]*Player{}, Lovers: lovers}
	for _, st := range seats {
		s.Seats = append(s.Seats, st.id)
		s.Players[st.id] = &Player{ID: st.id, Team: st.team, Alive: st.alive}
	}
	return s
}

const (
	W = TeamWolves
	V = TeamVillage
)

func TestWin_Conditions(t *testing.T) {
	couple := &LoversLink{A: "l1", B: "l2"}
	tests := []struct {
		name   string
		state  State
		over   bool
		winner Team
		draw   bool
	}{
		{"game goes on while wolves are fewer than villagers",
			handBuilt(nil, alive("w", W), alive("a", V), alive("b", V)), false, 0, false},
		{"wolves win when equal to villagers",
			handBuilt(nil, alive("w", W), alive("a", V)), true, TeamWolves, false},
		{"wolves win when more than villagers",
			handBuilt(nil, alive("w1", W), alive("w2", W), alive("a", V)), true, TeamWolves, false},
		{"village wins when no wolf is left",
			handBuilt(nil, dead("w", W), alive("a", V), alive("b", V)), true, TeamVillage, false},
		{"everybody dead is a draw",
			handBuilt(nil, dead("w", W), dead("a", V)), true, TeamNone, true},
		{"cursed villager counts as a wolf",
			handBuilt(nil, alive("w", W), alive("cursed", W), alive("a", V), alive("b", V), alive("c", V)), false, 0, false},
		{"cursed villager tips the count to the wolves",
			handBuilt(nil, alive("w", W), alive("cursed", W), alive("a", V), alive("b", V)), true, TeamWolves, false},
		{"last wolf cursed away from the village keeps the game going",
			handBuilt(nil, alive("cursed", W), alive("a", V), alive("b", V)), false, 0, false},
		{"cross-team couple alone: lovers win, not wolves",
			handBuilt(couple, alive("l1", W), alive("l2", V), dead("a", V)), true, TeamLovers, false},
		{"cross-team couple with a villager left: no winner yet",
			handBuilt(couple, alive("l1", W), alive("l2", V), alive("a", V)), false, 0, false},
		{"couple is excluded from both counts: the other wolf outnumbers the lone villager",
			handBuilt(couple, alive("l1", W), alive("l2", V), alive("w", W), alive("a", V)), true, TeamWolves, false},
		{"couple is excluded: the wolf lover does not help the wolves",
			handBuilt(couple, alive("l1", W), alive("l2", V), alive("a", V), alive("b", V)), false, 0, false},
		{"wolves need a wolf outside the couple even when no villager is left",
			handBuilt(couple, alive("l1", W), alive("l2", V), alive("x", V), dead("a", V)), false, 0, false},
		{"a wolf in love still stops the village from winning",
			handBuilt(couple, alive("l1", W), alive("l2", V), alive("a", V), alive("b", V), alive("c", V)), false, 0, false},
		{"same-team couple of wolves counts as wolves",
			handBuilt(couple, alive("l1", W), alive("l2", W), alive("a", V)), true, TeamWolves, false},
		{"same-team couple of villagers counts as villagers",
			handBuilt(couple, dead("w", W), alive("l1", V), alive("l2", V)), true, TeamVillage, false},
		{"couple cursed into the same team is no longer the third team",
			handBuilt(couple, alive("l1", W), alive("l2", W)), true, TeamWolves, false},
		{"couple broken by a death is not the third team",
			handBuilt(couple, dead("l1", W), alive("l2", V), alive("a", V)), true, TeamVillage, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, over := CheckWin(tc.state)
			if over != tc.over {
				t.Fatalf("game over = %v, want %v (outcome %+v)", over, tc.over, got)
			}
			if over && (got.Winner != tc.winner || got.Draw != tc.draw) {
				t.Fatalf("outcome %+v, want winner %v draw %v", got, tc.winner, tc.draw)
			}
		})
	}
}

func TestWin_AfterDeaths_EndedGameSkipsChiefSuccession(t *testing.T) {
	s := handBuilt(nil, alive("w", W), alive("chief", V), alive("a", V))
	s.Chief = "chief"
	s.Settings = DefaultSettings()
	died := []Event{PlayerDied{Player: "chief"}}
	evs := afterDeaths(s, died, []PlayerID{"chief"})
	if len(evs) != 1 {
		t.Fatalf("events %v, want only GameEnded", evs)
	}
	end, ok := evs[0].(GameEnded)
	if !ok || end.Outcome.Winner != TeamWolves {
		t.Fatalf("event %#v, want wolves winning", evs[0])
	}
	after := Fold(s, append(died, evs...)...)
	if after.Phase != PhaseEnded || after.Winner == nil || *after.Winner != TeamWolves {
		t.Fatalf("phase %v winner %v, want Ended with wolves", after.Phase, after.Winner)
	}
}

func TestWin_AfterDeaths_EverybodyDeadAtOnce_IsADraw(t *testing.T) {
	s := handBuilt(nil, alive("w", W), alive("a", V))
	s.Chief = "a"
	s.Settings = DefaultSettings()
	died := []Event{PlayerDied{Player: "w"}, PlayerDied{Player: "a"}}
	evs := afterDeaths(s, died, []PlayerID{"w", "a"})
	after := Fold(s, append(died, evs...)...)
	if after.Phase != PhaseEnded || after.Winner != nil {
		t.Fatalf("phase %v winner %v, want Ended with no winner", after.Phase, after.Winner)
	}
	if !evs[0].(GameEnded).Outcome.Draw {
		t.Fatalf("outcome %+v, want draw", evs[0])
	}
}

func TestWin_AfterDeaths_GameGoesOn_ChiefStillNamesSuccessor(t *testing.T) {
	s := handBuilt(nil, alive("w", W), alive("chief", V), alive("a", V), alive("b", V))
	s.Chief = "chief"
	s.Settings = DefaultSettings()
	died := []Event{PlayerDied{Player: "chief"}}
	evs := afterDeaths(s, died, []PlayerID{"chief"})
	if len(evs) != 1 {
		t.Fatalf("events %v, want the succession start", evs)
	}
	if _, ok := evs[0].(ChiefSuccessionStarted); !ok {
		t.Fatalf("event %#v, want ChiefSuccessionStarted", evs[0])
	}
}

func TestWin_AfterDeaths_NoDeaths_NoWinCheck(t *testing.T) {
	s := handBuilt(nil, alive("w", W), alive("a", V)) // already "won" on paper
	if evs := afterDeaths(s, nil, nil); len(evs) != 0 {
		t.Fatalf("events %v, want none", evs)
	}
}
