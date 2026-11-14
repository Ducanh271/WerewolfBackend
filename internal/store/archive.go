package store

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"masoi/internal/engine"
)

// GameArchive keeps a summary of every finished game (docs/database.md). The
// summary is derived from the event log, so losing one is not losing the game.
type GameArchive interface {
	SaveGame(GameRecord) error
	// ReplaceToken follows a player's new rejoin token (rejoin by PIN hands
	// out a new one and the old one stops working) into the room's saved
	// games, so the new phone can read them and the old one no longer can.
	ReplaceToken(logKey string, player engine.PlayerID, tokenHash string) error
}

// History reads the finished games a player took part in, for POST /history
// (docs/api/screens/history.md).
type History interface {
	// GamesFor returns the games ended at or after since in which one of the
	// token hashes played, newest first, at most limit of them.
	GamesFor(tokenHashes []string, since time.Time, limit int) ([]SavedGame, error)
}

// Retention is how long a finished game is kept (docs/database.md).
const Retention = 90 * 24 * time.Hour

// SavedGame is a GameRecord with the id the archive gave it.
type SavedGame struct {
	ID int64
	GameRecord
}

// TokenHash is what the archive keeps of a rejoin token: never the token
// itself. Tokens are long random secrets, so a plain SHA-256 is enough.
func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
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
	// TokenHash is TokenHash of the player's rejoin token when the game
	// ended: what lets them read the game later. Empty when unknown.
	TokenHash string
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
