package roles

import "masoi/internal/engine"

// Hunter is called by no night step. When he dies, from any cause, his role is
// revealed and he shoots one living player (rules §6 "Thợ săn"). The flow, the
// command (engine.HunterShoot) and the events live in the engine because the
// shot ends in the shared aftermath of deaths; the role only opts in.
type hunter struct{}

// Hunter is the ID of the hunter.
const Hunter engine.RoleID = "Hunter"

func init() { engine.Register(hunter{}) }

func (hunter) ID() engine.RoleID            { return Hunter }
func (hunter) BaseTeam() engine.Team        { return engine.TeamVillage }
func (hunter) CardOrder() int               { return 80 } // rank in the rules §9b table
func (hunter) NightCall() *engine.NightCall { return nil }
func (hunter) ShootsOnDeath() bool          { return true }

func (hunter) Decide(engine.State, engine.Command) ([]engine.Event, bool, error) {
	return nil, false, nil
}

func (hunter) Evolve(*engine.State, engine.Event) bool { return false }
