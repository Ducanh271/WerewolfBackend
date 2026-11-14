package mysql_test

import (
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
	"masoi/internal/store/mysql"
)

var ended = time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)

// saveGame stores a finished game of three seats in room logKey; seat i has
// the token "<logKey>/<i>".
func saveGame(t *testing.T, s *mysql.Store, logKey string, no int, end time.Time) {
	t.Helper()
	if err := s.Append(logKey, engine.NewGameStarted{}); err != nil { // creates the room
		t.Fatal(err)
	}
	rec := store.GameRecord{LogKey: logKey, Code: logKey[:4], GameNo: no, Mode: engine.Online,
		Winner: engine.TeamWolves, Nights: 3, Started: end.Add(-time.Hour), Ended: end}
	for i, role := range []engine.RoleID{"Werewolf", "Seer", "Villager"} {
		team := engine.TeamVillage
		if i == 0 {
			team = engine.TeamWolves
		}
		rec.Players = append(rec.Players, store.GamePlayer{
			ID: engine.PlayerID([]string{"p1", "p2", "p3"}[i]), Seat: i, Nickname: []string{"Ân", "Bình", "Chi"}[i],
			Role: role, Team: team, Alive: i != 1, Won: i == 0,
			TokenHash: store.TokenHash(logKey + "/" + string(rune('0'+i))),
		})
	}
	if err := s.SaveGame(rec); err != nil {
		t.Fatal(err)
	}
}

func hashes(tokens ...string) []string {
	out := make([]string, len(tokens))
	for i, tok := range tokens {
		out[i] = store.TokenHash(tok)
	}
	return out
}

func TestMySQL_History_ReadsOnlyTheGamesOfTheTokens_NewestFirst(t *testing.T) {
	s := openTest(t)
	saveGame(t, s, "AAAA-000000000001", 1, ended.Add(-2*time.Hour))
	saveGame(t, s, "AAAA-000000000001", 2, ended.Add(-time.Hour))
	saveGame(t, s, "BBBB-000000000002", 1, ended)
	since := ended.Add(-store.Retention)

	got, err := s.GamesFor(hashes("AAAA-000000000001/1"), since, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].GameNo != 2 || got[1].GameNo != 1 || got[0].Code != "AAAA" {
		t.Fatalf("games = %+v, want room AAAA's #2 then #1", got)
	}
	g := got[0]
	if g.Mode != engine.Online || g.Winner != engine.TeamWolves || g.Nights != 3 || !g.Ended.Equal(ended.Add(-time.Hour)) || len(g.Players) != 3 {
		t.Fatalf("game = %+v", g.GameRecord)
	}
	if p := g.Players[0]; p.Nickname != "Ân" || p.Role != "Werewolf" || p.Team != engine.TeamWolves || !p.Won || !p.Alive || p.Seat != 0 {
		t.Fatalf("seat 0 = %+v", p)
	}
	if got, _ := s.GamesFor(hashes("nobody"), since, 10); len(got) != 0 {
		t.Fatalf("a stranger's token read %d games", len(got))
	}
	if got, _ := s.GamesFor(hashes("AAAA-000000000001/0", "BBBB-000000000002/2"), since, 2); len(got) != 2 || got[0].Code != "BBBB" {
		t.Fatalf("two rooms, limit 2: %+v, want BBBB first and two games", got)
	}
	if got, _ := s.GamesFor(hashes("BBBB-000000000002/0"), ended.Add(time.Minute), 10); len(got) != 0 {
		t.Fatalf("a game ended before since was read")
	}
}

func TestMySQL_ReplaceToken_MovesTheRoomsGamesToTheNewToken(t *testing.T) {
	s := openTest(t)
	saveGame(t, s, "CCCC-000000000003", 1, ended)
	saveGame(t, s, "CCCC-000000000003", 2, ended.Add(time.Minute))
	if err := s.ReplaceToken("CCCC-000000000003", "p2", store.TokenHash("new phone")); err != nil {
		t.Fatal(err)
	}
	since := ended.Add(-time.Hour)
	if got, _ := s.GamesFor(hashes("CCCC-000000000003/1"), since, 10); len(got) != 0 {
		t.Errorf("the old token still reads %d games", len(got))
	}
	if got, _ := s.GamesFor(hashes("new phone"), since, 10); len(got) != 2 {
		t.Errorf("the new token reads %d games, want 2", len(got))
	}
	if got, _ := s.GamesFor(hashes("CCCC-000000000003/0"), since, 10); len(got) != 2 {
		t.Errorf("another seat reads %d games, want 2 (unchanged)", len(got))
	}
}

func TestMySQL_Purge_DropsOldRoomsWithTheirEventsAndGames(t *testing.T) {
	s := openTest(t)
	saveGame(t, s, "OLD1-000000000004", 1, ended)
	saveGame(t, s, "NEW1-000000000005", 1, ended)
	if _, err := s.DB().Exec(`UPDATE rooms SET created_at = ? WHERE code = 'OLD1'`, time.Now().UTC().Add(-store.Retention-time.Hour)); err != nil {
		t.Fatal(err)
	}
	n, err := s.Purge(time.Now().UTC().Add(-store.Retention))
	if err != nil || n != 1 {
		t.Fatalf("purged %d rooms, %v; want 1", n, err)
	}
	for table, want := range map[string]int{"rooms": 1, "room_events": 1, "games": 1, "game_players": 3} {
		if got := count(t, s, "SELECT COUNT(*) FROM "+table); got != want {
			t.Errorf("%s: %d rows left, want %d", table, got, want)
		}
	}
}
