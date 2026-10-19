package room

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// The input of the broadcast benchmark (audit R5): one whole game played by
// simple bots straight on the engine, with every living player changing their
// nomination several times each day, the busiest moment for the phones.

const (
	benchSeed             = 20251019
	benchNominationRounds = 4
	benchMaxCommands      = 50000
)

// benchTables are the two tables measured: a common one and the largest one
// rules §2 allows.
var benchTables = []struct {
	players int
	cast    map[engine.RoleID]int
}{
	{13, map[engine.RoleID]int{
		roles.Werewolf: 2, roles.WolfSeer: 1, roles.Seer: 1, roles.Guard: 1, roles.Witch: 1,
		roles.Hunter: 1, roles.Visitor: 1, roles.Cupid: 1, roles.Villager: 4,
	}},
	{30, map[engine.RoleID]int{
		roles.Werewolf: 6, roles.CursedWolf: 1, roles.WolfSeer: 1, roles.Seer: 1, roles.Guard: 1,
		roles.Witch: 1, roles.Hunter: 1, roles.Visitor: 1, roles.Cupid: 1, roles.RustyKnight: 1,
		roles.Villager: 15,
	}},
}

type benchGame struct {
	rng *rand.Rand
	s   engine.State
	log []engine.Event
	// nominations is the log length after each accepted Nominate: the points
	// where a broadcast can be measured.
	nominations []int
	commands    int
}

// benchCut is a point of the game to measure a broadcast at: the log up to
// events.
type benchCut struct {
	name   string
	events int
}

func playBenchGame(tb testing.TB, players int, cast map[engine.RoleID]int) *benchGame {
	tb.Helper()
	g := &benchGame{rng: rand.New(rand.NewPCG(benchSeed, uint64(players))), s: engine.NewState()}
	ids := make([]engine.PlayerID, players)
	for i := range ids {
		ids[i] = engine.PlayerID(fmt.Sprintf("p%d", i+1))
		g.must(tb, engine.JoinRoom{Player: ids[i], Nickname: fmt.Sprintf("P%d", i+1)})
	}
	g.must(tb, engine.SetSeats{Order: ids})
	g.must(tb, engine.ConfigureRoles{Roles: cast})
	g.must(tb, engine.DealRoles{Seed: g.rng.Int64()})
	for g.s.Phase != engine.PhaseEnded {
		if g.commands > benchMaxCommands {
			tb.Fatalf("%d players: stalled in %s, night %d", players, g.s.Phase, g.s.Night)
		}
		g.play()
		if g.s.Phase != engine.PhaseEnded {
			g.must(tb, engine.PhaseTimeout{Phase: g.s.Phase, Seq: g.s.Seq, Seed: g.rng.Int64()})
		}
	}
	return g
}

// cuts are a nomination on the first day, the one nearest the middle of the
// log, and the last one before the game ended.
func (g *benchGame) cuts() []benchCut {
	first, last := g.nominations[0], g.nominations[len(g.nominations)-1]
	middle := first
	for _, n := range g.nominations {
		if n <= len(g.log)/2 {
			middle = n
		}
	}
	return []benchCut{{"start", first}, {"middle", middle}, {"end", last}}
}

func (g *benchGame) try(c engine.Command) bool {
	evs, err := engine.Decide(g.s, c)
	if err != nil {
		return false // the bots are not careful: the engine refuses some moves
	}
	g.commands++
	g.log = append(g.log, evs...)
	g.s = engine.Fold(g.s, evs...)
	if _, ok := c.(engine.Nominate); ok {
		g.nominations = append(g.nominations, len(g.log))
	}
	return true
}

func (g *benchGame) must(tb testing.TB, c engine.Command) {
	tb.Helper()
	evs, err := engine.Decide(g.s, c)
	if err != nil {
		tb.Fatalf("%T: %v", c, err)
	}
	g.commands++
	g.log = append(g.log, evs...)
	g.s = engine.Fold(g.s, evs...)
}

// play lets the bots act once in the current phase; the caller times it out.
func (g *benchGame) play() {
	switch g.s.Phase {
	case engine.PhaseChiefElection:
		for _, id := range g.living() {
			g.try(engine.VoteChief{Voter: id, Candidate: g.other(id)})
		}
	case engine.PhaseNight:
		g.night()
	case engine.PhaseDay, engine.PhaseNominationLock:
		for round := 0; round < benchNominationRounds; round++ {
			for _, id := range g.living() {
				g.try(engine.Nominate{Voter: id, Target: g.other(id)})
			}
		}
	case engine.PhaseVerdict:
		for _, id := range g.living() {
			g.try(engine.CastVerdict{Voter: id, Kill: g.rng.IntN(10) < 7})
		}
	case engine.PhaseHunterShot:
		g.try(engine.HunterShoot{By: g.s.Shot.Hunter, Target: g.other(g.s.Shot.Hunter)})
	case engine.PhaseChiefSuccession:
		g.try(engine.NameSuccessor{By: g.s.Succession.Former, Successor: g.other(g.s.Succession.Former)})
	}
}

func (g *benchGame) night() {
	step, ok := engine.CurrentNightRole(g.s)
	if !ok {
		return
	}
	if step == roles.Werewolf { // the pack agrees on one living non-wolf
		var prey []engine.PlayerID
		for _, id := range g.living() {
			if g.s.Players[id].Team != engine.TeamWolves {
				prey = append(prey, id)
			}
		}
		if len(prey) == 0 {
			return
		}
		target := prey[g.rng.IntN(len(prey))]
		for _, id := range g.living() {
			if g.s.Players[id].Team == engine.TeamWolves {
				g.try(roles.WolfVote{By: id, Target: target})
			}
		}
		return
	}
	for _, id := range g.living() {
		if g.s.Players[id].Role != step {
			continue
		}
		for _, c := range g.nightMoves(step, id) {
			g.try(c)
		}
	}
}

func (g *benchGame) nightMoves(role engine.RoleID, by engine.PlayerID) []engine.Command {
	switch role {
	case roles.Guard:
		return []engine.Command{roles.GuardProtect{By: by, Target: g.other("")}}
	case roles.Seer:
		return []engine.Command{roles.SeerCheck{By: by, Target: g.other(by)}}
	case roles.WolfSeer:
		return []engine.Command{roles.WolfSeerCheck{By: by, Target: g.other(by)}}
	case roles.CursedWolf:
		return []engine.Command{roles.CursedWolfCurse{By: by, Target: g.other(by)}}
	case roles.Visitor:
		return []engine.Command{roles.VisitorVisit{By: by, Target: g.other(by)}}
	case roles.Cupid:
		a := g.other("")
		return []engine.Command{roles.CupidPair{By: by, A: a, B: g.other(a)}}
	case roles.Witch:
		return []engine.Command{roles.WitchHeal{By: by}, roles.WitchPoison{By: by, Target: g.other(by)}}
	}
	return nil
}

func (g *benchGame) living() []engine.PlayerID {
	var out []engine.PlayerID
	for _, id := range g.s.Seats {
		if g.s.Players[id].Alive {
			out = append(out, id)
		}
	}
	return out
}

// other is a random living player other than except ("" for anybody).
func (g *benchGame) other(except engine.PlayerID) engine.PlayerID {
	var pool []engine.PlayerID
	for _, id := range g.living() {
		if id != except {
			pool = append(pool, id)
		}
	}
	if len(pool) == 0 {
		return ""
	}
	return pool[g.rng.IntN(len(pool))]
}

// TestBenchGame_PlaysToTheEnd keeps the benchmark's input honest: both games
// end, and the measured points come in order with day nominations in between.
func TestBenchGame_PlaysToTheEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("plays two whole games; skipped with -short (make test-fast)")
	}
	for _, table := range benchTables {
		g := playBenchGame(t, table.players, table.cast)
		cuts := g.cuts()
		if !(cuts[0].events < cuts[1].events && cuts[1].events < cuts[2].events) {
			t.Errorf("%d players: cuts out of order: %+v", table.players, cuts)
		}
		t.Logf("%d players: %d events, %d commands, %d nominations, nights %d, cuts %+v",
			table.players, len(g.log), g.commands, len(g.nominations), g.s.Night, cuts)
	}
}
