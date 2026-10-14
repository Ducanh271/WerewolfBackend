package enginetest

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"masoi/internal/engine"
)

// stage tracks where the scenario is, so impossible sequences fail loudly.
type stage int

const (
	stageLobby   stage = iota // no roles yet
	stageSetup                // table ready, no night played
	stageNight                // inside Night()
	stageDawn                 // Dawn() done, day not started
	stageDay                  // inside Day(), no outcome yet
	stageDayDone              // day over, next night allowed
)

// Game is one scenario: an engine state plus the events that built it.
type Game struct {
	t      testing.TB
	state  engine.State
	events []engine.Event
	driver Driver
	stage  stage
}

// NewGame builds a table: players join, sit in the given order and, with Cast
// or Roles, receive their roles. It fails the test if the table breaks rules §2.
func NewGame(t testing.TB, opts ...Option) *Game {
	t.Helper()
	cfg := &setup{driver: realDriver{}}
	for _, opt := range opts {
		opt(cfg)
	}
	g := &Game{t: t, state: engine.NewState(), driver: cfg.driver}
	if len(cfg.errs) > 0 {
		t.Fatalf("bad scenario setup: %s", cfg.errs[0])
	}
	if len(cfg.seats) == 0 {
		t.Fatalf("bad scenario setup: no Seats given")
	}
	if cfg.cast != nil && cfg.roles != nil {
		t.Fatalf("bad scenario setup: use either Cast or Roles, not both")
	}

	order := make([]engine.PlayerID, len(cfg.seats))
	for i, name := range cfg.seats {
		order[i] = engine.PlayerID(name)
		g.Do(engine.JoinRoom{Player: order[i], Nickname: name})
	}
	g.Do(engine.SetSeats{Order: order})

	switch {
	case cfg.cast != nil:
		g.castRoles(cfg.cast)
		g.stage = stageSetup
	case cfg.roles != nil:
		g.Do(engine.ConfigureRoles{Roles: cfg.roles})
		g.Do(engine.DealRoles{Seed: cfg.seed})
		g.stage = stageSetup
	}
	return g
}

// castRoles validates the cast with ConfigureRoles, then deals it as given.
func (g *Game) castRoles(cast map[string]engine.RoleID) {
	g.t.Helper()
	counts := map[engine.RoleID]int{}
	assignments := map[engine.PlayerID]engine.RoleID{}
	for name, role := range cast {
		if _, seated := g.state.Players[engine.PlayerID(name)]; !seated {
			g.t.Fatalf("bad scenario setup: Cast names %s, who is not in Seats", name)
		}
		counts[role]++
		assignments[engine.PlayerID(name)] = role
	}
	for _, id := range g.state.Seats {
		if _, ok := assignments[id]; !ok {
			g.t.Fatalf("bad scenario setup: %s is seated but has no role in Cast", id)
		}
	}
	if _, err := engine.Decide(g.state, engine.ConfigureRoles{Roles: counts}); err != nil {
		g.t.Fatalf("this table breaks the setup rules (rules §2): %v", err)
	}
	g.Do(engine.ConfigureRoles{Roles: counts})
	g.record(engine.RolesDealt{Assignments: assignments})
}

// record folds an event that no command produces (the cast deal).
func (g *Game) record(evs ...engine.Event) {
	g.state = engine.Fold(g.state, evs...)
	g.events = append(g.events, evs...)
}

// Do sends a command and fails the test if the engine rejects it.
func (g *Game) Do(c engine.Command) {
	g.t.Helper()
	evs, err := engine.Decide(g.state, c)
	if err != nil {
		g.t.Fatalf("%T was rejected: %v", c, err)
	}
	g.record(evs...)
}

// DoExpectErr sends a command that must be rejected with want (matched with
// errors.Is). A rejected command changes nothing.
func (g *Game) DoExpectErr(c engine.Command, want error) {
	g.t.Helper()
	evs, err := engine.Decide(g.state, c)
	if err == nil {
		g.t.Fatalf("%T was accepted (%d events), want error %v", c, len(evs), want)
	}
	if !errors.Is(err, want) {
		g.t.Fatalf("%T was rejected with %v, want %v", c, err, want)
	}
}

// doAll sends the commands of a driver step.
func (g *Game) doAll(cmds []engine.Command, err error) {
	g.t.Helper()
	if err != nil {
		g.t.Fatalf("cannot advance the game: %v", err)
	}
	for _, c := range cmds {
		g.Do(c)
	}
}

// State returns the current engine state.
func (g *Game) State() engine.State { return g.state }

// Events returns every event so far, in order.
func (g *Game) Events() []engine.Event { return append([]engine.Event(nil), g.events...) }

// Player returns the ID of a seated player and fails on an unknown name.
func (g *Game) Player(name string) engine.PlayerID {
	g.t.Helper()
	id := engine.PlayerID(name)
	if _, ok := g.state.Players[id]; !ok {
		g.t.Fatalf("no player named %s at this table", name)
	}
	return id
}

// deadSet lists the dead players' names, sorted.
func (g *Game) deadSet() []string {
	var dead []string
	for id, p := range g.state.Players {
		if !p.Alive {
			dead = append(dead, string(id))
		}
	}
	sort.Strings(dead)
	return dead
}

func (g *Game) String() string {
	return fmt.Sprintf("game(%d players, %s)", len(g.state.Players), g.state.Phase)
}
