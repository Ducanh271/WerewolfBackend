package engine

import "time"

// Player count limits (rules §2: "Một ván có 5–30 người").
const (
	MinPlayers = 5
	MaxPlayers = 30
)

// State is the whole game. It is rebuilt by folding events with Evolve and is
// never mutated in place: every Evolve returns a new State.
type State struct {
	Phase     Phase
	Seq       int // increments on every phase change
	Night     int // 1-based night counter
	NightStep int // index into the night call order
	Seats     []PlayerID
	Players   map[PlayerID]*Player
	Pending   NightActions
	Day       DayState
	Lovers    *LoversLink
	Settings  Settings
	Dealt     bool // roles were dealt to the current player list
	Winner    *Team
}

// Player is one participant. Role-specific bookkeeping lives in Flags so that
// adding a role never changes this struct.
type Player struct {
	ID       PlayerID
	Nickname string
	Role     RoleID
	Team     Team
	Alive    bool
	Flags    map[string]int
}

// NightActions holds the choices collected during one night; they are only
// resolved at dawn. Filled in by the night tasks.
type NightActions struct{}

// DayState holds nominations, skip votes, the trial and the day clock.
// Filled in by the day phase task.
type DayState struct{}

// LoversLink is the pair created by Cupid.
type LoversLink struct {
	A PlayerID
	B PlayerID
}

// Settings is everything the host configures before the deal: the role list
// and the phase timers.
type Settings struct {
	// Roles maps a role to how many players get it. The sum equals the
	// number of players.
	Roles map[RoleID]int

	NightStep  time.Duration // each role's turn at night
	Day        time.Duration // discussion and nominations
	Defense    time.Duration // bào chữa
	Verdict    time.Duration // vote giết hay tha
	HunterShot time.Duration // thợ săn chọn người
}

// DefaultSettings returns the default timers from rules §2.
func DefaultSettings() Settings {
	return Settings{
		NightStep:  30 * time.Second,
		Day:        7 * time.Minute,
		Defense:    1 * time.Minute,
		Verdict:    30 * time.Second,
		HunterShot: 15 * time.Second,
	}
}

// NewState returns an empty game waiting in the lobby.
func NewState() State {
	return State{
		Phase:    PhaseLobby,
		Players:  map[PlayerID]*Player{},
		Settings: DefaultSettings(),
	}
}

// clone returns a copy that shares nothing mutable with s, so Evolve can
// write to the copy without touching the caller's State.
func (s State) clone() State {
	out := s

	out.Players = make(map[PlayerID]*Player, len(s.Players))
	for id, p := range s.Players {
		cp := *p
		cp.Flags = copyFlags(p.Flags)
		out.Players[id] = &cp
	}

	if s.Seats != nil {
		out.Seats = append([]PlayerID(nil), s.Seats...)
	}
	out.Settings.Roles = copyRoles(s.Settings.Roles)

	if s.Lovers != nil {
		lovers := *s.Lovers
		out.Lovers = &lovers
	}
	if s.Winner != nil {
		winner := *s.Winner
		out.Winner = &winner
	}
	return out
}

func copyFlags(in map[string]int) map[string]int {
	if in == nil {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyRoles(in map[RoleID]int) map[RoleID]int {
	if in == nil {
		return nil
	}
	out := make(map[RoleID]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
