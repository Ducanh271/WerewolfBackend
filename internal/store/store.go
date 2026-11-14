package store

import (
	"sort"
	"sync"
	"time"

	"masoi/internal/engine"
)

// EventStore keeps the events of every room, in order. The room writes
// every event here before it tells anybody about it.
type EventStore interface {
	// Append adds events to the room's log, all or none.
	Append(room string, events ...engine.Event) error
	// Load returns a copy of the room's log. The server does not call it yet
	// (restoring rooms after a restart is T33c, docs/database.md); tests read
	// what the room wrote with it.
	Load(room string) ([]engine.Event, error)
}

// Memory is the store without a database: everything lives in the process.
// It is also a GameArchive, which tests read with Games.
type Memory struct {
	mu    sync.Mutex
	logs  map[string][]engine.Event
	games []GameRecord
}

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{logs: map[string][]engine.Event{}}
}

// Append implements EventStore.
func (m *Memory) Append(room string, events ...engine.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs[room] = append(m.logs[room], events...)
	return nil
}

// Load implements EventStore.
func (m *Memory) Load(room string) ([]engine.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]engine.Event(nil), m.logs[room]...), nil
}

// SaveGame keeps the record of a finished game.
func (m *Memory) SaveGame(rec GameRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.games = append(m.games, cloneRecord(rec))
	return nil
}

// ReplaceToken updates the player's token hash in the room's saved games.
func (m *Memory) ReplaceToken(logKey string, player engine.PlayerID, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.games {
		if m.games[i].LogKey != logKey {
			continue
		}
		for j := range m.games[i].Players {
			if m.games[i].Players[j].ID == player {
				m.games[i].Players[j].TokenHash = tokenHash
			}
		}
	}
	return nil
}

// GamesFor is History over the saved records; ids are positions from 1.
func (m *Memory) GamesFor(tokenHashes []string, since time.Time, limit int) ([]SavedGame, error) {
	want := map[string]bool{}
	for _, h := range tokenHashes {
		if h != "" {
			want[h] = true
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []SavedGame
	for i, g := range m.games {
		if g.Ended.Before(since) {
			continue
		}
		for _, p := range g.Players {
			if want[p.TokenHash] {
				out = append(out, SavedGame{ID: int64(i + 1), GameRecord: cloneRecord(g)})
				break
			}
		}
	}
	// Newest first, as the MySQL store answers.
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].Ended.Equal(out[j].Ended) {
			return out[i].Ended.After(out[j].Ended)
		}
		return out[i].ID > out[j].ID
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func cloneRecord(g GameRecord) GameRecord {
	g.Players = append([]GamePlayer(nil), g.Players...)
	return g
}

// Games returns the records saved so far, oldest first.
func (m *Memory) Games() []GameRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]GameRecord, len(m.games))
	for i, g := range m.games {
		out[i] = cloneRecord(g)
	}
	return out
}
