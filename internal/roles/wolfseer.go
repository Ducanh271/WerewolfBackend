package roles

import (
	"masoi/internal/engine"
	"masoi/internal/i18n"
)

// WolfSeer checks one player a night and learns whether they are the Seer
// (rules §5 "Sói tiên tri"). The answer goes to the whole pack, which is
// called right after. The Wolf seer also bites with the pack: the pack step
// accepts every wolf role.
type wolfSeer struct{}

// WolfSeer is the ID of the Wolf seer.
const WolfSeer engine.RoleID = "WolfSeer"

// WolfSeerCheck is the Wolf seer's choice for tonight. Any living player may
// be checked, wolves and the Wolf seer himself included.
type WolfSeerCheck struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// WolfSeerChecked is the result, shared with the pack. It is a role check, not
// a team check: a cursed Seer is still the Seer.
type WolfSeerChecked struct {
	engine.EventMarker
	WolfSeer engine.PlayerID
	Target   engine.PlayerID
	IsSeer   bool
}

func init() { engine.Register(wolfSeer{}) }

// Audience: the whole pack. A player cursed earlier tonight gets it from the
// phone check on, like the rest of the pack screen (engine.VisibleTeam).
func (WolfSeerChecked) Audience(engine.State) engine.Audience {
	return engine.Audience{Wolves: true}
}

// Narrate: the end-of-game summary (audit K4) says what he saw.
func (e WolfSeerChecked) Narrate(name func(engine.PlayerID) string) (int, string, engine.PlayerID, bool) {
	key := i18n.SummaryWolfSeerNotSeer
	if e.IsSeer {
		key = i18n.SummaryWolfSeerIsSeer
	}
	return engine.NarrateWolfSeer, i18n.Vi(key, name(e.Target)), e.Target, true
}

func (wolfSeer) ID() engine.RoleID            { return WolfSeer }
func (wolfSeer) BaseTeam() engine.Team        { return engine.TeamWolves }
func (wolfSeer) NightCall() *engine.NightCall { return &engine.NightCall{Order: 4} }

// Commands: the Wolf seer's check on the wire.
func (wolfSeer) Commands() []engine.CommandSpec {
	return []engine.CommandSpec{
		engine.WireCommand("wolfSeerCheck", func(a targetArgs, by engine.PlayerID) engine.Command {
			return WolfSeerCheck{By: by, Target: a.Target}
		}),
	}
}

func (wolfSeer) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(WolfSeerCheck)
	if !ok {
		return nil, false, nil
	}
	_, err := engine.NightActor(s, cmd.By, WolfSeer)
	if err != nil {
		return nil, true, err
	}
	if s.Pending.WolfSeerTarget != "" {
		return nil, true, engine.ErrAlreadyActed
	}
	target, err := engine.LivingTarget(s, cmd.Target)
	if err != nil {
		return nil, true, err
	}
	return []engine.Event{WolfSeerChecked{
		WolfSeer: cmd.By,
		Target:   cmd.Target,
		IsSeer:   target.Role == Seer,
	}}, true, nil
}

func (wolfSeer) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(WolfSeerChecked)
	if !ok {
		return false
	}
	s.Pending.WolfSeerTarget = ev.Target
	return true
}
