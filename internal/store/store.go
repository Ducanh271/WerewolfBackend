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
	// Load returns a copy of the room's log.
	Load(room string) ([]engine.Event, error)
}

// Memory is the v1 store: everything lives in the process.
type Memory struct {
	mu   sync.Mutex
	logs map[string][]engine.Event
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
