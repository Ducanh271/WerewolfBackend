package enginetest

import "masoi/internal/engine"

// Game returns the game the day belongs to.
func (d *Day) Game() *Game { return d.g }

// Nominate is the voter's public nomination of target for the trial stand.
func (d *Day) Nominate(voter, target string) *Day {
	d.g.t.Helper()
	d.g.Do(engine.Nominate{Voter: d.g.Player(voter), Target: d.g.Player(target)})
	return d
}

// NominateNobody is the voter choosing "no one".
func (d *Day) NominateNobody(voter string) *Day {
	d.g.t.Helper()
	d.g.Do(engine.Nominate{Voter: d.g.Player(voter)})
	return d
}

// NominateExpectErr sends a nomination that must be rejected with want.
func (d *Day) NominateExpectErr(voter, target string, want error) *Day {
	d.g.t.Helper()
	d.g.DoExpectErr(engine.Nominate{Voter: engine.PlayerID(voter), Target: engine.PlayerID(target)}, want)
	return d
}

// Skip presses the voter's skip button.
func (d *Day) Skip(voter string) *Day {
	d.g.t.Helper()
	d.g.Do(engine.SkipDay{Voter: d.g.Player(voter), On: true})
	return d
}

// Unskip releases the voter's skip button.
func (d *Day) Unskip(voter string) *Day {
	d.g.t.Helper()
	d.g.Do(engine.SkipDay{Voter: d.g.Player(voter)})
	return d
}

// Kill is the voter's public vote to hang the player on trial.
func (d *Day) Kill(voter string) *Day {
	d.g.t.Helper()
	d.g.Do(engine.CastVerdict{Voter: d.g.Player(voter), Kill: true})
	return d
}

// Spare is the voter's public vote to let the player on trial go.
func (d *Day) Spare(voter string) *Day {
	d.g.t.Helper()
	d.g.Do(engine.CastVerdict{Voter: d.g.Player(voter)})
	return d
}

// VerdictExpectErr sends a kill or spare vote that must be rejected with want.
func (d *Day) VerdictExpectErr(voter string, kill bool, want error) *Day {
	d.g.t.Helper()
	d.g.DoExpectErr(engine.CastVerdict{Voter: engine.PlayerID(voter), Kill: kill}, want)
	return d
}

// TimeOut lets the running timer (day clock, defense or verdict) run out.
func (d *Day) TimeOut() *Day {
	d.g.t.Helper()
	d.g.TimeOut()
	return d
}

// ExpectPhase checks the engine phase.
func (d *Day) ExpectPhase(phase engine.Phase) *Day {
	d.g.t.Helper()
	d.g.ExpectPhase(phase)
	return d
}

// ExpectOnTrial checks who stands trial.
func (d *Day) ExpectOnTrial(name string) *Day {
	d.g.t.Helper()
	if got := d.g.state.Day.Defendant; got != d.g.Player(name) {
		d.g.t.Fatalf("expected %s on trial, but the player on trial is %q", name, got)
	}
	return d
}

// ExpectAnnounced checks that the narrator announced the message with the
// given i18n key at some point in the game.
func (d *Day) ExpectAnnounced(key string) *Day {
	d.g.t.Helper()
	if !contains(announcementKeys(d.g.events), key) {
		d.g.t.Fatalf("expected announcement %q, but the game announced: %v", key, announcementKeys(d.g.events))
	}
	return d
}
