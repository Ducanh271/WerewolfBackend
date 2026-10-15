package roles

import (
	"fmt"
	"math/rand"

	"masoi/internal/engine"
)

// Cupid pairs two living players on night 1 (rules §7). He may pair himself.
// The lovers learn of the pair at the phone check; Cupid never sees the pair on
// his screen afterwards. His own team stays Village: a couple of one wolf and
// one villager is the third team, worked out for the couple in engine.CheckWin,
// never by changing Player.Team.
type cupid struct{}

// Cupid is the ID of Cupid.
const Cupid engine.RoleID = "Cupid"

// cupidMinPlayers is the smallest table with Cupid (rules §2).
const cupidMinPlayers = 10

// CupidPair is Cupid's choice of the two players to make a couple.
type CupidPair struct {
	engine.CommandMarker
	By engine.PlayerID
	A  engine.PlayerID
	B  engine.PlayerID
}

// LoversPaired records the couple. Only Cupid gets it, as the answer to his
// choice; the lovers are told at the phone check.
type LoversPaired struct {
	engine.EventMarker
	Cupid engine.PlayerID
	A     engine.PlayerID
	B     engine.PlayerID
}

func init() { engine.Register(cupid{}) }

// Audience: Cupid alone.
func (e LoversPaired) Audience(engine.State) engine.Audience {
	return engine.Audience{Players: []engine.PlayerID{e.Cupid}}
}

func (cupid) ID() engine.RoleID     { return Cupid }
func (cupid) BaseTeam() engine.Team { return engine.TeamVillage }
func (cupid) MinPlayers() int       { return cupidMinPlayers }
func (cupid) CreatesLovers() bool   { return true }

// NightCall: first of the night, the first night only.
func (cupid) NightCall() *engine.NightCall {
	return &engine.NightCall{Order: 0, Nights: []int{1}}
}

func (cupid) Decide(s engine.State, c engine.Command) ([]engine.Event, bool, error) {
	cmd, ok := c.(CupidPair)
	if !ok {
		return nil, false, nil
	}
	me, err := engine.NightActor(s, cmd.By, Cupid)
	if err != nil {
		return nil, true, err
	}
	if me.Role != Cupid {
		return nil, true, fmt.Errorf("%w: %s is not Cupid", engine.ErrNotYourTurn, cmd.By)
	}
	if s.Lovers != nil {
		return nil, true, engine.ErrAlreadyActed
	}
	if cmd.A == cmd.B {
		return nil, true, fmt.Errorf("%w: a couple needs two different players", engine.ErrInvalidTarget)
	}
	for _, id := range []engine.PlayerID{cmd.A, cmd.B} {
		if _, err := engine.LivingTarget(s, id); err != nil {
			return nil, true, err
		}
	}
	return []engine.Event{LoversPaired{Cupid: cmd.By, A: cmd.A, B: cmd.B}}, true, nil
}

// OnNightTimeout pairs two random living players when Cupid's step ran out
// without a choice; the seed comes from the room.
func (cupid) OnNightTimeout(s engine.State, seed int64) []engine.Event {
	var me *engine.Player
	var pool []engine.PlayerID
	for _, id := range s.Seats {
		p := s.Players[id]
		if p == nil || !p.Alive {
			continue
		}
		pool = append(pool, id)
		if p.Role == Cupid {
			me = p
		}
	}
	if me == nil || s.Lovers != nil || len(pool) < 2 {
		return nil
	}
	pick := rand.New(rand.NewSource(seed)).Perm(len(pool))
	return []engine.Event{LoversPaired{Cupid: me.ID, A: pool[pick[0]], B: pool[pick[1]]}}
}

func (cupid) Evolve(s *engine.State, e engine.Event) bool {
	ev, ok := e.(LoversPaired)
	if !ok {
		return false
	}
	s.Lovers = &engine.LoversLink{A: ev.A, B: ev.B}
	return true
}
