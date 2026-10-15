package roles

import (
	"fmt"

	"masoi/internal/engine"
)

// Werewolf is the plain wolf. Its night step is the pack step, where every wolf
// (Werewolf, CursedWolf, WolfSeer) votes on the bite (rules §5 "Sói").
type werewolf struct{}

// Werewolf is the ID of the plain wolf and of the pack step.
const Werewolf engine.RoleID = "Werewolf"

// WolfVote sets a wolf's bite choice for tonight; sending it again changes the
// vote.
type WolfVote struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// WolfVoted records a wolf's current bite choice.
type WolfVoted struct {
	engine.EventMarker
	Wolf   engine.PlayerID
	Target engine.PlayerID
}

func init() { engine.Register(werewolf{}) }

// Audience: the live votes are shown to the whole pack (rules §3).
func (WolfVoted) Audience(engine.State) engine.Audience { return engine.Audience{Wolves: true} }

func (werewolf) ID() engine.RoleID            { return Werewolf }
func (werewolf) BaseTeam() engine.Team        { return engine.TeamWolves }
func (werewolf) AllowsMultipleCopies() bool   { return true }
func (werewolf) AlwaysCalled() bool           { return true }
func (werewolf) NightCall() *engine.NightCall { return &engine.NightCall{Order: 5} }

func (werewolf) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(WolfVote)
	if !ok {
		return nil, false, nil
	}
	if _, err := engine.NightActor(s, cmd.By, Werewolf); err != nil {
		return nil, true, err
	}
	if !isVoter(s, cmd.By) {
		return nil, true, fmt.Errorf("%w: %s cannot vote tonight", engine.ErrNotYourTurn, cmd.By)
	}
	target, err := engine.LivingTarget(s, cmd.Target)
	if err != nil {
		return nil, true, err
	}
	if target.Team == engine.TeamWolves {
		return nil, true, fmt.Errorf("%w: %s", engine.ErrTargetIsWolf, cmd.Target)
	}
	return []engine.Event{WolfVoted{Wolf: cmd.By, Target: cmd.Target}}, true, nil
}

func (werewolf) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(WolfVoted)
	if !ok {
		return false
	}
	if s.Pending.WolfVotes == nil {
		s.Pending.WolfVotes = map[engine.PlayerID]engine.PlayerID{}
	}
	s.Pending.WolfVotes[ev.Wolf] = ev.Target
	return true
}

func isVoter(s engine.State, id engine.PlayerID) bool {
	for _, v := range engine.WolfVoters(s) {
		if v == id {
			return true
		}
	}
	return false
}
