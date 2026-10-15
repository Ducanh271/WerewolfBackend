package engine

import "errors"

// All errors returned by Decide. Player input never panics: it comes back as
// one of these, wrapped with %w when the message carries detail.
var (
	// ErrUnknownCommand means no part of the engine and no role claimed the
	// command.
	ErrUnknownCommand = errors.New("unknown command")
	// ErrWrongPhase means the command is not allowed in the current phase.
	ErrWrongPhase = errors.New("command not allowed in this phase")

	// Lobby.
	ErrInvalidPlayerID = errors.New("invalid player id")
	ErrInvalidNickname = errors.New("invalid nickname")
	ErrNicknameTaken   = errors.New("nickname already taken")
	ErrAlreadyJoined   = errors.New("player already joined")
	ErrRoomFull        = errors.New("room is full")
	ErrUnknownPlayer   = errors.New("unknown player")
	ErrSeatsMismatch   = errors.New("seating must list every player exactly once")
	ErrDuplicateSeat   = errors.New("player seated twice")
	// ErrTimerOutOfRange means a phase time is outside its allowed range.
	ErrTimerOutOfRange = errors.New("timer out of range")
	// ErrModeUnsupported means the game mode exists in the model but is not
	// built yet (Online).
	ErrModeUnsupported = errors.New("game mode not supported yet")

	// Dealing (rules §2).
	ErrRolesNotConfigured = errors.New("roles are not configured")
	ErrSeatsNotSet        = errors.New("seats must be set for every player before dealing")
	ErrAlreadyDealt       = errors.New("roles were already dealt")

	// Role configuration (rules §2).
	ErrTooFewPlayers        = errors.New("too few players")
	ErrTooManyPlayers       = errors.New("too many players")
	ErrUnknownRole          = errors.New("unknown role")
	ErrInvalidRoleCount     = errors.New("invalid role count")
	ErrRoleMustBeUnique     = errors.New("role can only be dealt to one player")
	ErrRoleCountMismatch    = errors.New("role slots must match the number of players")
	ErrNoWolves             = errors.New("the game needs at least one wolf")
	ErrTooManyWolves        = errors.New("too many wolves for this number of players")
	ErrRoleNeedsMorePlayers = errors.New("role needs more players")

	// Night actions (rules §3, §5, §6).
	ErrNotYourTurn     = errors.New("not your turn")
	ErrNotAlive        = errors.New("player is dead")
	ErrInvalidTarget   = errors.New("invalid target")
	ErrTargetIsWolf    = errors.New("the pack cannot bite a wolf")
	ErrAlreadyActed    = errors.New("already acted this night")
	ErrSameGuardTarget = errors.New("the guard cannot protect the same player two nights in a row")

	// Day (rules §4).
	ErrSelfNomination = errors.New("a player cannot nominate themselves")
	ErrAlreadySpared  = errors.New("this player was already spared today")
	ErrOnTrial        = errors.New("the player on trial cannot vote")
)
