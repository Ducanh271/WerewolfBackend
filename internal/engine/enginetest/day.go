package enginetest

import "masoi/internal/engine"

// Day starts the day after a dawn. Finish it with NobodyHanged or Hanged;
// only then may the next Night begin.
func (g *Game) Day() *Day {
	g.t.Helper()
	switch g.stage {
	case stageDawn:
	case stageNight:
		g.t.Fatalf("Day() called during a night: call Dawn() first")
	case stageDay:
		g.t.Fatalf("Day() called twice: finish the day with NobodyHanged() or Hanged(...)")
	default:
		g.t.Fatalf("Day() called but no night has passed: play Night() and Dawn() first")
	}
	switch g.state.Phase {
	case engine.PhaseDawn: // the narrator has read the announcement
		g.TimeOut()
	case engine.PhaseHunterShot:
		g.t.Fatalf("Day() called while the dead Hunter still chooses: call g.Hunter(...).Shoots(...) or TimeOut() first")
	case engine.PhaseChiefSuccession:
		g.t.Fatalf("Day() called while the dead Chief still names a successor: call NameSuccessor(...) or TimeOut() first")
	}
	g.stage = stageDay
	return &Day{g: g}
}

// Day is one day of discussion and voting.
type Day struct {
	g *Game
}

// NobodyHanged ends the day without an execution.
func (d *Day) NobodyHanged() *Game {
	d.g.t.Helper()
	cmds, err := d.g.driver.NobodyHanged(d.g.state)
	d.g.doAll(cmds, err)
	d.g.stage = stageDayDone
	return d.g
}

// Hanged ends the day with the execution of the named player.
func (d *Day) Hanged(name string) *Game {
	d.g.t.Helper()
	cmds, err := d.g.driver.Hang(d.g.state, d.g.Player(name))
	d.g.doAll(cmds, err)
	d.g.stage = stageDayDone
	return d.g
}
