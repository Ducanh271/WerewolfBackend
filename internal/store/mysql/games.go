package mysql

import (
	"context"
	"fmt"
	"strings"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// SaveGame writes a finished game and its players in one transaction.
func (s *Store) SaveGame(rec store.GameRecord) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	r, err := s.room(ctx, rec.LogKey)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mysql: save game: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	res, err := tx.ExecContext(ctx, `INSERT INTO games
		(room_id, game_no, mode, winner, draw, aborted, nights, player_count, first_seq, last_seq, started_at, ended_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.id, rec.GameNo, modeName(rec.Mode), teamName(rec.Winner), rec.Draw, rec.Aborted, rec.Nights,
		len(rec.Players), rec.FirstSeq, rec.LastSeq, rec.Started.UTC(), rec.Ended.UTC())
	if err != nil {
		return fmt.Errorf("mysql: save game %s #%d: %w", rec.LogKey, rec.GameNo, err)
	}
	gameID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("mysql: save game: %w", err)
	}
	if len(rec.Players) > 0 {
		var sb strings.Builder
		sb.WriteString("INSERT INTO game_players (game_id, player_id, seat, nickname, role, team, alive, won) VALUES ")
		args := make([]any, 0, 8*len(rec.Players))
		for i, p := range rec.Players {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?, ?, ?, ?, ?, ?, ?, ?)")
			args = append(args, gameID, string(p.ID), p.Seat, p.Nickname, string(p.Role), teamName(p.Team), p.Alive, p.Won)
		}
		if _, err = tx.ExecContext(ctx, sb.String(), args...); err != nil {
			return fmt.Errorf("mysql: save players of %s #%d: %w", rec.LogKey, rec.GameNo, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("mysql: save game: %w", err)
	}
	return nil
}

// modeName is the mode as docs/api names it.
func modeName(m engine.Mode) string {
	if m == engine.Online {
		return "online"
	}
	return "inPerson"
}

// teamName is the team as docs/api names it, "" for none.
func teamName(t engine.Team) string {
	if t == engine.TeamNone {
		return ""
	}
	return t.String()
}
