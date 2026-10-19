package enginetest

import "masoi/internal/engine"

// Act lets the night run to the given role's turn and sends the command: the
// verb for any registered role, including one without verbs of its own yet.
//
//	n.Act(Seer, roles.SeerCheck{By: n.Player("Em"), Target: n.Player("An")})
func (n *Night) Act(role engine.RoleID, c engine.Command) *Night {
	n.g.t.Helper()
	n.advanceTo(role)
	return n.Do(c)
}

// ActExpectErr is Act for a command the engine must refuse with want.
func (n *Night) ActExpectErr(role engine.RoleID, c engine.Command, want error) *Night {
	n.g.t.Helper()
	n.advanceTo(role)
	return n.DoExpectErr(c, want)
}
