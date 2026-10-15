package roles_test

import (
	"masoi/internal/engine"
)

// Stand-ins for roles that are not built yet, so the view rules for them can
// be tested now. Each is deleted by the task that builds the real role.

// wolfSeerSaw stands for the Wolf seer's result (T16): it is shared with the
// whole pack (rules §5 "Sói tiên tri"). The real event must declare the same
// audience.
type wolfSeerSaw struct {
	engine.EventMarker
	Target engine.PlayerID
	IsSeer bool
}

func (wolfSeerSaw) Audience(engine.State) engine.Audience { return engine.Audience{Wolves: true} }

// cursed stands for the curse (T15): the player joins the wolf team.
func cursed(name string) engine.Event {
	return engine.TeamChanged{Player: engine.PlayerID(name), Team: engine.TeamWolves}
}
