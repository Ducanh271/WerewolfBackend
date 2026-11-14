package store_test

import (
	"testing"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// The third team wins (rules §7, §8): the two lovers keep their own team,
// Village and Wolves, and only they are winners.
func TestGameRecord_LoversWin_KeepTheirOwnTeamsAndOnlyTheyWin(t *testing.T) {
	s := engine.NewState()
	s.Seats = []engine.PlayerID{"p1", "p2", "p3"}
	s.Players = map[engine.PlayerID]*engine.Player{
		"p1": {ID: "p1", Nickname: "An", Role: "Werewolf", Team: engine.TeamWolves, Alive: true},
		"p2": {ID: "p2", Nickname: "Bình", Role: "Seer", Team: engine.TeamVillage, Alive: true},
		"p3": {ID: "p3", Nickname: "Chi", Role: "Villager", Team: engine.TeamVillage},
	}
	s.Night = 4
	end := engine.GameEnded{Outcome: engine.Outcome{Winner: engine.TeamLovers, Winners: []engine.PlayerID{"p1", "p2"}}}

	rec := store.NewGameRecord(s, end)
	if rec.Winner != engine.TeamLovers || rec.Draw || rec.Aborted || rec.Nights != 4 {
		t.Fatalf("record = %+v, want the lovers' win after 4 nights", rec)
	}
	want := []struct {
		team engine.Team
		won  bool
	}{{engine.TeamWolves, true}, {engine.TeamVillage, true}, {engine.TeamVillage, false}}
	for i, p := range rec.Players {
		if p.Seat != i || p.Team != want[i].team || p.Won != want[i].won {
			t.Errorf("seat %d = %+v, want team %v won %v", i, p, want[i].team, want[i].won)
		}
	}
}

func TestGameRecord_AbortedGame_IsADrawWithoutWinner(t *testing.T) {
	s := engine.NewState()
	rec := store.NewGameRecord(s, engine.GameEnded{Outcome: engine.Outcome{Draw: true, Aborted: true, Winner: engine.TeamWolves}})
	if !rec.Draw || !rec.Aborted || rec.Winner != engine.TeamNone {
		t.Fatalf("record = %+v, want an aborted draw with no winner", rec)
	}
}
