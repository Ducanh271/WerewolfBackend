package room

import (
	"crypto/rand"
	"errors"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	"masoi/internal/engine"
)

// Errors of the manager.
var (
	ErrRoomNotFound = errors.New("room not found")
	ErrTooManyRooms = errors.New("too many rooms")
)

// CodeAlphabet has no O, 0, I, 1 or L: those get misread when a code is typed
// or read out across the table.
const CodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// CodeLength is the length of a room code.
const CodeLength = 4

const (
	// IdleTimeout is how long a room may go without a player doing anything
	// before it is deleted.
	IdleTimeout = 2 * time.Hour
	sweepEvery  = time.Minute
	defaultMax  = 1000
)

// ManagerConfig wires a Manager. The zero value is the real thing.
type ManagerConfig struct {
	Clock    Clock
	Random   Randomness   // handed to every room; nil means crypto/rand
	Logger   *slog.Logger // handed to every room; nil means discard
	MaxRooms int          // 0 means 1000
	// NewCode draws a room code; nil means CodeLength letters of CodeAlphabet
	// from crypto/rand. Tests use it to force collisions.
	NewCode func() string
}

// Manager creates rooms, finds them by code and deletes the idle ones.
type Manager struct {
	cfg ManagerConfig

	mu     sync.Mutex
	rooms  map[string]*Room
	sweepT Timer
	closed bool
}

// NewManager starts a manager and its cleanup timer.
func NewManager(cfg ManagerConfig) *Manager {
	if cfg.Clock == nil {
		cfg.Clock = RealClock{}
	}
	if cfg.MaxRooms <= 0 {
		cfg.MaxRooms = defaultMax
	}
	if cfg.NewCode == nil {
		cfg.NewCode = randomCode
	}
	m := &Manager{cfg: cfg, rooms: map[string]*Room{}}
	m.sweepT = cfg.Clock.AfterFunc(sweepEvery, m.tick)
	return m
}

// Create opens a room in the given mode and returns it. Online rooms are
// refused (engine.ErrModeUnsupported).
func (m *Manager) Create(mode engine.Mode) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrClosed
	}
	if len(m.rooms) >= m.cfg.MaxRooms {
		return nil, ErrTooManyRooms
	}
	code := m.cfg.NewCode()
	for tries := 0; m.rooms[code] != nil; tries++ {
		if tries > 100 {
			return nil, ErrTooManyRooms
		}
		code = m.cfg.NewCode()
	}
	r, err := New(Config{Code: code, Mode: mode, Clock: m.cfg.Clock, Random: m.cfg.Random, Logger: m.cfg.Logger})
	if err != nil {
		return nil, err
	}
	m.rooms[code] = r
	return r, nil
}

// Get finds a room by its code. Codes are typed by hand, so case and
// surrounding spaces do not matter.
func (m *Manager) Get(code string) (*Room, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	m.mu.Lock()
	defer m.mu.Unlock()
	if r := m.rooms[code]; r != nil {
		return r, nil
	}
	return nil, ErrRoomNotFound
}

// Len is the number of open rooms.
func (m *Manager) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rooms)
}

// Close stops the cleanup timer and every room.
func (m *Manager) Close() {
	m.mu.Lock()
	m.closed = true
	m.sweepT.Stop()
	rooms := m.rooms
	m.rooms = map[string]*Room{}
	m.mu.Unlock()
	for _, r := range rooms {
		r.Close()
	}
}

func (m *Manager) tick() {
	m.sweep()
	m.mu.Lock()
	if !m.closed {
		m.sweepT = m.cfg.Clock.AfterFunc(sweepEvery, m.tick)
	}
	m.mu.Unlock()
}

// sweep deletes every room that has been idle for IdleTimeout.
func (m *Manager) sweep() {
	now := m.cfg.Clock.Now()
	var idle []*Room
	m.mu.Lock()
	for code, r := range m.rooms {
		if now.Sub(r.LastActive()) >= IdleTimeout {
			delete(m.rooms, code)
			idle = append(idle, r)
		}
	}
	m.mu.Unlock()
	for _, r := range idle {
		r.Close() // outside the lock: it waits for the room goroutine
	}
}

func randomCode() string {
	n := big.NewInt(int64(len(CodeAlphabet)))
	b := make([]byte, CodeLength)
	for i := range b {
		v, err := rand.Int(rand.Reader, n)
		if err != nil {
			panic("crypto/rand failed: " + err.Error()) // no safe fallback for a secret-ish value
		}
		b[i] = CodeAlphabet[v.Int64()]
	}
	return string(b)
}
