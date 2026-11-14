package ws

import (
	"encoding/json"
)

// historySchemas adds the bodies of POST /history (docs/api/screens/history.md).
func historySchemas(g *schemaGen) {
	integer := func(desc string) map[string]any {
		return map[string]any{"type": "integer", "minimum": 0, "description": desc}
	}
	boolean := func(desc string) map[string]any { return map[string]any{"type": "boolean", "description": desc} }
	teams := func(desc string) map[string]any {
		return map[string]any{"type": "string", "enum": teamNames(), "description": desc}
	}
	g.defs["HistoryRequest"] = object(map[string]any{
		"tokens": withDescription(map[string]any{
			"type": "array", "minItems": 1, "maxItems": historyMaxTokens,
			"items": map[string]any{"type": "string", "minLength": 1, "maxLength": historyMaxToken},
		}, "The rejoin tokens this phone kept, one per room it played in. Secret: only in the body, never in the URL."),
	}, "tokens")
	g.defs["HistoryPlayer"] = object(map[string]any{
		"seat":     integer("0-based, clockwise."),
		"player":   g.tagValue(playerIDType),
		"nickname": str(),
		"role":     g.tagValue(roleIDType),
		"team":     teams("The player's own team at the end, after a curse. A third-team lover keeps Village or Wolves (rules §7); see won."),
		"alive":    boolean("Alive when the game ended."),
		"won":      boolean("In the game's winners (rules §8)."),
		"you":      boolean("This seat belongs to one of the request's tokens."),
	}, "seat", "player", "nickname", "role", "team", "alive", "won", "you")
	g.defs["HistoryGame"] = object(map[string]any{
		"id":        integer("Stable id of the saved game."),
		"room":      roomCodeSchema(),
		"gameNo":    map[string]any{"type": "integer", "minimum": 1, "description": "1 for the room's first game, +1 for each newGame."},
		"mode":      map[string]any{"type": "string", "enum": gameModeNames()},
		"winner":    teams("Absent on a draw."),
		"draw":      boolean("Nobody won: everybody died at once, or the game was aborted."),
		"aborted":   boolean("Called off by an abort vote (rules §8b); always a draw too."),
		"nights":    integer("Nights played."),
		"startedAt": integer("Unix milliseconds: roles dealt."),
		"endedAt":   integer("Unix milliseconds."),
		"players":   map[string]any{"type": "array", "items": ref("HistoryPlayer"), "description": "Seating order."},
	}, "id", "room", "gameNo", "mode", "draw", "aborted", "nights", "startedAt", "endedAt", "players")
	g.defs["HistoryReply"] = object(map[string]any{
		"games": map[string]any{"type": "array", "maxItems": historyMaxGames, "items": ref("HistoryGame"),
			"description": "Newest first, finished games only, kept 90 days. Empty when nothing matches or the server keeps no history."},
	}, "games")
}

// historyExample asks for An's games once the first game has ended.
func (g *exGame) historyExample() {
	t := g.t
	req, _ := json.Marshal(map[string]any{"tokens": []string{g.phone("An").token}})
	status, body := g.httpJSON("POST", "/history", req)
	if status != 200 {
		t.Fatalf("POST /history: %d %s", status, body)
	}
	var reply historyReply
	mustUnmarshal(t, body, &reply)
	if len(reply.Games) != 1 {
		t.Fatalf("POST /history after the first game: %d games, want 1", len(reply.Games))
	}
	g.recordHTTP("history.request", "HistoryRequest", req)
	g.recordHTTP("history.reply", "HistoryReply", body)
	status, body = g.httpJSON("POST", "/history", []byte(`{"tokens":[]}`))
	if status != 400 {
		t.Fatalf("POST /history with no token: %d", status)
	}
	g.recordHTTP("history.error.bad_message", "HTTPError", body)
}
