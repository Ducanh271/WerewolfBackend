package store

import (
	"time"

	"masoi/internal/engine"
)

// GameArchive keeps a summary of every finished game (docs/database.md). The
// summary is derived from the event log, so losing one is not losing the game.
type GameArchive interface {
	SaveGame(GameRecord) error
}

// GameRecord is one finished game.
type GameRecord struct {
	LogKey   string // the room's event log (EventStore key)
	Code     string // the room code, as players typed it
	GameNo   int    // 1 for the room's first game, +1 for each newGame
	Mode     engine.Mode
	Winner   engine.Team // TeamNone on a draw
	Draw     bool
	Aborted  bool
	Nights   int
	FirstSeq int // the game's events are the log's [FirstSeq, LastSeq]
	LastSeq  int
	Started  time.Time // roles dealt
	Ended    time.Time
	Players  []GamePlayer // seating order
}

// GamePlayer is one player of a finished game.
type GamePlayer struct {
	ID       engine.PlayerID
	Seat     int
	Nickname string
	Role     engine.RoleID
	Team     engine.Team // at the end, after a curse
	Alive    bool
	Won      bool
}

// NewGameRecord builds the record of the game that end has just closed; s is
// the state with end applied. The caller fills in what neither knows: the log
// position, the room and the times.
func NewGameRecord(s engine.State, end engine.GameEnded) GameRecord {
	rec := GameRecord{
		Mode: s.Settings.Mode, Nights: s.Night,
		Winner: end.Outcome.Winner, Draw: end.Outcome.Draw, Aborted: end.Outcome.Aborted,
	}
	if rec.Draw {
		rec.Winner = engine.TeamNone
	}
	won := map[engine.PlayerID]bool{}
	for _, id := range end.Outcome.Winners {
		won[id] = true
	}
	for i, id := range s.Seats {
		p := s.Players[id]
		if p == nil {
			continue
		}
		rec.Players = append(rec.Players, GamePlayer{
			ID: id, Seat: i, Nickname: p.Nickname, Role: p.Role,
			Team: p.Team, Alive: p.Alive, Won: won[id],
		})
	}
	return rec
}
