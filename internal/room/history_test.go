package room

import (
	"errors"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
	"masoi/internal/store"
)

// A saved game carries a hash of each player's rejoin token, never the token:
// that is what lets its players, and only them, read it later.
func TestRoom_GameRecord_HoldsEachPlayersTokenHash(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	g.abortByVote()

	rec := g.store.Games()[0]
	for _, p := range rec.Players {
		want := store.TokenHash(g.cred[p.ID].Token)
		if p.TokenHash != want || p.TokenHash == g.cred[p.ID].Token {
			t.Errorf("%s: token hash %q, want %q (the hash, not the token)", p.ID, p.TokenHash, want)
		}
	}
}

// Rejoining by PIN hands out a new token and the old one stops working
// (rules §2 "Vào lại phòng"): the same goes for reading the saved games.
func TestRoom_RejoinByPIN_MovesTheSavedGamesToTheNewToken(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	g.abortByVote()

	old := g.cred["p3"]
	_, phone := g.newSink("")
	_, creds, err := g.r.RejoinPIN(g.nick["p3"], old.PIN, phone)
	if err != nil {
		t.Fatal(err)
	}
	since := g.clk.Now().Add(-time.Hour)
	if got, _ := g.store.GamesFor([]string{store.TokenHash(old.Token)}, since, 10); len(got) != 0 {
		t.Errorf("the old token still reads %d games, want none", len(got))
	}
	if got, _ := g.store.GamesFor([]string{store.TokenHash(creds.Token)}, since, 10); len(got) != 1 {
		t.Errorf("the new token reads %d games, want 1", len(got))
	}
	if got, _ := g.store.GamesFor([]string{store.TokenHash(g.cred["p2"].Token)}, since, 10); len(got) != 1 {
		t.Errorf("another player's token reads %d games, want 1 (unchanged)", len(got))
	}
}

// Only finished games are saved, so a game being played is never read.
func TestRoom_GameInProgress_IsNotInTheHistory(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	tok := store.TokenHash(g.cred["p1"].Token)
	if got, _ := g.store.GamesFor([]string{tok}, g.clk.Now().Add(-time.Hour), 10); len(got) != 0 || len(g.store.Games()) != 0 {
		t.Fatalf("a game being played was saved: %d records", len(g.store.Games()))
	}
}

// A player who sat in a game keeps reading it after being kicked from the
// lobby of the next one: they played it (docs/api/screens/history.md).
func TestRoom_KickedBetweenGames_StillReadsTheGamesTheyPlayed(t *testing.T) {
	g := newGame(t, 6)
	order := []engine.PlayerID{"p1", "p2", "p3", "p4", "p5"}
	cast := map[engine.RoleID]int{roles.Werewolf: 1, roles.Guard: 1, roles.Seer: 1, roles.Villager: 2}
	g.must(g.cl["p1"].KickPlayer("p6")) // the table is the five in order
	g.must(g.cl["p1"].Submit(as(engine.SetSeats{Order: order})))
	g.must(g.cl["p1"].Submit(as(engine.ConfigureRoles{Roles: cast})))
	g.must(g.cl["p1"].Submit(as(engine.DealRoles{})))
	g.toDay()
	g.abortByVote()
	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))
	g.must(g.cl["p1"].KickPlayer("p5"))

	got, _ := g.store.GamesFor([]string{store.TokenHash(g.cred["p5"].Token)}, g.clk.Now().Add(-time.Hour), 10)
	if len(got) != 1 {
		t.Fatalf("the kicked player reads %d games, want the 1 they played", len(got))
	}
	if got, _ := g.store.GamesFor([]string{store.TokenHash(g.cred["p6"].Token)}, g.clk.Now().Add(-time.Hour), 10); len(got) != 0 {
		t.Fatalf("a player kicked before the deal reads %d games, want none", len(got))
	}
}

// flakyTokens refuses the next `fail` token changes.
type flakyTokens struct {
	*store.Memory
	fail int
}

func (f *flakyTokens) ReplaceToken(logKey string, p engine.PlayerID, hash string) error {
	if f.fail > 0 {
		f.fail--
		return errors.New("connection refused")
	}
	return f.Memory.ReplaceToken(logKey, p, hash)
}

// A token change the archive refused is tried again with the room's next
// event, so the old token stops reading the saved games.
func TestRoom_RejoinByPIN_ArchiveDown_TokenChangeIsRetried(t *testing.T) {
	g := newGame(t, 5)
	arch := &flakyTokens{Memory: g.store, fail: 1}
	_ = g.r.do(func() error { g.r.arch.archive = arch; return nil })
	g.deal()
	g.toDay()
	g.abortByVote()

	old := g.cred["p3"]
	_, phone := g.newSink("")
	_, creds, err := g.r.RejoinPIN(g.nick["p3"], old.PIN, phone)
	if err != nil {
		t.Fatal(err)
	}
	since := g.clk.Now().Add(-time.Hour)
	if got, _ := g.store.GamesFor([]string{store.TokenHash(creds.Token)}, since, 10); len(got) != 0 {
		t.Fatal("the refused change was applied anyway")
	}
	g.must(g.cl["p1"].Submit(as(engine.NewGame{}))) // the room's next event
	if got, _ := g.store.GamesFor([]string{store.TokenHash(old.Token)}, since, 10); len(got) != 0 {
		t.Errorf("the old token still reads %d games after the retry", len(got))
	}
	if got, _ := g.store.GamesFor([]string{store.TokenHash(creds.Token)}, since, 10); len(got) != 1 {
		t.Errorf("the new token reads %d games after the retry, want 1", len(got))
	}
}
