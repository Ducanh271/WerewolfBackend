package room

import (
	"errors"
	"fmt"
	"time"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// Errors of the room layer. Engine errors pass through unchanged.
var (
	ErrNotHost      = errors.New("only the host can do this")
	ErrForbidden    = errors.New("command not allowed from a player")
	ErrDisconnected = errors.New("this connection is no longer the player's")
	ErrClosed       = errors.New("room is closed")
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

	inbox chan func()
	done  chan struct{}

	// Owned by the room goroutine.
	state     engine.State
	log       []engine.Event
	nextID    int
	joined    []engine.PlayerID // join order
	host      engine.PlayerID
	conns     map[engine.PlayerID]*connection
	day       dayClock
	armedFor  armKey
	phaseT    Timer
	dayT      Timer
	deadline  time.Time
	closeOnce bool
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
	r := &Room{
		code: cfg.Code, clock: cfg.Clock, rnd: cfg.Random, ident: cfg.Identity,
		store: cfg.Store, pol: cfg.Speaker,
		inbox: make(chan func(), 256),
		done:  make(chan struct{}),
		state: engine.NewState(),
		conns: map[engine.PlayerID]*connection{},
	}
	go r.loop()
	return r, nil
}

// Code is the room code players type in.
func (r *Room) Code() string { return r.code }

func (r *Room) loop() {
	for {
		select {
		case f := <-r.inbox:
			f()
		case <-r.done:
			return
		}
	}
}

// do runs f on the room goroutine and waits for its result.
func (r *Room) do(f func() error) error {
	res := make(chan error, 1)
	select {
	case r.inbox <- func() { res <- f() }:
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
// It does not tell anybody; see changed.
func (r *Room) apply(c engine.Command) error {
	r.day.tick(r.clock.Now())
	evs, err := engine.Decide(r.state, c)
	if err != nil || len(evs) == 0 {
		return err
	}
	if err := r.store.Append(r.code, evs...); err != nil {
		return err
	}
	r.log = append(r.log, evs...)
	r.state = engine.Fold(r.state, evs...)
	for _, e := range evs {
		if started, ok := e.(engine.DayStarted); ok {
			r.day.reset(started.Duration, r.clock.Now())
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
