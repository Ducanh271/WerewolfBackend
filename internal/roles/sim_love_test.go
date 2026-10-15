package roles_test

import (
	"fmt"

	"masoi/internal/engine"
)

// checkLovers is the invariant of rules §7 for the many-games test: once a
// chain of deaths is over, a couple has either both alive or both dead, or the
// survivor is marked to die in the next night (and that mark only exists by
// day). It also checks that a lover's screen shows the pair only after the
// phone check told them.
func (g *simGame) checkLovers() error {
	if g.s.Lovers == nil || g.s.Phase == engine.PhaseHunterShot || g.s.Phase == engine.PhaseEnded {
		return nil
	}
	a, b := g.s.Players[g.s.Lovers.A], g.s.Players[g.s.Lovers.B]
	if a.Alive != b.Alive {
		survivor := a
		if !a.Alive {
			survivor = b
		}
		if survivor.Flags[engine.FlagGrieving] == 0 {
			return fmt.Errorf("%s lives on though their lover died, and is not marked to die (night %d, phase %v, %s)",
				survivor.ID, g.s.Night, g.s.Phase, g.desc)
		}
	}
	for _, p := range []*engine.Player{a, b} {
		v := engine.ViewFor(g.s, engine.PlayerViewer(p.ID))
		if p.Alive && v.Lover != nil && p.Flags[engine.FlagLoveKnown] == 0 {
			return fmt.Errorf("%s sees the pair before the phone check (%s)", p.ID, g.desc)
		}
	}
	return nil
}
