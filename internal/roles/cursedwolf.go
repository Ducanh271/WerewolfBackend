package roles

import (
	"fmt"

	"masoi/internal/engine"
	"masoi/internal/i18n"
)

// CursedWolf curses one living non-wolf, once per game, on one of the first
// three nights (rules §5 "Sói nguyền"). The curse takes effect at once: the
// target is on the wolf team, keeps their role and bites with the pack the
// same night. Their phone and the pack's only change at the phone check step
// that follows (engine.StepPhoneCheck). The Guard cannot stop it: it does not
// pass through the bite.
type cursedWolf struct{}

// CursedWolf is the ID of the Cursed wolf.
const CursedWolf engine.RoleID = "CursedWolf"

// flagCurseUsed is kept on the Cursed wolf once the curse is spent.
const flagCurseUsed = "curse.used"

// CursedWolfCurse is the Cursed wolf's choice of whom to curse.
type CursedWolfCurse struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// PlayerCursed records the curse. Only the Cursed wolf is told at once; the
// target learns at the phone check.
type PlayerCursed struct {
	engine.EventMarker
	Wolf   engine.PlayerID
	Target engine.PlayerID
}

func init() { engine.Register(cursedWolf{}) }

// Audience: the Cursed wolf alone.
func (e PlayerCursed) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Wolf}}
}

// Narrate: the end-of-game summary (audit K4) says whom he cursed.
func (e PlayerCursed) Narrate(name func(engine.PlayerID) string) (int, string, bool) {
	return engine.NarrateCurse, i18n.Vi(i18n.SummaryCurse, name(e.Target)), true
}

func (cursedWolf) ID() engine.RoleID     { return CursedWolf }
func (cursedWolf) BaseTeam() engine.Team { return engine.TeamWolves }
func (cursedWolf) CursesVillager() bool  { return true }

// NightCall: before the Wolf seer and the pack, on the first three nights.
func (cursedWolf) NightCall() *engine.NightCall {
	return &engine.NightCall{Order: 3, Nights: []int{1, 2, 3}}
}

func (cursedWolf) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(CursedWolfCurse)
	if !ok {
		return nil, false, nil
	}
	me, err := engine.NightActor(s, cmd.By, CursedWolf)
	if err != nil {
		return nil, true, err
	}
	if me.Flags[flagCurseUsed] != 0 {
		return nil, true, engine.ErrAlreadyActed
	}
	target, err := engine.LivingTarget(s, cmd.Target)
	if err != nil {
		return nil, true, err
	}
	if target.Team == engine.TeamWolves {
		return nil, true, fmt.Errorf("%w: a wolf cannot be cursed", engine.ErrInvalidTarget)
	}
	return []engine.Event{
		PlayerCursed{Wolf: cmd.By, Target: cmd.Target},
		engine.TeamChanged{Player: cmd.Target, Team: engine.TeamWolves},
	}, true, nil
}

func (cursedWolf) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(PlayerCursed)
	if !ok {
		return false
	}
	if p := s.Players[ev.Wolf]; p != nil {
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		p.Flags[flagCurseUsed] = 1
	}
	return true
}
