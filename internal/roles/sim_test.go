package roles_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"

	"masoi/internal/engine"
)

// The "many games" test: bots play whole games on random tables, driven by
// timeouts only (the fake clock is the PhaseTimeout command, so a game takes
// milliseconds). A failure prints the seed; replay one game with
//
//	SIM_SEED=<seed> go test ./internal/roles -run TestSim -v
//
// and change the number of games with SIM_GAMES.
const (
	simGamesDefault = 300
	simBaseSeed     = 20250
	simMaxCommands  = 20000
	simMaxNights    = 80
)

type simGame struct {
	seed  uint64
	rng   *rand.Rand
	s     engine.State
	log   []engine.Event
	desc  string
	count int // commands applied
	// rejected counts commands the engine refused with a typed error; the bots
	// try some wrong moves on purpose.
	rejected int
}

func newSimGame(seed uint64) *simGame {
	return &simGame{seed: seed, rng: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)), s: engine.NewState()}
}

// do applies one command. A typed engine error is returned to the caller; an
// untyped one (an internal error) is a bug and panics the game's goroutine
// into the test's report.
func (g *simGame) do(c engine.Command) error {
	evs, err := engine.Decide(g.s, c)
	if err != nil {
		if !isTypedError(err) {
			panic(fmt.Sprintf("internal (untyped) error from %T: %v", c, err))
		}
		g.rejected++
		return err
	}
	g.count++
	g.log = append(g.log, evs...)
	g.s = engine.Fold(g.s, evs...)
	return nil
}

func TestSim_ManyGames_FinishAndNeverLeakSecrets(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: many simulated games; skipped with -short (make test-fast)")
	}
	games, first := simGamesDefault, uint64(simBaseSeed)
	if v := os.Getenv("SIM_GAMES"); v != "" {
		games, _ = strconv.Atoi(v)
	}
	if v := os.Getenv("SIM_SEED"); v != "" {
		n, _ := strconv.ParseUint(v, 10, 64)
		first, games = n, 1
	}
	outcomes := map[string]int{}
	used := map[string]int{}
	for i := 0; i < games; i++ {
		seed := first + uint64(i)
		out, err := playSimGame(seed, used)
		if err != nil {
			t.Fatalf("game with seed %d failed: %v\nreplay: SIM_SEED=%d go test ./internal/roles -run TestSim -v", seed, err, seed)
		}
		outcomes[out]++
	}
	t.Logf("%d games played; outcomes %v; events seen %v", games, outcomes, used)
}

// playSimGame plays one game to its end and returns how it ended.
func playSimGame(seed uint64, used map[string]int) (outcome string, err error) {
	g := newSimGame(seed)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic (%s): %v", g.desc, r)
		}
	}()
	if err := g.setup(); err != nil {
		return "", fmt.Errorf("%s: %w", g.desc, err)
	}
	for g.s.Phase != engine.PhaseEnded {
		if g.count > simMaxCommands || g.s.Night > simMaxNights {
			return "", fmt.Errorf("stalled after %d commands, night %d, phase %v (%s)", g.count, g.s.Night, g.s.Phase, g.desc)
		}
		phase, seq := g.s.Phase, g.s.Seq
		g.playStep()
		if err := g.timeout(); err != nil {
			return "", fmt.Errorf("%s: %w", g.desc, err)
		}
		if g.s.Phase == phase && g.s.Seq == seq && g.s.Phase != engine.PhaseEnded {
			return "", fmt.Errorf("stalled: a timeout changed nothing in phase %v (%s)", phase, g.desc)
		}
		g.checkViews()
		if err := g.checkLovers(); err != nil {
			return "", err
		}
		if phase == engine.PhaseNight && g.s.Phase == engine.PhaseDawn {
			if g.s.Night%3 == 1 { // replaying the log per viewer is costly
				g.checkEvents()
			} else if living := g.living(); len(living) > 0 { // one viewer each other night, in turn
				g.checkEventsOf(living[g.s.Night%len(living)])
			}
		}
	}
	for _, e := range g.log {
		used[fmt.Sprintf("%T", e)]++
	}
	return g.finish()
}

// finish checks how the game ended.
func (g *simGame) finish() (string, error) {
	for i := len(g.log) - 1; i >= 0; i-- {
		if ended, ok := g.log[i].(engine.GameEnded); ok {
			if ended.Outcome.Draw {
				return "draw", nil
			}
			if g.s.Winner == nil {
				return "", fmt.Errorf("GameEnded names %v but the state has no winner (%s)", ended.Outcome.Winner, g.desc)
			}
			return g.s.Winner.String(), nil
		}
	}
	return "", fmt.Errorf("the game is in PhaseEnded without a GameEnded event (%s)", g.desc)
}

// timeout runs the current phase's timer out, with a fresh seed as the room
// does.
func (g *simGame) timeout() error {
	if g.s.Phase == engine.PhaseEnded {
		return nil
	}
	return g.do(engine.PhaseTimeout{Phase: g.s.Phase, Seq: g.s.Seq, Seed: g.rng.Int64()})
}
