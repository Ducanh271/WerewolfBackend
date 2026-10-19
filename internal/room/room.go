package room

import (
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// Errors of the room layer. Engine errors pass through unchanged.
var (
	ErrNotHost      = engine.NewError("not_host", "only the host can do this")
	ErrForbidden    = engine.NewError("forbidden", "command not allowed from a player")
	ErrDisconnected = engine.NewError("disconnected", "this connection is no longer the player's")
	ErrClosed       = engine.NewError("room_closed", "room is closed")
	// ErrInternal answers a command that hit a bug (a panic) inside the room.
	// The command changed nothing and the room goes on; the client is told
	// "internal" and nothing more.
	ErrInternal = errors.New("internal error")
)

// disconnectReason is an error passed to Sink.Evict: it matches
// ErrDisconnected (same error code to the client) but carries a message that
// says why, for the phone being replaced.
type disconnectReason struct{ msg string }

func (e disconnectReason) Error() string        { return e.msg }
func (e disconnectReason) Is(target error) bool { return target == ErrDisconnected }

// Code is ErrDisconnected's wire code.
func (e disconnectReason) Code() string {
	code, _ := engine.ErrorCode(ErrDisconnected)
	return code
}

// Reasons a connection is evicted by a later rejoin (rules §2).
var (
	// ReasonAnotherDevice is for a rejoin by token: the same player's secret
	// code was used from a different connection (two tabs, a reload racing
	// an old tab, a borrowed phone).
	ReasonAnotherDevice error = disconnectReason{"disconnected: joined from another device"}
	// ReasonPINTakenOver is for a rejoin by nickname + PIN: another phone
	// just proved it has the PIN and took the seat.
	ReasonPINTakenOver error = disconnectReason{"someone joined with your PIN"}
	// ReasonKicked is for a player the host removed from the lobby (audit K2).
	ReasonKicked error = disconnectReason{"removed from the lobby by the host"}
)

// Config wires a room. Only Code is required; the rest default to the real
// thing.
type Config struct {
	Code     string
	Mode     engine.Mode
	Clock    Clock
	Random   Randomness
	Identity Identity
	Store    store.EventStore
	Speaker  SpeakerPolicy
	Logger   *slog.Logger // where a recovered panic is reported; default: discard
	// HostKey is the creator's one-time secret (rules §2, audit S8): whoever's
	// first `join` carries it becomes the host. Empty means nobody can claim
	// the host this way (tests that do not care about who is host).
	HostKey string
}

// Room is one game. A single goroutine owns all of its state; everything else
// reaches it through do and post.
type Room struct {
	code  string
	clock Clock
	rnd   Randomness
	ident Identity
	store store.EventStore
	pol   SpeakerPolicy
	logf  *slog.Logger

	inbox chan func()
	done  chan struct{}

	// Owned by the room goroutine.
	state      engine.State
	log        []engine.Event
	nextID     int
	joined     []engine.PlayerID // join order
	host       engine.PlayerID
	hostKey    string          // the creator's unused secret; cleared the moment it grants host
	speaker    engine.PlayerID // who is reading announcements aloud now, "" before the first choice (rules §1)
	conns      map[engine.PlayerID]*connection
	day        dayClock
	armedFor   armKey
	phaseT     Timer
	dayT       Timer
	hostGraceT Timer // pending host-powers transfer, armed while the host is disconnected before the deal (rules §2)
	deadline   time.Time
	closeOnce  bool

	active atomic.Int64 // unix nanoseconds of the last thing a player did
}

// New starts a room. Only InPerson is supported in v1.
func New(cfg Config) (*Room, error) {
	if cfg.Mode != engine.InPerson {
		return nil, fmt.Errorf("%w: %s", engine.ErrModeUnsupported, cfg.Mode)
	}
	if cfg.Clock == nil {
		cfg.Clock = RealClock{}
	}
	if cfg.Random == nil {
		cfg.Random = CryptoRandom{}
	}
	if cfg.Identity == nil {
		cfg.Identity = NewSecretIdentity(cfg.Clock, cfg.Random)
	}
	if cfg.Store == nil {
		cfg.Store = store.NewMemory()
	}
	if cfg.Speaker == nil {
		cfg.Speaker = InPersonSpeaker{}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	r := &Room{
		code: cfg.Code, clock: cfg.Clock, rnd: cfg.Random, ident: cfg.Identity,
		store: cfg.Store, pol: cfg.Speaker, logf: cfg.Logger, hostKey: cfg.HostKey,
		inbox: make(chan func(), 256),
		done:  make(chan struct{}),
		state: engine.NewState(),
		conns: map[engine.PlayerID]*connection{},
	}
	r.active.Store(cfg.Clock.Now().UnixNano())
	go r.loop()
	return r, nil
}

// LastActive is when a player last did anything in the room (joined, sent a
// command, dropped). Timers do not count: the manager uses it to drop rooms
// nobody plays in.
func (r *Room) LastActive() time.Time { return time.Unix(0, r.active.Load()) }

// Done is closed when the room stops.
func (r *Room) Done() <-chan struct{} { return r.done }

// Code is the room code players type in.
func (r *Room) Code() string { return r.code }

func (r *Room) loop() {
	for {
		select {
		case f := <-r.inbox:
			r.run(f)
		case <-r.done:
			return
		}
	}
}

// run does one job of the room goroutine. A job from do reports its own panic
// to its caller; this catches the others (a timer's).
func (r *Room) run(f func()) { _ = r.safe(func() error { f(); return nil }) }

// safe runs f and turns a panic in it (a bug in the engine or a role) into
// ErrInternal, logged with the room code. Only that job is lost: the room, its
// timers and the other rooms go on.
func (r *Room) safe(f func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			r.logf.Error("panic in room", "room", r.code, "panic", fmt.Sprint(p), "stack", string(debug.Stack()))
			err = ErrInternal
		}
	}()
	return f()
}

// do runs f on the room goroutine and waits for its result. If f panics the
// caller gets ErrInternal.
func (r *Room) do(f func() error) error {
	r.active.Store(r.clock.Now().UnixNano())
	res := make(chan error, 1)
	job := func() { res <- r.safe(f) }
	select {
	case r.inbox <- job:
	case <-r.done:
		return ErrClosed
	}
	select {
	case err := <-res:
		return err
	case <-r.done:
		return ErrClosed
	}
}

// post queues f without waiting. Timers use it.
func (r *Room) post(f func()) {
	select {
	case r.inbox <- f:
	case <-r.done:
	}
}

// Close stops the room and its timers.
func (r *Room) Close() {
	_ = r.do(func() error {
		if !r.closeOnce {
			r.closeOnce = true
			r.stopTimers()
			close(r.done)
		}
		return nil
	})
}

// apply decides a command, stores its events, then folds them into the state.
// It does not tell anybody; see changed. The new state is worked out before
// anything is kept, so a panic in Decide or Evolve leaves the room as it was.
func (r *Room) apply(c engine.Command) error {
	r.day.tick(r.clock.Now())
	evs, err := engine.Decide(r.state, c)
	if err != nil || len(evs) == 0 {
		return err
	}
	next := engine.Fold(r.state, evs...)
	if err := r.store.Append(r.code, evs...); err != nil {
		return err
	}
	r.log = append(r.log, evs...)
	r.state = next
	for _, e := range evs {
		if started, ok := e.(engine.DayStarted); ok {
			r.day.reset(started.Duration, r.clock.Now())
		}
		if sound, ok := e.(engine.SoundConfigured); ok {
			r.setSoundPolicy(sound.Mode)
		}
	}
	r.rearm()
	return nil
}

// execute is apply followed by telling everybody.
func (r *Room) execute(c engine.Command) error {
	if err := r.apply(c); err != nil {
		return err
	}
	r.changed()
	return nil
}
