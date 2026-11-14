package mysql_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/roles"
	"masoi/internal/store"
	"masoi/internal/store/mysql"
	"masoi/internal/store/storetest"
)

// openTest opens an emptied database from MASOI_MYSQL_DSN (make
// mysql-test-up, make test-mysql). The tests skip without it and under -short.
func openTest(t *testing.T) *mysql.Store {
	t.Helper()
	dsn := os.Getenv("MASOI_MYSQL_DSN")
	if dsn == "" || testing.Short() {
		t.Skip("needs MySQL: set MASOI_MYSQL_DSN (make mysql-test-up && make test-mysql)")
	}
	s := openRetry(t, dsn)
	for _, table := range []string{"game_players", "games", "room_events", "rooms", "schema_version"} {
		if _, err := s.DB().Exec("DROP TABLE IF EXISTS " + table); err != nil {
			t.Fatal(err)
		}
	}
	_ = s.Close()
	s = openRetry(t, dsn) // migrates the empty database
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// openRetry waits for a MySQL that is still starting (a fresh container).
func openRetry(t *testing.T, dsn string) *mysql.Store {
	t.Helper()
	var err error
	for i := 0; i < 30; i++ {
		var s *mysql.Store
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s, err = mysql.Open(ctx, dsn)
		cancel()
		if err == nil {
			return s
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("open MySQL: %v", err)
	return nil
}

func reopen(t *testing.T) *mysql.Store {
	t.Helper()
	s := openRetry(t, os.Getenv("MASOI_MYSQL_DSN"))
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func count(t *testing.T, s *mysql.Store, query string, args ...any) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func TestMySQL_Open_MigratesOnceAndAgainIsANoOp(t *testing.T) {
	s := openTest(t)
	reopen(t)
	if n := count(t, s, "SELECT COUNT(*) FROM schema_version"); n != 2 {
		t.Fatalf("schema_version rows = %d, want 2 (001, 002, each applied once)", n)
	}
}

func TestMySQL_EveryEventWithEveryField_ReadsBackTheSame(t *testing.T) {
	s := openTest(t)
	want := storetest.FilledEvents()
	if err := s.Append("K7QP-000000000001", want...); err != nil {
		t.Fatal(err)
	}
	got, err := reopen(t).Load("K7QP-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("read back %d events, want %d; first difference:\n%s", len(got), len(want), firstDiff(got, want))
	}
}

func firstDiff(got, want []engine.Event) string {
	for i := range want {
		if i >= len(got) || !reflect.DeepEqual(got[i], want[i]) {
			if i >= len(got) {
				return fmt.Sprintf("#%d missing: %#v", i, want[i])
			}
			return fmt.Sprintf("#%d\n got  %#v\n want %#v", i, got[i], want[i])
		}
	}
	return "extra events"
}

// A process that comes back (or a second Store) carries on the same log.
func TestMySQL_Append_ContinuesTheLogAfterAReopen(t *testing.T) {
	s := openTest(t)
	key := "ABCD-000000000002"
	first := []engine.Event{engine.PlayerJoined{Player: "p1", Nickname: "Ân"}, engine.PlayerJoined{Player: "p2", Nickname: "An"}}
	if err := s.Append(key, first...); err != nil {
		t.Fatal(err)
	}
	if err := reopen(t).Append(key, engine.PlayerLeft{Player: "p1"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load(key)
	if err != nil {
		t.Fatal(err)
	}
	want := append(first, engine.PlayerLeft{Player: "p1"})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("log = %#v, want %#v (nicknames kept exactly, in order)", got, want)
	}
	if n := count(t, s, "SELECT COUNT(*) FROM rooms WHERE code = 'ABCD'"); n != 1 {
		t.Fatalf("rooms with code ABCD = %d, want 1", n)
	}
}

func TestMySQL_Load_UnknownRoomIsEmpty_UnknownEventTypeIsAnError(t *testing.T) {
	s := openTest(t)
	if got, err := s.Load("NONE-000000000000"); err != nil || len(got) != 0 {
		t.Fatalf("unknown room: %v, %v; want empty", got, err)
	}
	key := "BAD1-000000000003"
	if err := s.Append(key, engine.NewGameStarted{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB().Exec(`UPDATE room_events SET type = 'Gone' WHERE seq = 0 AND room_id = (SELECT id FROM rooms WHERE log_key = ?)`, key); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(key); err == nil {
		t.Fatal("a log with an unknown event read back without error")
	}
}

// Many rooms write at the same time, each through its own goroutine.
func TestMySQL_Append_ManyRoomsAtOnce_EachLogStaysInOrder(t *testing.T) {
	s := openTest(t)
	const rooms, events = 8, 40
	var wg sync.WaitGroup
	errs := make(chan error, rooms)
	for r := 0; r < rooms; r++ {
		wg.Add(1)
		go func(r int) {
			defer wg.Done()
			key := fmt.Sprintf("R%03d-000000000000", r)
			for i := 0; i < events; i++ {
				if err := s.Append(key, engine.ChatSent{By: "p1", Text: fmt.Sprint(i)}); err != nil {
					errs <- err
					return
				}
			}
		}(r)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	for r := 0; r < rooms; r++ {
		got, err := s.Load(fmt.Sprintf("R%03d-000000000000", r))
		if err != nil || len(got) != events {
			t.Fatalf("room %d: %d events, %v; want %d", r, len(got), err, events)
		}
		for i, e := range got {
			if e.(engine.ChatSent).Text != fmt.Sprint(i) {
				t.Fatalf("room %d event %d = %v, out of order", r, i, e)
			}
		}
	}
}

// A whole game, played by the scenario kit: its log reads back to the same
// state, and its record lands in games and game_players.
func TestMySQL_FinishedGame_LogReplaysAndRecordIsSaved(t *testing.T) {
	s := openTest(t)
	g := NewGame(t, Seats("An", "Binh", "Chi", "Dung", "Em"),
		Cast("An", roles.Werewolf, "Binh", roles.Seer, "Chi", roles.Villager, "Dung", roles.Villager, "Em", roles.Villager))
	g.Night()
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("An")
	g.ExpectPhase(engine.PhaseEnded)

	key := "GAME-000000000004"
	log := g.Events()
	if err := s.Append(key, log...); err != nil {
		t.Fatal(err)
	}
	back, err := reopen(t).Load(key)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(engine.Fold(engine.NewState(), back...), g.State()) {
		t.Fatal("the stored log folds to another state")
	}

	end := log[len(log)-1].(engine.GameEnded)
	rec := store.NewGameRecord(g.State(), end)
	rec.LogKey, rec.Code, rec.GameNo, rec.LastSeq = key, "GAME", 1, len(log)-1
	rec.Started = time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)
	rec.Ended = rec.Started.Add(25 * time.Minute)
	if err := s.SaveGame(rec); err != nil {
		t.Fatal(err)
	}

	var mode, winner string
	var draw, aborted bool
	var nights, players, last int
	var started, ended time.Time
	if err := s.DB().QueryRow(`SELECT mode, winner, draw, aborted, nights, player_count, last_seq, started_at, ended_at
		FROM games WHERE game_no = 1`).Scan(&mode, &winner, &draw, &aborted, &nights, &players, &last, &started, &ended); err != nil {
		t.Fatal(err)
	}
	if mode != "inPerson" || winner != "Village" || draw || aborted || nights != 1 || players != 5 || last != len(log)-1 ||
		!started.Equal(rec.Started) || !ended.Equal(rec.Ended) {
		t.Fatalf("games row = %s %s draw=%v aborted=%v nights=%d players=%d last=%d %v..%v",
			mode, winner, draw, aborted, nights, players, last, started, ended)
	}
	if n := count(t, s, "SELECT COUNT(*) FROM game_players WHERE won"); n != 4 {
		t.Errorf("winners = %d, want the 4 villagers", n)
	}
	var role, team string
	var alive, won bool
	if err := s.DB().QueryRow(`SELECT role, team, alive, won FROM game_players WHERE nickname = 'An'`).Scan(&role, &team, &alive, &won); err != nil {
		t.Fatal(err)
	}
	if role != string(roles.Werewolf) || team != "Wolves" || alive || won {
		t.Errorf("An = %s %s alive=%v won=%v, want a hanged, losing Werewolf", role, team, alive, won)
	}
	if err := s.SaveGame(rec); err == nil {
		t.Error("the same game saved twice")
	}
	if n := count(t, s, "SELECT COUNT(*) FROM game_players"); n != 5 {
		t.Errorf("game_players = %d after the refused second save, want 5 (all or nothing)", n)
	}
}
