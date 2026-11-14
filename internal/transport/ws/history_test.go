package ws

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/room"
	"masoi/internal/store"
)

var historyNow = time.Date(2026, 11, 14, 20, 0, 0, 0, time.UTC)

func historyServer(t *testing.T, h store.History, logs io.Writer) *httptest.Server {
	t.Helper()
	rooms := room.NewManager(room.ManagerConfig{})
	t.Cleanup(rooms.Close)
	if logs == nil {
		logs = io.Discard
	}
	handler, err := New(Config{Rooms: rooms, History: h, AllowAnyOrigin: true,
		Now: func() time.Time { return historyNow }, Logger: slog.New(slog.NewTextHandler(logs, nil))})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func postHistory(t *testing.T, srv *httptest.Server, body string) (int, historyReply, map[string]any) {
	t.Helper()
	resp, err := http.Post(srv.URL+"/history", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var reply historyReply
	var generic map[string]any
	_ = json.Unmarshal(raw, &reply)
	_ = json.Unmarshal(raw, &generic)
	return resp.StatusCode, reply, generic
}

// savedGame is a finished 3-seat draw that ended `ago` before historyNow; the
// seats' tokens are "tok-<game>-<seat>".
func savedGame(t *testing.T, m *store.Memory, name string, ago time.Duration) {
	t.Helper()
	rec := store.GameRecord{LogKey: "K7QP-" + name, Code: "K7QP", GameNo: 1, Draw: true, Aborted: true, Nights: 2,
		Started: historyNow.Add(-ago - 30*time.Minute), Ended: historyNow.Add(-ago)}
	for i, role := range []engine.RoleID{"Werewolf", "Seer", "Villager"} {
		rec.Players = append(rec.Players, store.GamePlayer{
			ID: engine.PlayerID("p" + string(rune('1'+i))), Seat: i, Nickname: "N" + string(rune('1'+i)),
			Role: role, Team: engine.TeamVillage, Alive: true, TokenHash: store.TokenHash("tok-" + name + "-" + string(rune('0'+i))),
		})
	}
	if err := m.SaveGame(rec); err != nil {
		t.Fatal(err)
	}
}

func TestHistory_OnlyThePlayersOfAGameReadIt(t *testing.T) {
	m := store.NewMemory()
	savedGame(t, m, "a", time.Hour)
	savedGame(t, m, "b", 2*time.Hour)
	srv := historyServer(t, m, nil)

	tests := []struct {
		name   string
		tokens []string
		want   []string // rooms' log keys, newest first
		you    int      // the seat marked "you" in each game
	}{
		{"a stranger's token finds nothing", []string{"not-a-token"}, nil, -1},
		{"a player reads their game", []string{"tok-a-1"}, []string{"a"}, 1},
		{"tokens of two rooms read both, newest first", []string{"tok-b-2", "tok-a-2"}, []string{"a", "b"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]any{"tokens": tt.tokens})
			status, reply, _ := postHistory(t, srv, string(body))
			if status != 200 || len(reply.Games) != len(tt.want) {
				t.Fatalf("status %d, %d games; want 200, %d", status, len(reply.Games), len(tt.want))
			}
			for i, g := range reply.Games {
				if g.EndedAt != historyNow.Add(-time.Duration(i+1)*time.Hour).UnixMilli() {
					t.Errorf("game %d ended %d, out of order", i, g.EndedAt)
				}
				if g.Winner != "" || !g.Draw || !g.Aborted || g.Room != "K7QP" || len(g.Players) != 3 {
					t.Errorf("game %d = %+v, want the saved aborted draw without a winner", i, g)
				}
				for _, p := range g.Players {
					if p.You != (p.Seat == tt.you) {
						t.Errorf("game %d seat %d you = %v, want only seat %d", i, p.Seat, p.You, tt.you)
					}
				}
			}
		})
	}
}

// Games are kept 90 days (docs/database.md "Retention").
func TestHistory_GamesOlderThanTheRetentionAreNotRead(t *testing.T) {
	m := store.NewMemory()
	savedGame(t, m, "old", store.Retention+time.Hour)
	savedGame(t, m, "new", store.Retention-time.Hour)
	_, reply, _ := postHistory(t, historyServer(t, m, nil), `{"tokens":["tok-old-0","tok-new-0"]}`)
	if len(reply.Games) != 1 || reply.Games[0].EndedAt != historyNow.Add(-store.Retention+time.Hour).UnixMilli() {
		t.Fatalf("games = %+v, want only the one inside 90 days", reply.Games)
	}
}

func TestHistory_BadRequestsAreRefused(t *testing.T) {
	srv := historyServer(t, store.NewMemory(), nil)
	many, _ := json.Marshal(map[string]any{"tokens": strings.Split(strings.Repeat("t,", historyMaxTokens+1), ",")[:historyMaxTokens+1]})
	tests := []struct {
		name, body string
		status     int
		code       string
	}{
		{"no tokens", `{"tokens":[]}`, 400, codeBadMessage},
		{"missing tokens", `{}`, 400, codeBadMessage},
		{"too many tokens", string(many), 400, codeBadMessage},
		{"empty token", `{"tokens":[""]}`, 400, codeBadMessage},
		{"token too long", `{"tokens":["` + strings.Repeat("x", historyMaxToken+1) + `"]}`, 400, codeBadMessage},
		{"unknown field", `{"tokens":["a"],"all":true}`, 400, codeBadMessage},
		{"not json", `tokens`, 400, codeBadMessage},
		{"body too large", `{"tokens":["` + strings.Repeat("x", historyMaxBody) + `"]}`, 413, codeTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, body := postHistory(t, srv, tt.body)
			if status != tt.status || body["code"] != tt.code {
				t.Fatalf("status %d code %v, want %d %s", status, body["code"], tt.status, tt.code)
			}
		})
	}
}

func TestHistory_ServerWithoutDatabase_AnswersNoGames(t *testing.T) {
	status, reply, body := postHistory(t, historyServer(t, nil, nil), `{"tokens":["tok-a-0"]}`)
	if status != 200 || reply.Games == nil || len(reply.Games) != 0 {
		t.Fatalf("status %d body %v, want 200 with an empty games list", status, body)
	}
}

type downHistory struct{}

func (downHistory) GamesFor([]string, time.Time, int) ([]store.SavedGame, error) {
	return nil, errors.New("connection refused")
}

// The database failing is told as history_unavailable, and the log never
// holds a token.
func TestHistory_DatabaseDown_IsUnavailable_TokensStayOutOfTheLog(t *testing.T) {
	var logs bytes.Buffer
	status, _, body := postHistory(t, historyServer(t, downHistory{}, &logs), `{"tokens":["secret-token-123"]}`)
	if status != 503 || body["code"] != codeHistoryDown {
		t.Fatalf("status %d code %v, want 503 %s", status, body["code"], codeHistoryDown)
	}
	if strings.Contains(logs.String(), "secret-token-123") || strings.Contains(logs.String(), store.TokenHash("secret-token-123")) {
		t.Fatalf("log holds the token: %s", logs.String())
	}
}

// The lovers' win goes out as winner "Lovers" while each lover keeps their own
// team (docs/api/screens/history.md: mark winners with won, not team).
func TestHistory_LoversWin_WinnerIsLoversTeamsStayOwn(t *testing.T) {
	m := store.NewMemory()
	rec := store.GameRecord{LogKey: "LOVE-1", Code: "LOVE", GameNo: 1, Winner: engine.TeamLovers,
		Started: historyNow.Add(-time.Hour), Ended: historyNow}
	for i, team := range []engine.Team{engine.TeamWolves, engine.TeamVillage, engine.TeamVillage} {
		rec.Players = append(rec.Players, store.GamePlayer{ID: engine.PlayerID("p" + string(rune('1'+i))), Seat: i,
			Nickname: "N", Role: "Villager", Team: team, Won: i < 2, TokenHash: store.TokenHash("love-" + string(rune('0'+i)))})
	}
	if err := m.SaveGame(rec); err != nil {
		t.Fatal(err)
	}
	_, reply, _ := postHistory(t, historyServer(t, m, nil), `{"tokens":["love-2"]}`)
	if len(reply.Games) != 1 {
		t.Fatalf("games = %d, want 1", len(reply.Games))
	}
	g := reply.Games[0]
	if g.Winner != "Lovers" || g.Draw || g.Players[0].Team != "Wolves" || g.Players[1].Team != "Village" ||
		!g.Players[0].Won || !g.Players[1].Won || g.Players[2].Won || !g.Players[2].You {
		t.Fatalf("game = %+v", g)
	}
}
