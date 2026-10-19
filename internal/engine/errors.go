package engine

import "errors"

// All errors returned by Decide. Player input never panics: it comes back as
// one of these, wrapped with %w when the message carries detail. Each carries
// its wire code (docs/protocol.md "Errors"), the snake_case of its name.
var (
	// ErrUnknownCommand means no part of the engine and no role claimed the
	// command.
	ErrUnknownCommand = NewError("unknown_command", "unknown command")
	// ErrWrongPhase means the command is not allowed in the current phase.
	ErrWrongPhase = NewError("wrong_phase", "command not allowed in this phase")

	// Lobby.
	ErrInvalidPlayerID = NewError("invalid_player_id", "invalid player id")
	ErrInvalidNickname = NewError("invalid_nickname", "invalid nickname")
	ErrNicknameTaken   = NewError("nickname_taken", "nickname already taken")
	ErrAlreadyJoined   = NewError("already_joined", "player already joined")
	ErrRoomFull        = NewError("room_full", "room is full")
	ErrUnknownPlayer   = NewError("unknown_player", "unknown player")
	ErrSeatsMismatch   = NewError("seats_mismatch", "seating must list every player exactly once")
	ErrDuplicateSeat   = NewError("duplicate_seat", "player seated twice")
	// ErrTimerOutOfRange means a phase time is outside its allowed range.
	ErrTimerOutOfRange = NewError("timer_out_of_range", "timer out of range")
	// ErrModeUnsupported means the game mode exists in the model but is not
	// built yet (Online).
	ErrModeUnsupported = NewError("mode_unsupported", "game mode not supported yet")
	// ErrInvalidSoundMode means ConfigureSound was sent an unknown mode.
	ErrInvalidSoundMode = NewError("invalid_sound_mode", "invalid sound mode")

	// Dealing (rules §2).
	ErrRolesNotConfigured = NewError("roles_not_configured", "roles are not configured")
	ErrSeatsNotSet        = NewError("seats_not_set", "seats must be set for every player before dealing")
	ErrAlreadyDealt       = NewError("already_dealt", "roles were already dealt")

	// Role configuration (rules §2).
	ErrTooFewPlayers        = NewError("too_few_players", "too few players")
	ErrTooManyPlayers       = NewError("too_many_players", "too many players")
	ErrUnknownRole          = NewError("unknown_role", "unknown role")
	ErrInvalidRoleCount     = NewError("invalid_role_count", "invalid role count")
	ErrRoleMustBeUnique     = NewError("role_must_be_unique", "role can only be dealt to one player")
	ErrRoleCountMismatch    = NewError("role_count_mismatch", "role slots must match the number of players")
	ErrNoWolves             = NewError("no_wolves", "the game needs at least one wolf")
	ErrTooManyWolves        = NewError("too_many_wolves", "too many wolves for this number of players")
	ErrRoleNeedsMorePlayers = NewError("role_needs_more_players", "role needs more players")
	// ErrRoleRequired means the role list is missing a role that every game
	// must include at least one copy of (rules §2: at least one Villager).
	// It has no wire code yet, so a client is told "internal".
	ErrRoleRequired = errors.New("role list is missing a required role")

	// Night actions (rules §3, §5, §6).
	ErrNotYourTurn       = NewError("not_your_turn", "not your turn")
	ErrNotAlive          = NewError("not_alive", "player is dead")
	ErrInvalidTarget     = NewError("invalid_target", "invalid target")
	ErrAlreadyActed      = NewError("already_acted", "already acted this night")
	ErrBlocked           = NewError("blocked", "blocked tonight: no night ability")
	ErrPotionUsed        = NewError("potion_used", "this potion is already used up")
	ErrAbilityUsed       = NewError("ability_used", "this once-a-game ability is already used")
	ErrNothingToHeal     = NewError("nothing_to_heal", "nobody is bitten tonight")
	ErrSameGuardTarget   = NewError("same_guard_target", "the guard cannot protect the same player two nights in a row")
	ErrSameVisitorTarget = NewError("same_visitor_target", "the Visitor cannot visit the same player two nights in a row")

	// Day (rules §4).
	ErrSelfNomination = NewError("self_nomination", "a player cannot nominate themselves")
	ErrAlreadySpared  = NewError("already_spared", "this player was already spared today")
	ErrOnTrial        = NewError("on_trial", "the player on trial cannot vote")
)

// codedError is an error with its wire code.
type codedError struct{ code, msg string }

func (e *codedError) Error() string { return e.msg }

// Code is the error's code on the wire, e.g. "not_your_turn".
func (e *codedError) Code() string { return e.code }

// NewError makes an error a client may be told about: the transport sends
// code and the error text, so a new error needs no table outside its own
// declaration. Roles declare their own errors with it, in their own file.
func NewError(code, msg string) error { return &codedError{code: code, msg: msg} }

// ErrorCode is the wire code of err or of an error it wraps; ok is false when
// none has one (the client is then told "internal").
func ErrorCode(err error) (code string, ok bool) {
	var c interface{ Code() string }
	if errors.As(err, &c) {
		return c.Code(), true
	}
	return "", false
}
