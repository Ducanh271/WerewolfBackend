package enginetest

import "masoi/internal/engine"

// Night starts the next night and returns the builder for what the roles do in
// it. A night may only follow the table setup or a finished day: after a night
// you must call Dawn, and then Day with NobodyHanged or Hanged, before the next
// Night. Calling Night twice in a row fails the test.
//
// Only the roles you mention take their turn. Every other role is treated as
// timed out and does nothing that night (rules §3: "Hết giờ mà chưa chọn: coi
// như không thực hiện"), exactly like a Guard who sleeps through the night.
func (g *Game) Night() *Night {
	g.t.Helper()
	switch g.stage {
	case stageLobby:
		g.t.Fatalf("Night() needs roles: give NewGame a Cast or Roles option")
	case stageNight:
		g.t.Fatalf("Night() called while a night is running: call Dawn() to end it first")
	case stageDawn:
		g.t.Fatalf("two nights in a row: between nights there is a day; call g.Day().NobodyHanged() or g.Day().Hanged(...)")
	case stageDay:
		g.t.Fatalf("Night() called before the day ended: finish it with NobodyHanged() or Hanged(...)")
	}
	cmds, err := g.driver.StartNight(g.state)
	g.doAll(cmds, err)
	g.stage = stageNight
	return &Night{g: g}
}

// Night collects the actions of one night. Role files add their verbs as
// methods on it, for example Seer("An").Checks("Binh").
type Night struct {
	g *Game
}

// Game returns the game the night belongs to.
func (n *Night) Game() *Game { return n.g }

// Player returns the ID of a seated player; for use by role verbs.
func (n *Night) Player(name string) engine.PlayerID {
	n.g.t.Helper()
	return n.g.Player(name)
}

// Do sends a role's command and fails the test if it is rejected.
func (n *Night) Do(c engine.Command) *Night {
	n.g.t.Helper()
	n.g.Do(c)
	return n
}

// DoExpectErr sends a command that must be rejected with want.
func (n *Night) DoExpectErr(c engine.Command, want error) *Night {
	n.g.t.Helper()
	n.g.DoExpectErr(c, want)
	return n
}
