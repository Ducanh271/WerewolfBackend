package enginetest

import "masoi/internal/engine"

// Election opens the Chief election: the role-view countdown runs out and the
// public vote begins. Cast the votes with Vote, then end it with Close. A
// scenario that never calls Election gets one where nobody votes (or, with the
// Chief option, everybody votes for that player).
func (g *Game) Election() *Election {
	g.t.Helper()
	if g.stage != stageSetup {
		g.t.Fatalf("Election() must come right after the table setup, before any night")
	}
	g.Do(engine.PhaseTimeout{Phase: g.state.Phase, Seq: g.state.Seq})
	g.stage = stageElection
	return &Election{g: g}
}

// Election is the Chief vote in progress.
type Election struct {
	g *Game
}

// Vote casts the voter's public vote; a later vote replaces the earlier one.
func (e *Election) Vote(voter, candidate string) *Election {
	e.g.t.Helper()
	e.g.Do(engine.VoteChief{Voter: e.g.Player(voter), Candidate: e.g.Player(candidate)})
	return e
}

// VoteExpectErr casts a vote that must be rejected with want.
func (e *Election) VoteExpectErr(voter, candidate string, want error) *Election {
	e.g.t.Helper()
	e.g.DoExpectErr(engine.VoteChief{Voter: engine.PlayerID(voter), Candidate: engine.PlayerID(candidate)}, want)
	return e
}

// Close runs the minute out. seed settles a tie or an election nobody voted
// in; the room layer would generate it. Night 1 starts straight after.
func (e *Election) Close(seed int64) *Game {
	e.g.t.Helper()
	e.g.Do(engine.PhaseTimeout{Phase: e.g.state.Phase, Seq: e.g.state.Seq, Seed: seed})
	e.g.stage = stageElected
	return e.g
}

// TimeOut lets the running phase's time run out.
func (g *Game) TimeOut() *Game {
	g.t.Helper()
	g.Do(engine.PhaseTimeout{Phase: g.state.Phase, Seq: g.state.Seq})
	return g
}

// NameSuccessor is the dead Chief choosing the new Chief.
func (g *Game) NameSuccessor(chief, successor string) *Game {
	g.t.Helper()
	g.Do(engine.NameSuccessor{By: g.Player(chief), Successor: g.Player(successor)})
	return g
}

// ExpectChief checks who holds the title.
func (g *Game) ExpectChief(name string) *Game {
	g.t.Helper()
	if got := g.state.Chief; got != g.Player(name) {
		g.t.Fatalf("expected %s to be Chief, but the Chief is %q", name, got)
	}
	return g
}

// ExpectNoChief checks that nobody holds the title.
func (g *Game) ExpectNoChief() *Game {
	g.t.Helper()
	if got := g.state.Chief; got != "" {
		g.t.Fatalf("expected no Chief, but the Chief is %s", got)
	}
	return g
}
