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
}

// note looks at events just kept at log[first:]; r.state already has them.
// A record that cannot be saved is logged, not returned: the events are safe
// in the store and the record can be rebuilt from them.
func (a *gameArchiver) note(r *Room, first int, evs []engine.Event) {
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
