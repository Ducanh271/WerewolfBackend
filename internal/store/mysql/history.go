package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
)

var _ store.History = (*Store)(nil)

// ReplaceToken points the player's rows in the room's saved games at their
// new token.
func (s *Store) ReplaceToken(logKey string, player engine.PlayerID, tokenHash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	if _, err := s.db.ExecContext(ctx, `UPDATE game_players p
		JOIN games g ON g.id = p.game_id JOIN rooms r ON r.id = g.room_id
		SET p.token_hash = ? WHERE r.log_key = ? AND p.player_id = ?`, tokenHash, logKey, string(player)); err != nil {
		return fmt.Errorf("mysql: replace token in %s: %w", logKey, err)
	}
	return nil
}

// GamesFor returns the finished games, ended at or after since, that one of
// the token hashes played in: newest first, at most limit.
func (s *Store) GamesFor(tokenHashes []string, since time.Time, limit int) ([]store.SavedGame, error) {
	var hashes []any
	for _, h := range tokenHashes {
		if h != "" {
			hashes = append(hashes, h)
		}
	}
	if len(hashes) == 0 || limit <= 0 {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	marks := strings.TrimSuffix(strings.Repeat("?, ", len(hashes)), ", ")
	args := append(append([]any{}, hashes...), since.UTC(), limit)
	rows, err := s.db.QueryContext(ctx, `SELECT g.id, r.log_key, r.code, g.game_no, g.mode, g.winner, g.draw, g.aborted,
		g.nights, g.first_seq, g.last_seq, g.started_at, g.ended_at
		FROM games g JOIN rooms r ON r.id = g.room_id
		WHERE g.id IN (SELECT game_id FROM game_players WHERE token_hash IN (`+marks+`)) AND g.ended_at >= ?
		ORDER BY g.ended_at DESC, g.id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: history: %w", err)
	}
	var out []store.SavedGame
	byID := map[int64]int{}
	for rows.Next() {
		var g store.SavedGame
		var mode, winner string
		if err := rows.Scan(&g.ID, &g.LogKey, &g.Code, &g.GameNo, &mode, &winner, &g.Draw, &g.Aborted,
			&g.Nights, &g.FirstSeq, &g.LastSeq, &g.Started, &g.Ended); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("mysql: history: %w", err)
		}
		g.Mode, g.Winner = modeOf(mode), teamOf(winner)
		byID[g.ID] = len(out)
		out = append(out, g)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("mysql: history: %w", err)
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, s.addPlayers(ctx, out, byID)
}

func (s *Store) addPlayers(ctx context.Context, games []store.SavedGame, byID map[int64]int) error {
	ids := make([]any, 0, len(games))
	for _, g := range games {
		ids = append(ids, g.ID)
	}
	marks := strings.TrimSuffix(strings.Repeat("?, ", len(ids)), ", ")
	rows, err := s.db.QueryContext(ctx, `SELECT game_id, player_id, seat, nickname, role, team, alive, won, token_hash
		FROM game_players WHERE game_id IN (`+marks+`) ORDER BY game_id, seat`, ids...)
	if err != nil {
		return fmt.Errorf("mysql: history players: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var p store.GamePlayer
		var player, role, team string
		if err := rows.Scan(&id, &player, &p.Seat, &p.Nickname, &role, &team, &p.Alive, &p.Won, &p.TokenHash); err != nil {
			return fmt.Errorf("mysql: history players: %w", err)
		}
		p.ID, p.Role, p.Team = engine.PlayerID(player), engine.RoleID(role), teamOf(team)
		g := &games[byID[id]]
		g.Players = append(g.Players, p)
	}
	return rows.Err()
}

// Purge deletes every room created before the given time, with its events
// and games (docs/database.md "Retention"). It returns how many rooms went.
func (s *Store) Purge(before time.Time) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	res, err := s.db.ExecContext(ctx, `DELETE FROM rooms WHERE created_at < ?`, before.UTC())
	if err != nil {
		return 0, fmt.Errorf("mysql: purge: %w", err)
	}
	return res.RowsAffected()
}

func modeOf(name string) engine.Mode {
	if name == modeName(engine.Online) {
		return engine.Online
	}
	return engine.InPerson
}

func teamOf(name string) engine.Team {
	for _, t := range []engine.Team{engine.TeamVillage, engine.TeamWolves, engine.TeamLovers} {
		if teamName(t) == name {
			return t
		}
	}
	return engine.TeamNone
}
