package engine

import (
	"errors"
	"fmt"
	"testing"
)

// TestDay_AbortVote_PassesAtTwoThirdsOfLiving checks the exact boundary of
// rules §8b: yes*3 >= alive*2, computed with integers.
func TestDay_AbortVote_PassesAtTwoThirdsOfLiving(t *testing.T) {
	tests := []struct {
		alive, yes int
		passes     bool
	}{
		{3, 1, false},
		{3, 2, true},
		{4, 2, false},
		{4, 3, true},
		{5, 3, false},
		{5, 4, true},
		{6, 3, false},
		{6, 4, true},
		{9, 5, false},
		{9, 6, true},
	}
	for _, tc := range tests {
		s := NewState()
		s.Phase = PhaseDay
		var voters []PlayerID
		for i := 0; i < tc.alive; i++ {
			id := PlayerID(fmt.Sprintf("p%d", i))
			s.Players[id] = &Player{ID: id, Alive: true}
			voters = append(voters, id)
		}
		for _, id := range voters[:tc.yes-1] {
			s = Fold(s, AbortVoteCast{Voter: id, Yes: true})
		}
		evs, err := Decide(s, CastAbortVote{Voter: voters[tc.yes-1], Yes: true})
		if err != nil {
			t.Fatalf("%d yes of %d living: %v", tc.yes, tc.alive, err)
		}
		_, ended := evs[len(evs)-1].(GameEnded)
		if ended != tc.passes {
			t.Errorf("%d yes of %d living: aborted = %v, want %v", tc.yes, tc.alive, ended, tc.passes)
		}
	}
}

func TestDay_AbortVote_EndsTheGameWithNoWinner(t *testing.T) {
	s := NewState()
	s.Phase = PhaseDay
	for _, id := range []PlayerID{"a", "b", "c"} {
		s.Players[id] = &Player{ID: id, Alive: true}
	}
	s = Fold(s, AbortVoteCast{Voter: "a", Yes: true})
	s = apply(t, s, CastAbortVote{Voter: "b", Yes: true})

	if s.Phase != PhaseEnded {
		t.Fatalf("phase = %s, want Ended", s.Phase)
	}
	if s.Winner != nil {
		t.Fatalf("winner = %v, want none", *s.Winner)
	}
	if len(s.Winners) != 0 {
		t.Fatalf("winners = %v, want none", s.Winners)
	}
}

func TestDay_AbortVote_OnlyLivingPlayersVote(t *testing.T) {
	s := NewState()
	s.Phase = PhaseDay
	s.Players["a"] = &Player{ID: "a", Alive: false}

	_, err := Decide(s, CastAbortVote{Voter: "a", Yes: true})
	if !errors.Is(err, ErrNotAlive) {
		t.Fatalf("err = %v, want ErrNotAlive", err)
	}
}

func TestDay_AbortVote_OnlyAllowedByDay(t *testing.T) {
	allowed := map[Phase]bool{
		PhaseDay: true, PhaseNominationLock: true, PhaseTrial: true,
		PhaseVerdict: true, PhaseDusk: true,
	}
	disallowed := []Phase{PhaseLobby, PhaseRoleView, PhaseChiefElection, PhaseChiefSuccession, PhaseNight, PhaseDawn, PhaseHunterShot, PhaseEnded}

	for phase := range allowed {
		s := NewState()
		s.Phase = phase
		s.Players["a"] = &Player{ID: "a", Alive: true}
		if _, err := Decide(s, CastAbortVote{Voter: "a", Yes: true}); err != nil {
			t.Errorf("phase %s: err = %v, want none", phase, err)
		}
	}
	for _, phase := range disallowed {
		s := NewState()
		s.Phase = phase
		s.Players["a"] = &Player{ID: "a", Alive: true}
		if _, err := Decide(s, CastAbortVote{Voter: "a", Yes: true}); !errors.Is(err, ErrWrongPhase) {
			t.Errorf("phase %s: err = %v, want ErrWrongPhase", phase, err)
		}
	}
}

func TestDay_AbortVote_WithdrawnVoteDoesNotCount(t *testing.T) {
	s := NewState()
	s.Phase = PhaseDay
	for _, id := range []PlayerID{"a", "b", "c"} {
		s.Players[id] = &Player{ID: id, Alive: true}
	}
	s = apply(t, s, CastAbortVote{Voter: "a", Yes: true})
	s = apply(t, s, CastAbortVote{Voter: "a", Yes: false}) // withdraws it
	evs := apply(t, s, CastAbortVote{Voter: "b", Yes: true})
	if evs.Phase == PhaseEnded {
		t.Fatalf("2 of 3 should not be enough once the first vote withdrew: phase = %s", evs.Phase)
	}
}

// TestDay_AbortVote_ClearsAtNightStart: a pending proposal does not carry
// into the next day; the table must propose again.
func TestDay_AbortVote_ClearsAtNightStart(t *testing.T) {
	s := NewState()
	s.Phase = PhaseDusk
	for _, id := range []PlayerID{"a", "b", "c"} {
		s.Players[id] = &Player{ID: id, Alive: true}
	}
	s = apply(t, s, CastAbortVote{Voter: "a", Yes: true}) // 1 of 3, not enough to end it
	if len(s.Day.AbortVotes) == 0 {
		t.Fatalf("expected a pending abort vote before the night")
	}
	s = Fold(s, NightStarted{Night: s.Night + 1})
	if len(s.Day.AbortVotes) != 0 {
		t.Fatalf("abort votes = %v, want cleared once the night starts", s.Day.AbortVotes)
	}
}

// TestDay_AbortVote_DeathDuringTheDayRechecksTheThreshold: a hanging that
// thins the table can bring a vote over 2/3 on its own, without anyone
// casting a new one (rules §8b).
func TestDay_AbortVote_DeathDuringTheDayRechecksTheThreshold(t *testing.T) {
	s := NewState()
	s.Phase = PhaseVerdict
	s.Day.Defendant = "doomed"
	s.Day.Verdicts = map[PlayerID]bool{}
	// One wolf (c) among three villagers so the hanging below does not also
	// win the game for the village: wolves 1, villagers 2 after it.
	s.Players["a"] = &Player{ID: "a", Alive: true, Team: TeamVillage}
	s.Players["b"] = &Player{ID: "b", Alive: true, Team: TeamVillage}
	s.Players["c"] = &Player{ID: "c", Alive: true, Team: TeamWolves}
	s.Players["doomed"] = &Player{ID: "doomed", Alive: true, Team: TeamVillage}
	// 4 alive needs 3 yes; only 2 voted so far (a, b). Once "doomed" is
	// hanged, only 3 remain alive, and 2 yes is already 2/3 of 3.
	s = Fold(s, AbortVoteCast{Voter: "a", Yes: true})
	s = Fold(s, AbortVoteCast{Voter: "b", Yes: true})
	s = Fold(s, VerdictCast{Voter: "a", Kill: true})
	s = Fold(s, VerdictCast{Voter: "b", Kill: true})

	evs, err := Decide(s, CastVerdict{Voter: "c", Kill: true})
	if err != nil {
		t.Fatalf("Decide(CastVerdict) = %v, want no error", err)
	}
	after := Fold(s, evs...)
	if after.Phase != PhaseEnded {
		t.Fatalf("phase = %s, want Ended (the hanging should have pushed the abort vote over 2/3)", after.Phase)
	}
	if after.Winner != nil {
		t.Fatalf("winner = %v, want none: this is an abort, not a village/wolf win", *after.Winner)
	}
}

func TestDay_AbortVote_IsPublicInTheView(t *testing.T) {
	s := NewState()
	s.Phase = PhaseDay
	s.Seats = []PlayerID{"a", "b"}
	s.Players["a"] = &Player{ID: "a", Nickname: "A", Alive: true}
	s.Players["b"] = &Player{ID: "b", Nickname: "B", Alive: true}
	s = apply(t, s, CastAbortVote{Voter: "a", Yes: true})

	v := ViewFor(s, PlayerViewer("b"))
	if v.Votes.AbortVotes["a"] != true {
		t.Fatalf("votes.abortVotes = %v, want {a: true} visible to another living player", v.Votes.AbortVotes)
	}
}
