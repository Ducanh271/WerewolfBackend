package room

import (
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// gameArchiver follows the log to send a record of each finished game to the
// archive (docs/database.md "Write path"). Room goroutine only.
type gameArchiver struct {
	archive store.GameArchive // nil keeps no records
	first   int               // log index where the current game starts
	started time.Time         // when the current game's roles were dealt
	games   int               // games ended so far in this room
	// tokens is store.TokenHash of each player's current rejoin token (never
	// the token): a record lets its players read it later (POST /history).
	tokens map[engine.PlayerID]string
	// pending holds token changes the archive refused (a database down for
	// a moment); note tries them again, so an old token does not keep
	// reading the saved games for long.
	pending map[engine.PlayerID]string
}

// tokenIssued notes a player's new token: on joining, and on a rejoin by PIN,
// which replaces the old one in the games already saved.
func (a *gameArchiver) tokenIssued(r *Room, id engine.PlayerID, token string) {
	if a.tokens == nil {
		a.tokens = map[engine.PlayerID]string{}
	}
	hash := store.TokenHash(token)
	replaced := a.tokens[id] != ""
	a.tokens[id] = hash
	if replaced && a.archive != nil && a.games > 0 {
		if a.pending == nil {
			a.pending = map[engine.PlayerID]string{}
		}
		a.pending[id] = hash
		a.flushTokens(r)
	}
}

// flushTokens sends the pending token changes to the archive.
func (a *gameArchiver) flushTokens(r *Room) {
	for id, hash := range a.pending {
		if err := a.archive.ReplaceToken(r.logKey, id, hash); err != nil {
			r.logf.Error("game records keep the old token, retrying", "room", r.code, "err", err)
			return
		}
		delete(a.pending, id)
	}
}

// note looks at events just kept at log[first:]; r.state already has them.
// A record that cannot be saved is logged, not returned: the events are safe
// in the store and the record can be rebuilt from them.
func (a *gameArchiver) note(r *Room, first int, evs []engine.Event) {
	if len(a.pending) > 0 {
		a.flushTokens(r)
	}
	for i, e := range evs {
		switch ev := e.(type) {
		case engine.RolesDealt:
			a.started = r.clock.Now()
		case engine.GameEnded:
			a.games++
			seq := first + i
			if a.archive != nil {
				a.save(r, ev, seq)
			}
			a.first = seq + 1
		}
	}
}

func (a *gameArchiver) save(r *Room, end engine.GameEnded, seq int) {
	rec := store.NewGameRecord(r.state, end)
	for i := range rec.Players {
		rec.Players[i].TokenHash = a.tokens[rec.Players[i].ID]
	}
	rec.LogKey, rec.Code, rec.GameNo = r.logKey, r.code, a.games
	rec.FirstSeq, rec.LastSeq = a.first, seq
	rec.Ended = r.clock.Now()
	rec.Started = a.started
	if rec.Started.IsZero() {
		rec.Started = rec.Ended
	}
	if err := a.archive.SaveGame(rec); err != nil {
		r.logf.Error("game record not saved", "room", r.code, "game", a.games, "err", err)
	}
}
