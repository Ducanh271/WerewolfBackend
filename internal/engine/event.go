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
