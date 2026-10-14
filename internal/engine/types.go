package engine

// PlayerID identifies a player inside one game. It is assigned by the room
// layer when a player joins and never reused.
type PlayerID string

// RoleID identifies a role, e.g. "Werewolf". Role IDs are declared by the
// roles themselves (see Registry); the engine never hardcodes them.
type RoleID string

// Team is a win condition group. A player's team can change during the game
// (a cursed villager becomes a wolf, a lover joins the lovers team).
type Team int

// Teams.
const (
	TeamNone    Team = iota // not assigned yet (lobby)
	TeamVillage             // phe Dân
	TeamWolves              // phe Sói
	TeamLovers              // phe thứ ba: cặp đôi khác phe
)

func (t Team) String() string {
	switch t {
	case TeamVillage:
		return "Village"
	case TeamWolves:
		return "Wolves"
	case TeamLovers:
		return "Lovers"
	default:
		return "None"
	}
}

// Phase is the stage of the game. Seq increments on every phase change so
// stale timeouts can be ignored.
type Phase int

// Phases.
const (
	PhaseLobby Phase = iota
	PhaseRoleView
	PhaseNight
	PhaseDawn
	PhaseDay
	PhaseTrial
	PhaseVerdict
	PhaseHunterShot
	PhaseEnded
)

func (p Phase) String() string {
	switch p {
	case PhaseLobby:
		return "Lobby"
	case PhaseRoleView:
		return "RoleView"
	case PhaseNight:
		return "Night"
	case PhaseDawn:
		return "Dawn"
	case PhaseDay:
		return "Day"
	case PhaseTrial:
		return "Trial"
	case PhaseVerdict:
		return "Verdict"
	case PhaseHunterShot:
		return "HunterShot"
	case PhaseEnded:
		return "Ended"
	default:
		return "Unknown"
	}
}
