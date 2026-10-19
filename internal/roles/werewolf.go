package roles

import (
	"errors"
	"fmt"

	"masoi/internal/engine"
)

// Werewolf is the plain wolf. Its night step is the pack step, where every wolf
// (Werewolf, CursedWolf, WolfSeer) votes on the bite (rules §5 "Sói").
type werewolf struct{}

// Werewolf is the ID of the plain wolf and of the pack step.
const Werewolf engine.RoleID = "Werewolf"

// WolfVote sets a wolf's bite choice for tonight; sending it again changes the
// vote. An empty Target (engine.NoBite) is the vote "Không cắn ai".
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
func (werewolf) CardOrder() int               { return 10 } // rank in the rules §9b table
func (werewolf) AllowsMultipleCopies() bool   { return true }
func (werewolf) AlwaysCalled() bool           { return true }
func (werewolf) NightCall() *engine.NightCall { return &engine.NightCall{Order: 5} }

// wolfVoteArgs keeps "no target field" apart from "target": "", the vote
// "Không cắn ai": only the second is a vote.
type wolfVoteArgs struct {
	Target *engine.PlayerID `json:"target"`
}

func (a wolfVoteArgs) Validate() error {
	if a.Target == nil {
		return errors.New(`wolfVote needs "target" (use "" for no bite)`)
	}
	return nil
}

// Commands: every wolf's bite vote on the wire.
func (werewolf) Commands() []engine.CommandSpec {
	return []engine.CommandSpec{
		engine.WireCommand("wolfVote", func(a wolfVoteArgs, by engine.PlayerID) engine.Command {
			return WolfVote{By: by, Target: *a.Target}
		}),
	}
}

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
	if cmd.Target != engine.NoBite {
		if _, err := engine.LivingTarget(s, cmd.Target); err != nil {
			return nil, true, err
		}
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
