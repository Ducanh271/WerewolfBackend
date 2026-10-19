package roles_test

import (
	"testing"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/roles"
)

// The generic verb works like a role's own: the Seer's check through Act gives
// the same result as Seer("Em").Checks("An"), and a second check is refused.
func TestNight_ActIsTheVerbOfAnyRegisteredRole(t *testing.T) {
	g := newTable(t)
	n := g.Night()
	check := roles.SeerCheck{By: n.Player("Em"), Target: n.Player("An")}
	n.Act(Seer, check).ExpectSeerSaw("Em", "An", true)
	n.ActExpectErr(Seer, check, engine.ErrAlreadyActed)
}
