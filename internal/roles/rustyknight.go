package roles

import "masoi/internal/engine"

// RustyKnight has a passive ability: when the pack bites him (and the Guard
// does not shield him) the nearest living wolf on his left is poisoned and dies
// when the pack step starts the next night (rules §6 "Hiệp sĩ kiếm gỉ"). The
// night pipeline asks for it (step 4); he has no night step and no command, and
// the Visitor cannot block it.
type rustyKnight struct{}

// RustyKnight is the ID of the Rusty knight.
const RustyKnight engine.RoleID = "RustyKnight"

// leftOfKnight is the direction of "bên trái" in seating order: the next seat,
// the seats being numbered clockwise.
const leftOfKnight = 1

func init() { engine.Register(rustyKnight{}) }

func (rustyKnight) ID() engine.RoleID                       { return RustyKnight }
func (rustyKnight) BaseTeam() engine.Team                   { return engine.TeamVillage }
func (rustyKnight) NightCall() *engine.NightCall            { return nil }
func (rustyKnight) Evolve(*engine.State, engine.Event) bool { return false }
func (rustyKnight) Decide(engine.State, engine.Command) ([]engine.Event, bool, error) {
	return nil, false, nil
}

// AvengeBite poisons the nearest living wolf-team player to the left of the
// Knight, never the Knight himself. A cursed Knight keeps the ability.
func (rustyKnight) AvengeBite(s engine.State, holder engine.PlayerID) engine.PlayerID {
	return engine.NearestLivingWolf(s, holder, leftOfKnight)
}
