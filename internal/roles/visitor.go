package roles

import (
	"fmt"
	"math/rand"

	"masoi/internal/engine"
)

// Visitor sleeps over at one player a night; that player loses their night
// ability (rules §6 "Nguyệt Nữ"). Her choice is compulsory: when her step runs
// out the server picks for her. She never visits herself, nor the player of the
// night before.
type visitor struct{}

// Visitor is the ID of the Visitor.
const Visitor engine.RoleID = "Visitor"

// Flags kept on the Visitor's player: the night of her last visit and whom it
// was (seat index + 1, so 0 means nobody).
const (
	flagVisitNight  = "visitor.last_night"
	flagVisitTarget = "visitor.last_target"
)

// VisitorVisit is the Visitor's choice for tonight.
type VisitorVisit struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

// PlayerVisited records the visit. The target is blocked from it on; only the
// Visitor is told whom she visited.
type PlayerVisited struct {
	engine.EventMarker
	Visitor engine.PlayerID
	Target  engine.PlayerID
	Night   int
}

func init() { engine.Register(visitor{}) }

// Audience: only the Visitor knows. The visited player learns that they are
// blocked, not by whom.
func (e PlayerVisited) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Visitor}}
}

func (visitor) ID() engine.RoleID            { return Visitor }
func (visitor) BaseTeam() engine.Team        { return engine.TeamVillage }
func (visitor) NightCall() *engine.NightCall { return &engine.NightCall{Order: 1} }

func (visitor) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(VisitorVisit)
	if !ok {
		return nil, false, nil
	}
	me, err := engine.NightActor(s, cmd.By, Visitor)
	if err != nil {
		return nil, true, err
	}
	if hasVisited(s) {
		return nil, true, engine.ErrAlreadyActed
	}
	if _, err := engine.LivingTarget(s, cmd.Target); err != nil {
		return nil, true, err
	}
	if cmd.Target == cmd.By {
		return nil, true, fmt.Errorf("%w: the Visitor cannot visit herself", engine.ErrInvalidTarget)
	}
	if visitedLastNight(s, me, cmd.Target) {
		return nil, true, fmt.Errorf("%w: %s", engine.ErrSameVisitorTarget, cmd.Target)
	}
	return []engine.Event{PlayerVisited{Visitor: cmd.By, Target: cmd.Target, Night: s.Night}}, true, nil
}

// OnNightTimeout makes the compulsory visit when her step ran out: a random
// choice among the valid players, the seed coming from the room. A dead Visitor
// (a fake call), a blocked one, one who is dead by the rules from this step
// (grief or poison), or one who already chose does nothing, and so does a
// table with nobody left to visit.
func (visitor) OnNightTimeout(s engine.State, seed int64) []engine.Event {
	me := holder(s)
	if me == nil || !me.Alive || engine.Blocked(s, me.ID) || engine.Doomed(s, me.ID) || hasVisited(s) {
		return nil
	}
	var pool []engine.PlayerID
	for _, id := range s.Seats {
		if p := s.Players[id]; p != nil && p.Alive && id != me.ID && !visitedLastNight(s, me, id) {
			pool = append(pool, id)
		}
	}
	if len(pool) == 0 {
		return nil
	}
	target := pool[rand.New(rand.NewSource(seed)).Intn(len(pool))]
	return []engine.Event{PlayerVisited{Visitor: me.ID, Target: target, Night: s.Night}}
}

func (visitor) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(PlayerVisited)
	if !ok {
		return false
	}
	engine.Block(s, ev.Target, Visitor)
	if p := s.Players[ev.Visitor]; p != nil {
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		p.Flags[flagVisitNight] = ev.Night
		p.Flags[flagVisitTarget] = seatNumber(*s, ev.Target)
	}
	return true
}

// hasVisited: she already chose tonight.
func hasVisited(s engine.State) bool {
	for _, source := range s.Pending.Blocked {
		if source == Visitor {
			return true
		}
	}
	return false
}

func visitedLastNight(s engine.State, me *engine.Player, target engine.PlayerID) bool {
	return me.Flags[flagVisitNight] == s.Night-1 && me.Flags[flagVisitTarget] == seatNumber(s, target)
}

// holder is the player holding the Visitor role, nil when nobody does.
func holder(s engine.State) *engine.Player {
	for _, id := range s.Seats {
		if p := s.Players[id]; p != nil && p.Role == Visitor {
			return p
		}
	}
	return nil
}
