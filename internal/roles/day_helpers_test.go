package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
)

// All ten players, in seating order. Phuc is the Chief in chiefDay.
var all = []string{"An", "Binh", "Chi", "Dung", "Em", "Phuc", "Giang", "Hoa", "Khoa", "Lan"}

// chiefTable is the usual table with a known Chief; the election always makes
// somebody Chief, so a scenario that does not care picks a bystander.
func chiefTable(t *testing.T, chief string) *Game {
	return NewGame(t, append(tenPlayers(), Chief(chief))...)
}

// day starts day 1 after a quiet night. Lan is the Chief, and her double vote
// only matters in the scenarios that put her in: they say so, and the others
// leave her out of the count.
func day(t *testing.T) (*Game, *Day) {
	g := chiefTable(t, "Lan")
	g.Night()
	g.Dawn().ExpectNobodyDied()
	return g, g.Day()
}

// chiefDay is day 1 with Phuc as Chief.
func chiefDay(t *testing.T) (*Game, *Day) {
	g := chiefTable(t, "Phuc")
	g.Night()
	g.Dawn().ExpectNobodyDied()
	return g, g.Day()
}

// nobodyExcept has everyone but the named players choose "no one".
func nobodyExcept(d *Day, except ...string) {
	for _, name := range all {
		skip := false
		for _, e := range except {
			skip = skip || e == name
		}
		if !skip {
			d.NominateNobody(name)
		}
	}
}

// toVerdict puts target on trial and runs the defense out.
func toVerdict(d *Day, target string) {
	for _, name := range all {
		if name == target {
			d.NominateNobody(name)
		} else {
			d.Nominate(name, target)
		}
	}
	d.ExpectPhase(engine.PhaseNominationLock).TimeOut() // nobody changes their mind
	d.ExpectOnTrial(target).TimeOut().ExpectPhase(engine.PhaseVerdict)
}
