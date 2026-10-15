package roles

import (
	"fmt"

	"masoi/internal/engine"
)

// Witch has one heal potion and one poison potion for the whole game and may
// use both in the same night (rules §6 "Phù thủy").
type witch struct{}

// Witch is the ID of the witch.
const Witch engine.RoleID = "Witch"

// Flags kept on the Witch's player: 1 once the potion is used up.
const (
	flagHealUsed   = "witch.heal_used"
	flagPoisonUsed = "witch.poison_used"
)

// WitchHeal saves the player the pack is biting tonight. It has no target: she
// can only heal the one she is shown.
type WitchHeal struct {
	engine.CommandMarker
	By engine.PlayerID
}

// WitchPoison kills any living player, herself included, at dawn.
type WitchPoison struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// WitchHealed records the use of the heal potion.
type WitchHealed struct {
	engine.EventMarker
	Witch  engine.PlayerID
	Target engine.PlayerID
}

// WitchPoisoned records the use of the poison potion.
type WitchPoisoned struct {
	engine.EventMarker
	Witch  engine.PlayerID
	Target engine.PlayerID
}

func init() { engine.Register(witch{}) }

// Audience: only the Witch knows what she brewed.
func (e WitchHealed) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Witch}}
}

// Audience: only the Witch knows what she brewed.
func (e WitchPoisoned) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Witch}}
}

func (witch) ID() engine.RoleID            { return Witch }
func (witch) BaseTeam() engine.Team        { return engine.TeamVillage }
func (witch) NightCall() *engine.NightCall { return &engine.NightCall{Order: 6} }

// SeesBite: she is told whom the pack bites only while the heal potion is left.
func (witch) SeesBite(holder *engine.Player) bool { return holder.Flags[flagHealUsed] == 0 }

func (witch) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	switch cmd := c.(type) {
	case WitchHeal:
		evs, err := decideHeal(s, cmd)
		return evs, true, err
	case WitchPoison:
		evs, err := decidePoison(s, cmd)
		return evs, true, err
	}
	return nil, false, nil
}

// witchTurn checks the shared parts: her step (NightActor also refuses a blocked player), her role.
func witchTurn(s engine.State, by engine.PlayerID) (*engine.Player, error) {
	me, err := engine.NightActor(s, by, Witch)
	if err != nil {
		return nil, err
	}
	if me.Role != Witch {
		return nil, fmt.Errorf("%w: %s is not the Witch", engine.ErrNotYourTurn, by)
	}
	return me, nil
}

func decideHeal(s engine.State, cmd WitchHeal) ([]engine.Event, error) {
	me, err := witchTurn(s, cmd.By)
	if err != nil {
		return nil, err
	}
	if me.Flags[flagHealUsed] != 0 {
		return nil, fmt.Errorf("%w: heal", engine.ErrPotionUsed)
	}
	target, bitten := engine.BiteTarget(s, engine.WolfVoters(s))
	if !bitten {
		return nil, engine.ErrNothingToHeal
	}
	return []engine.Event{WitchHealed{Witch: cmd.By, Target: target}}, nil
}

func decidePoison(s engine.State, cmd WitchPoison) ([]engine.Event, error) {
	me, err := witchTurn(s, cmd.By)
	if err != nil {
		return nil, err
	}
	if me.Flags[flagPoisonUsed] != 0 {
		return nil, fmt.Errorf("%w: poison", engine.ErrPotionUsed)
	}
	if _, err := engine.LivingTarget(s, cmd.Target); err != nil {
		return nil, err
	}
	return []engine.Event{WitchPoisoned{Witch: cmd.By, Target: cmd.Target}}, nil
}

func (witch) Evolve(s *engine.State, e engine.Event) bool {
	switch ev := e.(type) {
	case WitchHealed:
		s.Pending.Healed = true
		useFlag(s, ev.Witch, flagHealUsed)
		return true
	case WitchPoisoned:
		s.Pending.Poisoned = ev.Target
		useFlag(s, ev.Witch, flagPoisonUsed)
		return true
	}
	return false
}

func useFlag(s *engine.State, id engine.PlayerID, flag string) {
	p := s.Players[id]
	if p == nil {
		return
	}
	if p.Flags == nil {
		p.Flags = map[string]int{}
	}
	p.Flags[flag] = 1
}
