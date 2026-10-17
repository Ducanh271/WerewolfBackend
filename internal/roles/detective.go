package roles

import (
	"masoi/internal/engine"
	"masoi/internal/i18n"
)

// Detective checks one player once a game and learns whether there is a wolf
// among that player and the nearest living neighbour on each side (rules §6
// "Thám tử"). He is called every night, even after using the ability: the
// later calls are fake and last as long, so the table cannot tell.
type detective struct{}

// Detective is the ID of the Detective.
const Detective engine.RoleID = "Detective"

const flagDetectiveUsed = "detective.used"

// DetectiveCheck is the Detective's choice for tonight. Any living player may
// be checked, himself included.
type DetectiveCheck struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// DetectiveChecked is the Detective's private result: only whether the zone
// holds a wolf, never how many or who. A cursed player counts as a wolf.
type DetectiveChecked struct {
	engine.EventMarker
	Detective engine.PlayerID
	Target    engine.PlayerID
	HasWolf   bool
}

func init() { engine.Register(detective{}) }

// Audience: the result is the Detective's alone.
func (e DetectiveChecked) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Detective}}
}

// Narrate: the end-of-game summary (audit K4) says what he found.
func (e DetectiveChecked) Narrate(name func(engine.PlayerID) string) (int, string, bool) {
	key := i18n.SummaryDetectiveClear
	if e.HasWolf {
		key = i18n.SummaryDetectiveWolf
	}
	return engine.NarrateDetective, i18n.Vi(key, name(e.Target)), true
}

func (detective) ID() engine.RoleID            { return Detective }
func (detective) BaseTeam() engine.Team        { return engine.TeamVillage }
func (detective) NightCall() *engine.NightCall { return &engine.NightCall{Order: 8} }

// MinPlayers: the Detective needs a table of more than 8 (rules §2).
func (detective) MinPlayers() int { return 9 }

func (detective) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(DetectiveCheck)
	if !ok {
		return nil, false, nil
	}
	me, err := engine.NightActor(s, cmd.By, Detective)
	if err != nil {
		return nil, true, err
	}
	if me.Flags[flagDetectiveUsed] != 0 {
		return nil, true, engine.ErrAbilityUsed
	}
	if _, err := engine.LivingTarget(s, cmd.Target); err != nil {
		return nil, true, err
	}
	hasWolf := false
	for _, id := range detectiveZone(s, cmd.Target) {
		if s.Players[id].Team == engine.TeamWolves {
			hasWolf = true
		}
	}
	return []engine.Event{DetectiveChecked{Detective: cmd.By, Target: cmd.Target, HasWolf: hasWolf}}, true, nil
}

func (detective) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(DetectiveChecked)
	if !ok {
		return false
	}
	if p := s.Players[ev.Detective]; p != nil {
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		p.Flags[flagDetectiveUsed] = 1
	}
	return true
}

// detectiveZone is the target and the nearest living player on each side, once
// each. With fewer than 3 living players it is simply everyone alive.
func detectiveZone(s engine.State, target engine.PlayerID) []engine.PlayerID {
	zone := []engine.PlayerID{target}
	for _, step := range []int{-1, 1} {
		id := engine.NearestLiving(s, target, step)
		if id == "" {
			continue
		}
		dup := false
		for _, z := range zone {
			dup = dup || z == id
		}
		if !dup {
			zone = append(zone, id)
		}
	}
	return zone
}
