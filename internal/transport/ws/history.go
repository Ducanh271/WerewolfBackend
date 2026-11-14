package ws

import (
	"io"
	"net/http"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// History limits (docs/api/screens/history.md).
const (
	historyMaxTokens = 20  // rooms one request may ask about
	historyMaxToken  = 128 // characters in one token
	historyMaxGames  = 50  // games in one reply
	historyMaxBody   = 8 << 10
)

// historyRequest is the body of POST /history: the rejoin tokens the phone
// kept, one per room it played in. Tokens travel in the body, never in the
// URL, which ends up in logs.
type historyRequest struct {
	Tokens []string `json:"tokens"`
}

// historyReply is the body of the answer.
type historyReply struct {
	Games []historyGame `json:"games"`
}

type historyGame struct {
	ID        int64           `json:"id"`
	Room      string          `json:"room"`
	GameNo    int             `json:"gameNo"`
	Mode      string          `json:"mode"`
	Winner    string          `json:"winner,omitempty"`
	Draw      bool            `json:"draw"`
	Aborted   bool            `json:"aborted"`
	Nights    int             `json:"nights"`
	StartedAt int64           `json:"startedAt"`
	EndedAt   int64           `json:"endedAt"`
	Players   []historyPlayer `json:"players"`
}

type historyPlayer struct {
	Seat     int             `json:"seat"`
	Player   engine.PlayerID `json:"player"`
	Nickname string          `json:"nickname"`
	Role     engine.RoleID   `json:"role"`
	Team     string          `json:"team"`
	Alive    bool            `json:"alive"`
	Won      bool            `json:"won"`
	You      bool            `json:"you"`
}

// history answers POST /history: the finished games, kept store.Retention,
// that one of the tokens played in. Only finished games are ever saved, so a
// game still being played is never in the answer. An unknown token simply
// finds nothing: the reply does not say which tokens were known.
func (s *Server) history(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, historyMaxBody))
	if err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, errorFrameMap(&protoError{codeTooLarge, "body too large"}))
		return
	}
	var req historyRequest
	if err := decodeStrict(raw, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorFrameMap(err))
		return
	}
	if len(req.Tokens) == 0 || len(req.Tokens) > historyMaxTokens {
		writeJSON(w, http.StatusBadRequest, errorFrameMap(badMessage("tokens: 1 to 20 of them")))
		return
	}
	mine := map[string]bool{}
	hashes := make([]string, 0, len(req.Tokens))
	for _, tok := range req.Tokens {
		if tok == "" || len(tok) > historyMaxToken {
			writeJSON(w, http.StatusBadRequest, errorFrameMap(badMessage("tokens: empty or too long")))
			return
		}
		h := store.TokenHash(tok)
		mine[h] = true
		hashes = append(hashes, h)
	}
	reply := historyReply{Games: []historyGame{}}
	if s.cfg.History == nil { // no database: nothing is kept
		writeJSON(w, http.StatusOK, reply)
		return
	}
	games, err := s.cfg.History.GamesFor(hashes, s.cfg.Now().Add(-store.Retention), historyMaxGames)
	if err != nil {
		s.log.Error("history not read", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, errorFrameMap(&protoError{codeHistoryDown, "history unavailable, try again later"}))
		return
	}
	for _, g := range games {
		reply.Games = append(reply.Games, historyGameOf(g, mine))
	}
	writeJSON(w, http.StatusOK, reply)
}

func historyGameOf(g store.SavedGame, mine map[string]bool) historyGame {
	out := historyGame{
		ID: g.ID, Room: g.Code, GameNo: g.GameNo, Mode: gameModeName(g.Mode),
		Draw: g.Draw, Aborted: g.Aborted, Nights: g.Nights,
		StartedAt: g.Started.UnixMilli(), EndedAt: g.Ended.UnixMilli(),
		Players: make([]historyPlayer, 0, len(g.Players)),
	}
	if !g.Draw && g.Winner != engine.TeamNone {
		out.Winner = g.Winner.String()
	}
	for _, p := range g.Players {
		out.Players = append(out.Players, historyPlayer{
			Seat: p.Seat, Player: p.ID, Nickname: p.Nickname, Role: p.Role, Team: p.Team.String(),
			Alive: p.Alive, Won: p.Won, You: p.TokenHash != "" && mine[p.TokenHash],
		})
	}
	return out
}
