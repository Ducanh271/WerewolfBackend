// Package mysql is the MySQL implementation of store.EventStore and
// store.GameArchive (docs/database.md).
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	driver "github.com/go-sql-driver/mysql"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// opTimeout bounds every call: the room goroutine waits for it.
const opTimeout = 5 * time.Second

// Store keeps the logs and the game records in MySQL. Each log key has one
// writer, its room goroutine; different rooms may call at the same time.
type Store struct {
	db *sql.DB

	mu    sync.Mutex
	rooms map[string]*roomLog // by log key, filled on first use
}

type roomLog struct {
	id   int64
	next int // seq of the next event
}

var (
	_ store.EventStore  = (*Store)(nil)
	_ store.GameArchive = (*Store)(nil)
)

// Open connects to the database named in dsn (go-sql-driver format, e.g.
// "user:pass@tcp(host:3306)/masoi") and applies the migrations. parseTime is
// forced on.
func Open(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: bad dsn: %w", err)
	}
	cfg.ParseTime = true
	cfg.Loc = time.UTC
	connector, err := driver.NewConnector(cfg)
	if err != nil {
		return nil, fmt.Errorf("mysql: %w", err)
	}
	db := sql.OpenDB(connector)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(4)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: connect: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db, rooms: map[string]*roomLog{}}, nil
}

// Close closes the connections.
func (s *Store) Close() error { return s.db.Close() }

// DB is the connection pool, for tests and tools.
func (s *Store) DB() *sql.DB { return s.db }

// Append adds events to the room's log in one transaction: all or none.
func (s *Store) Append(logKey string, events ...engine.Event) error {
	if len(events) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	r, err := s.room(ctx, logKey)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO room_events (room_id, seq, type, data) VALUES ")
	args := make([]any, 0, 4*len(events))
	for i, e := range events {
		name, data, err := store.EncodeEvent(e)
		if err != nil {
			return err
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?, ?, ?, ?)")
		args = append(args, r.id, r.next+i, name, string(data))
	}
	// One multi-row INSERT is atomic on InnoDB; no transaction needed.
	if _, err := s.db.ExecContext(ctx, sb.String(), args...); err != nil {
		return fmt.Errorf("mysql: append to %s: %w", logKey, err)
	}
	r.next += len(events)
	return nil
}

// Load returns the room's whole log, in order. An unknown room has an empty
// log.
func (s *Store) Load(logKey string) ([]engine.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, `SELECT e.seq, e.type, e.data FROM room_events e
		JOIN rooms r ON r.id = e.room_id WHERE r.log_key = ? ORDER BY e.seq`, logKey)
	if err != nil {
		return nil, fmt.Errorf("mysql: load %s: %w", logKey, err)
	}
	defer rows.Close()
	var out []engine.Event
	for rows.Next() {
		var (
			seq        int
			name, data string
		)
		if err := rows.Scan(&seq, &name, &data); err != nil {
			return nil, fmt.Errorf("mysql: load %s: %w", logKey, err)
		}
		if seq != len(out) {
			return nil, fmt.Errorf("mysql: load %s: event %d missing", logKey, len(out))
		}
		e, err := store.DecodeEvent(name, []byte(data))
		if err != nil {
			return nil, fmt.Errorf("mysql: load %s, event %d: %w", logKey, seq, err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// room finds or creates the rooms row of a log key. The code is the part of
// the key before its first "-" (room.Manager makes keys "K7QP-1a2b3c4d").
func (s *Store) room(ctx context.Context, logKey string) (*roomLog, error) {
	s.mu.Lock()
	r := s.rooms[logKey]
	s.mu.Unlock()
	if r != nil {
		return r, nil
	}
	if logKey == "" {
		return nil, errors.New("mysql: empty log key")
	}
	code, _, _ := strings.Cut(logKey, "-")
	if len(code) > 8 {
		code = code[:8]
	}
	if _, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO rooms (log_key, code) VALUES (?, ?)`, logKey, code); err != nil {
		return nil, fmt.Errorf("mysql: room %s: %w", logKey, err)
	}
	r = &roomLog{}
	if err := s.db.QueryRowContext(ctx, `SELECT r.id, COALESCE(MAX(e.seq) + 1, 0) FROM rooms r
		LEFT JOIN room_events e ON e.room_id = r.id WHERE r.log_key = ? GROUP BY r.id`, logKey).Scan(&r.id, &r.next); err != nil {
		return nil, fmt.Errorf("mysql: room %s: %w", logKey, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if old := s.rooms[logKey]; old != nil {
		return old, nil
	}
	s.rooms[logKey] = r
	return r, nil
}

// EnvDSN is the environment variable that names the database.
const EnvDSN = "MASOI_MYSQL_DSN"

// OpenEnv opens the database named by MASOI_MYSQL_DSN, or returns nil, nil
// when it is not set (the server then keeps everything in memory).
func OpenEnv(ctx context.Context, getenv func(string) string) (*Store, error) {
	dsn := getenv(EnvDSN)
	if dsn == "" {
		return nil, nil
	}
	return Open(ctx, dsn)
}
