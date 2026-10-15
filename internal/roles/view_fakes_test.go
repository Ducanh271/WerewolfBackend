package roles_test

import (
	"masoi/internal/engine"
)

// Stand-ins for roles that are not built yet, so the view rules for them can
// be tested now. Each is deleted by the task that builds the real role.

// fakeWitch is called in the Witch's place (rules §3, order 6) and sees the
// bite while the hand-built "heal" flag is set. Delete in T12: the real Witch
// registers the same ID and implements engine.BiteWitness.
type fakeWitch struct{}

func (fakeWitch) ID() engine.RoleID            { return "Witch" }
func (fakeWitch) BaseTeam() engine.Team        { return engine.TeamVillage }
func (fakeWitch) NightCall() *engine.NightCall { return &engine.NightCall{Order: 6} }
func (fakeWitch) Decide(engine.State, engine.Command) ([]engine.Event, bool, error) {
	return nil, false, nil
}
func (fakeWitch) Evolve(*engine.State, engine.Event) bool { return false }
func (fakeWitch) SeesBite(p *engine.Player) bool          { return p.Flags["heal"] > 0 }

func init() { engine.Register(fakeWitch{}) }

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
