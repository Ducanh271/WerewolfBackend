package engine

import "time"

// State is the whole game. It is rebuilt by folding events with Evolve and is
// never mutated in place: every Evolve returns a new State.
type State struct {
	Phase     Phase
	Seq       int // increments on every phase or night step change
	Night     int // 1-based night counter
	NightStep int // night steps called so far; the current step is steps[NightStep-1]
	Seats     []PlayerID
	Players   map[PlayerID]*Player
	Pending   NightActions
	Day       DayState
	Lovers    *LoversLink
	Settings  Settings
	Dealt     bool // roles were dealt to the current player list
	Winner    *Team

	Chief      PlayerID              // holder of the title, empty if nobody
	ChiefVotes map[PlayerID]PlayerID // voter -> candidate, during the election
	Succession Succession            // set while PhaseChiefSuccession runs
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
// resolved at dawn. Each role's own events fill in its part, and the night
// start clears it. More fields arrive with the later night roles.
type NightActions struct {
	WolfVotes  map[PlayerID]PlayerID // each wolf's current bite choice
	Protected  PlayerID              // the Guard's choice, empty if none
	SeerTarget PlayerID              // whom the Seer checked, empty if nobody
	Visited    PlayerID              // the Visitor's target (T14), empty if nobody
}

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

	// Mode is how the table plays. Only InPerson exists in v1.
	Mode Mode

	RoleView       time.Duration // everyone looks at their role after the deal
	Announce       time.Duration // pause after the dawn announcement and at dusk
	NightStep      time.Duration // each role's turn at night, real or fake
	Day            time.Duration // discussion and nominations
	Defense        time.Duration // bào chữa
	Verdict        time.Duration // vote giết hay tha
	NominationLock time.Duration // đếm ngược chốt sau khi mọi người đã chọn đề cử
	HunterShot     time.Duration // thợ săn chọn người

	ChiefElection   time.Duration // public vote for the Chief
	ChiefSuccession time.Duration // the dead Chief names a successor
}

// Mode is how a game is played. Only InPerson is supported in v1; Online is
// reserved so the rest of the design does not have to change later.
type Mode int

// Modes.
const (
	InPerson Mode = iota
	Online
)

func (m Mode) String() string {
	if m == Online {
		return "Online"
	}
	return "InPerson"
}

// DefaultSettings returns the default timers from rules §2.
func DefaultSettings() Settings {
	return Settings{
		RoleView:       30 * time.Second,
		Announce:       10 * time.Second,
		NightStep:      30 * time.Second,
		Day:            7 * time.Minute,
		Defense:        1 * time.Minute,
		Verdict:        30 * time.Second,
		NominationLock: 10 * time.Second,
		HunterShot:     15 * time.Second,

		ChiefElection:   1 * time.Minute,
		ChiefSuccession: 15 * time.Second,
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
	if s.Pending.WolfVotes != nil {
		out.Pending.WolfVotes = make(map[PlayerID]PlayerID, len(s.Pending.WolfVotes))
		for wolf, target := range s.Pending.WolfVotes {
			out.Pending.WolfVotes[wolf] = target
		}
	}

	if s.ChiefVotes != nil {
		out.ChiefVotes = make(map[PlayerID]PlayerID, len(s.ChiefVotes))
		for voter, cand := range s.ChiefVotes {
			out.ChiefVotes[voter] = cand
		}
	}
	out.Day = s.Day.clone()
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
