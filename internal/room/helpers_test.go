package room

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
	"masoi/internal/store"
)

// fixedRandom hands out predictable seeds, tokens and PINs.
type fixedRandom struct {
	mu     sync.Mutex
	seed   int64
	tokens int
	pins   int
	seeds  int // how many seeds were drawn
}

func (f *fixedRandom) Seed() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seeds++
	return f.seed
}

func (f *fixedRandom) Token() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokens++
	return fmt.Sprintf("token-%d", f.tokens)
}

func (f *fixedRandom) PIN() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pins++
	return fmt.Sprintf("%04d", 1000+f.pins)
}

func (f *fixedRandom) setSeed(s int64) {
	f.mu.Lock()
	f.seed = s
	f.mu.Unlock()
}

// recSink records what one phone was sent.
type recSink struct {
	mu      sync.Mutex
	msgs    []Message
	evicted []error
}

func (s *recSink) Deliver(m Message) {
	s.mu.Lock()
	s.msgs = append(s.msgs, m)
	s.mu.Unlock()
}

func (s *recSink) Evict(reason error) {
	s.mu.Lock()
	s.evicted = append(s.evicted, reason)
	s.mu.Unlock()
}

func (s *recSink) last() Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.msgs[len(s.msgs)-1]
}

// events rebuilds the list the phone knows, the way a client would.
func (s *recSink) events() []engine.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []engine.Event
	for _, m := range s.msgs {
		switch {
		case m.Events == nil:
		case m.Events.Reset:
			out = append([]engine.Event(nil), m.Events.Events...)
		default:
			out = append(out[:m.Events.From], m.Events.Events...)
		}
	}
	return out
}

// game is a room with a fake clock and a few joined players.
type game struct {
	t     *testing.T
	r     *Room
	clk   *FakeClock
	rnd   *fixedRandom
	store *store.Memory
	cl    map[engine.PlayerID]*Client
	sink  map[engine.PlayerID]*recSink
	cred  map[engine.PlayerID]Credentials
	nick  map[engine.PlayerID]string
}

var nicknames = []string{"An", "Binh", "Chi", "Dung", "Em", "Phuc", "Giang", "Hoa", "Khoa", "Lan", "Minh", "Nga"}

func newGame(t *testing.T, players int) *game {
	t.Helper()
	g := &game{
		t: t, clk: NewFakeClock(time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)),
		rnd: &fixedRandom{seed: 1}, store: store.NewMemory(),
		cl: map[engine.PlayerID]*Client{}, sink: map[engine.PlayerID]*recSink{},
		cred: map[engine.PlayerID]Credentials{}, nick: map[engine.PlayerID]string{},
	}
	r, err := New(Config{Code: "K7QP", Clock: g.clk, Random: g.rnd, Store: g.store})
	if err != nil {
		t.Fatal(err)
	}
	g.r = r
	t.Cleanup(r.Close)
	for i := 0; i < players; i++ {
		g.join(nicknames[i])
	}
	return g
}

func (g *game) join(nickname string) engine.PlayerID {
	g.t.Helper()
	sink := &recSink{}
	cl, creds, err := g.r.Join(nickname, sink)
	if err != nil {
		g.t.Fatalf("Join(%s): %v", nickname, err)
	}
	g.cl[cl.Player()], g.sink[cl.Player()], g.cred[cl.Player()], g.nick[cl.Player()] = cl, sink, creds, nickname
	return cl.Player()
}

// deal sets up five players as the host: Werewolf, Guard, Seer, 2 Villagers.
func (g *game) deal() {
	g.t.Helper()
	host := g.cl["p1"]
	order := []engine.PlayerID{"p1", "p2", "p3", "p4", "p5"}
	cast := map[engine.RoleID]int{roles.Werewolf: 1, roles.Guard: 1, roles.Seer: 1, roles.Villager: 2}
	g.must(host.Submit(func(engine.PlayerID) engine.Command { return engine.SetSeats{Order: order} }))
	g.must(host.Submit(func(engine.PlayerID) engine.Command { return engine.ConfigureRoles{Roles: cast} }))
	g.must(host.Submit(func(engine.PlayerID) engine.Command { return engine.DealRoles{} }))
}

func (g *game) must(err error) {
	g.t.Helper()
	if err != nil {
		g.t.Fatal(err)
	}
}

// step moves the clock and waits until the room has handled every timer.
func (g *game) step(d time.Duration) {
	g.clk.Advance(d)
	_ = g.r.do(func() error { return nil })
}

func (g *game) snap() engine.State {
	var s engine.State
	_ = g.r.do(func() error { s = g.r.state; return nil })
	return s
}

func (g *game) eventLog() []engine.Event {
	var l []engine.Event
	_ = g.r.do(func() error { l = append(l, g.r.log...); return nil })
	return l
}

func (g *game) who(role engine.RoleID) engine.PlayerID {
	g.t.Helper()
	for id, p := range g.snap().Players {
		if p.Role == role {
			return id
		}
	}
	g.t.Fatalf("nobody has role %s", role)
	return ""
}

// toNight1 plays the role view and the Chief election (nobody votes).
func (g *game) toNight1() {
	g.t.Helper()
	set := engine.DefaultSettings()
	g.step(set.RoleView)
	g.step(set.ChiefElection)
	if s := g.snap(); s.Phase != engine.PhaseNight {
		g.t.Fatalf("want night 1, phase = %s", s.Phase)
	}
}

// toDay plays night 1 out (nobody acts) and the dawn pause.
func (g *game) toDay() {
	g.t.Helper()
	g.toNight1()
	for i := 0; i < 3; i++ {
		g.step(engine.DefaultSettings().NightStep)
	}
	g.step(engine.DefaultSettings().Announce)
	if s := g.snap(); s.Phase != engine.PhaseDay {
		g.t.Fatalf("want day, phase = %s", s.Phase)
	}
}

func as(c engine.Command) PlayerCommand { return func(engine.PlayerID) engine.Command { return c } }

func errorsIs(err, target error) bool { return errors.Is(err, target) }
