package engine

// Event is a fact about the game. The game is the list of its events; State is
// the fold of that list through Evolve.
type Event interface {
	isEvent()
}

// EventMarker is embedded by every event, including the ones declared in the
// roles package, which cannot implement the unexported marker method on its
// own.
type EventMarker struct{}

func (EventMarker) isEvent() {}

// PlayerJoined records a player entering the lobby. Nickname is already
// trimmed and validated.
type PlayerJoined struct {
	EventMarker
	Player   PlayerID
	Nickname string
}

// PlayerLeft records a player leaving the lobby. Only their own seat is
// freed (the rest keep theirs); the chosen role list is untouched. DealRoles
// re-checks the role list against the new player count before dealing.
type PlayerLeft struct {
	EventMarker
	Player PlayerID
}

// PlayerKicked records the host removing a player from the lobby (audit
// K2). Handled exactly like PlayerLeft.
type PlayerKicked struct {
	EventMarker
	Player PlayerID
}

// SeatsSet records the seating circle.
type SeatsSet struct {
	EventMarker
	Order []PlayerID
}

// RolesConfigured records the chosen role list.
type RolesConfigured struct {
	EventMarker
	Roles map[RoleID]int
}

// RolesDealt records who got which role. Teams are derived from the role's
// BaseTeam when the event is applied.
type RolesDealt struct {
	EventMarker
	Assignments map[PlayerID]RoleID
}
