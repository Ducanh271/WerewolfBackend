package store

import (
	"sync"

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
	m.games = append(m.games, rec)
	return nil
}

// Games returns the records saved so far, oldest first.
func (m *Memory) Games() []GameRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]GameRecord(nil), m.games...)
}
