package engine

import "fmt"

// NightCall describes when a role is woken at night.
type NightCall struct {
	// Order places the role in the night call order (rules §3). Lower is
	// called first.
	Order int
	// Nights limits the call to those night numbers, e.g. {1, 2, 3} for the
	// cursed wolf. Empty means every night.
	Nights []int
}

// CalledOn reports whether the role is woken on the given 1-based night.
func (c NightCall) CalledOn(night int) bool {
	if len(c.Nights) == 0 {
		return true
	}
	for _, n := range c.Nights {
		if n == night {
			return true
		}
	}
	return false
}

// Role is one role. Each role lives in its own file in internal/roles and
// handles only its own commands and events.
type Role interface {
	ID() RoleID
	BaseTeam() Team
	// NightCall returns nil when the role is not called at night.
	NightCall() *NightCall
	// Decide handles the role's own commands. handled is false when the
	// command belongs to someone else.
	Decide(s State, c Command) (evs []Event, handled bool, err error)
	// Evolve applies the role's own events and reports whether it did.
	Evolve(s *State, e Event) bool
}

// MinPlayersRule is implemented by roles that need a minimum table size,
// e.g. Detective (more than 8 players) and Cupid (10 or more). Rules §2.
type MinPlayersRule interface {
	MinPlayers() int
}

// MultiCopyRole is implemented by the roles that may be dealt to more than one
// player: Villager and the plain Werewolf. Every other role is unique, at most
// one copy per game (rules §2 "Số lượng mỗi vai").
type MultiCopyRole interface {
	AllowsMultipleCopies() bool
}

// MinCopiesRole is implemented by a role every game must deal at least once,
// like the Villager (rules §2 "Dân làng: 1 trở lên"). Unlike MinPlayersRule
// (a table-size floor), this is a floor on the role's own count.
type MinCopiesRole interface {
	MinCopies() int
}

// CurseSource is implemented by roles that turn a villager into a wolf during
// the game. The wolf limit of rules §2 counts that future wolf, so one wolf
// slot is reserved when such a role is selected.
type CurseSource interface {
	CursesVillager() bool
}

// AlwaysCalled is implemented by the role of the wolf pack step. Every game has
// a wolf, and every wolf role bites in that step, so it is called every night
// even when the role itself is not in the role list (rules §3).
type AlwaysCalled interface {
	AlwaysCalled() bool
}

// BiteWitness is implemented by a role whose holder learns, on their own night
// step, whom the pack is biting: the Witch while she still has the heal potion
// (rules §6 "Phù thủy"). The view shows it only during that step and not when
// the holder was blocked.
type BiteWitness interface {
	SeesBite(holder *Player) bool
}

// NightTimeoutRole is implemented by a role that must still act when its night
// step runs out without a choice: the Visitor is forced to visit (rules §6
// "Nguyệt Nữ"). seed comes from the room layer (crypto/rand), never from a
// client. It returns no events when the holder is dead (a fake call) or has
// nothing valid to do.
type NightTimeoutRole interface {
	OnNightTimeout(s State, seed int64) []Event
}

// Registry holds every known role. Roles add themselves with Register in
// init(); the engine never names a role itself.
var Registry = map[RoleID]Role{}

// Register adds a role to Registry. It panics on a duplicate or empty ID,
// which can only be a programming error at startup.
func Register(r Role) {
	if r == nil {
		panic("engine: Register(nil)")
	}
	id := r.ID()
	if id == "" {
		panic("engine: role with empty ID")
	}
	if _, dup := Registry[id]; dup {
		panic(fmt.Sprintf("engine: role %q registered twice", id))
	}
	Registry[id] = r
}
