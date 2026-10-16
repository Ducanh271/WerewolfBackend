package roles

import "masoi/internal/engine"

// Villager has no power, only the day vote (rules §6).
type villager struct{}

// ID of the plain villager.
const Villager engine.RoleID = "Villager"

func init() { engine.Register(villager{}) }

func (villager) ID() engine.RoleID                       { return Villager }
func (villager) BaseTeam() engine.Team                   { return engine.TeamVillage }
func (villager) NightCall() *engine.NightCall            { return nil }
func (villager) AllowsMultipleCopies() bool              { return true }
func (villager) MinCopies() int                          { return 1 } // rules §2: "Dân làng: 1 trở lên"
func (villager) Evolve(*engine.State, engine.Event) bool { return false }
func (villager) Decide(engine.State, engine.Command) ([]engine.Event, bool, error) {
	return nil, false, nil
}
