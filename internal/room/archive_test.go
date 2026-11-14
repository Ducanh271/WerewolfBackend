package room

import (
	"bytes"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// abortByVote has every living player vote to call the game off (rules §8b),
// which ends it by day without playing it out.
func (g *game) abortByVote() {
	g.t.Helper()
	for id, p := range g.snap().Players {
		if p.Alive && g.snap().Phase != engine.PhaseEnded {
			g.must(g.cl[id].Submit(func(by engine.PlayerID) engine.Command {
				return engine.CastAbortVote{Voter: by, Yes: true}
			}))
		}
	}
	if g.snap().Phase != engine.PhaseEnded {
		g.t.Fatalf("abort vote did not end the game, phase = %s", g.snap().Phase)
	}
}

func lastIndexOf[T engine.Event](log []engine.Event) int {
	for i := len(log) - 1; i >= 0; i-- {
		if _, ok := log[i].(T); ok {
			return i
		}
	}
	return -1
}

func TestRoom_GameEnded_SavesARecordOfTheGame(t *testing.T) {
	g := newGame(t, 5)
	dealtAt := g.clk.Now()
	g.deal()
	g.toDay()
	g.abortByVote()

	games := g.store.Games()
	if len(games) != 1 {
		t.Fatalf("records = %d, want 1", len(games))
	}
	rec, s, log := games[0], g.snap(), g.eventLog()
	if rec.LogKey != "K7QP" || rec.Code != "K7QP" || rec.GameNo != 1 {
		t.Errorf("room = %q/%q #%d, want K7QP/K7QP #1", rec.LogKey, rec.Code, rec.GameNo)
	}
	if !rec.Draw || !rec.Aborted || rec.Winner != engine.TeamNone {
		t.Errorf("outcome = draw %v aborted %v winner %v, want an aborted draw", rec.Draw, rec.Aborted, rec.Winner)
	}
	if rec.FirstSeq != 0 || rec.LastSeq != lastIndexOf[engine.GameEnded](log) || rec.LastSeq != len(log)-1 {
		t.Errorf("seq = %d..%d, want 0..%d (the GameEnded event)", rec.FirstSeq, rec.LastSeq, len(log)-1)
	}
	if !rec.Started.Equal(dealtAt) || !rec.Ended.Equal(g.clk.Now()) || rec.Nights != 1 {
		t.Errorf("times = %v..%v nights %d, want %v..%v nights 1", rec.Started, rec.Ended, rec.Nights, dealtAt, g.clk.Now())
	}
	if len(rec.Players) != 5 {
		t.Fatalf("players = %d, want 5", len(rec.Players))
	}
	for i, p := range rec.Players {
		want := s.Players[s.Seats[i]]
		if p.ID != want.ID || p.Seat != i || p.Nickname != want.Nickname || p.Role != want.Role ||
			p.Team != want.Team || p.Alive != want.Alive || p.Won {
			t.Errorf("seat %d = %+v, want %+v, not a winner", i, p, *want)
		}
	}
}

func TestRoom_NewGame_SecondGameGetsTheNextNumberAndItsOwnEvents(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	g.abortByVote()
	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))
	g.rnd.setSeed(2)
	g.deal()
	g.toDay()
	g.abortByVote()

	games := g.store.Games()
	if len(games) != 2 {
		t.Fatalf("records = %d, want 2", len(games))
	}
	if games[1].GameNo != 2 || games[1].FirstSeq != games[0].LastSeq+1 || games[1].LastSeq != len(g.eventLog())-1 {
		t.Errorf("second game #%d seq %d..%d, want #2 from %d to %d",
			games[1].GameNo, games[1].FirstSeq, games[1].LastSeq, games[0].LastSeq+1, len(g.eventLog())-1)
	}
}

type failingArchive struct{}

func (failingArchive) SaveGame(store.GameRecord) error { return errors.New("disk full") }

// The events are what counts: a record that cannot be saved is logged and the
// room goes on (docs/database.md "Write path").
func TestRoom_ArchiveFails_GameStillEndsAndTheErrorIsLogged(t *testing.T) {
	var logs bytes.Buffer
	g := newGame(t, 5)
	_ = g.r.do(func() error {
		g.r.arch.archive = failingArchive{}
		g.r.logf = slog.New(slog.NewTextHandler(&logs, nil))
		return nil
	})
	g.deal()
	g.toDay()
	g.abortByVote()

	if !strings.Contains(logs.String(), "game record not saved") || !strings.Contains(logs.String(), "disk full") {
		t.Fatalf("log = %q, want the failure", logs.String())
	}
	if n := len(g.eventLog()); n == 0 {
		t.Fatal("no events")
	}
}

// Room codes come back once a room is gone, so rooms sharing a store each get
// their own log key.
func TestManager_SharedStore_EachRoomLogsUnderItsOwnKey(t *testing.T) {
	shared := store.NewMemory()
	m := NewManager(ManagerConfig{Clock: NewFakeClock(time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)), Random: &fixedRandom{seed: 1}, Store: shared, Archive: shared})
	t.Cleanup(m.Close)
	a, _, err := m.Create(engine.InPerson)
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := m.Create(engine.Online)
	if err != nil {
		t.Fatal(err)
	}
	key := regexp.MustCompile(`^[A-Z0-9]{4}-[0-9a-f]{12}$`)
	for _, r := range []*Room{a, b} {
		if !key.MatchString(r.logKey) || !strings.HasPrefix(r.logKey, r.Code()+"-") {
			t.Errorf("log key %q, want %s-<12 hex>", r.logKey, r.Code())
		}
	}
	// An online room starts its log with ModeConfigured (the in-person default
	// needs no event).
	if log, err := shared.Load(b.logKey); err != nil || len(log) == 0 || log[0] != (engine.ModeConfigured{Mode: engine.Online}) {
		t.Errorf("online room %s: log = %v, %v; want its ModeConfigured", b.logKey, log, err)
	}
	if a.logKey == b.logKey {
		t.Fatalf("both rooms log under %q", a.logKey)
	}
}

// flakyStore fails the next `fail` appends, as a database down for a moment.
type flakyStore struct {
	store.EventStore
	fail int
}

func (f *flakyStore) Append(room string, evs ...engine.Event) error {
	if f.fail > 0 {
		f.fail--
		return errors.New("connection refused")
	}
	return f.EventStore.Append(room, evs...)
}

// A timeout is the only way a phase ends: one the store could not keep is
// logged and tried again, instead of leaving the room standing still.
func TestRoom_StoreFailsOnATimeout_RoomTriesAgainAndGoesOn(t *testing.T) {
	var logs bytes.Buffer
	g := newGame(t, 5)
	g.deal()
	_ = g.r.do(func() error {
		g.r.store = &flakyStore{EventStore: g.store, fail: 1}
		g.r.logf = slog.New(slog.NewTextHandler(&logs, nil))
		return nil
	})
	g.step(engine.DefaultSettings().RoleView)
	if s := g.snap(); s.Phase != engine.PhaseRoleView {
		t.Fatalf("phase = %s after a timeout the store refused, want RoleView unchanged", s.Phase)
	}
	if !strings.Contains(logs.String(), "timeout not stored, retrying") {
		t.Fatalf("log = %q, want the retry", logs.String())
	}
	g.step(storeRetry)
	if s := g.snap(); s.Phase == engine.PhaseRoleView {
		t.Fatal("still in RoleView after the store came back")
	}
	stored, _ := g.store.Load("K7QP")
	if len(stored) != len(g.eventLog()) {
		t.Fatalf("store has %d events, room %d: they must match", len(stored), len(g.eventLog()))
	}
}
