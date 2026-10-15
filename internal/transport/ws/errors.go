package ws

import (
	"errors"
	"time"

	"masoi/internal/engine"
	"masoi/internal/room"
)

// Codes of errors the transport itself raises.
const (
	codeBadMessage      = "bad_message"
	codeUnknownType     = "unknown_type"
	codeUnknownCommand  = "unknown_command"
	codeTooLarge        = "message_too_large"
	codeRateLimited     = "rate_limited"
	codeNotJoined       = "not_joined"
	codeAlreadyJoined   = "already_joined"
	codeInternal        = "internal"
	codeSecretInURL     = "secret_in_url"
	codeBadOrigin       = "bad_origin"
	codeTimerOutOfRange = "timer_out_of_range"
)

// protoError is a refusal the transport makes before the room hears of the
// message.
type protoError struct{ code, message string }

func (e *protoError) Error() string { return e.message }

func badMessage(msg string) error { return &protoError{codeBadMessage, msg} }

// codeTable names every error that may reach a client. The code is the
// snake_case name of the Go error (ErrNotYourTurn -> not_your_turn).
var codeTable = []struct {
	err  error
	code string
}{
	{engine.ErrUnknownCommand, "unknown_command"},
	{engine.ErrWrongPhase, "wrong_phase"},
	{engine.ErrInvalidPlayerID, "invalid_player_id"},
	{engine.ErrInvalidNickname, "invalid_nickname"},
	{engine.ErrNicknameTaken, "nickname_taken"},
	{engine.ErrAlreadyJoined, "already_joined"},
	{engine.ErrRoomFull, "room_full"},
	{engine.ErrUnknownPlayer, "unknown_player"},
	{engine.ErrSeatsMismatch, "seats_mismatch"},
	{engine.ErrDuplicateSeat, "duplicate_seat"},
	{engine.ErrTimerOutOfRange, codeTimerOutOfRange},
	{engine.ErrModeUnsupported, "mode_unsupported"},
	{engine.ErrRolesNotConfigured, "roles_not_configured"},
	{engine.ErrSeatsNotSet, "seats_not_set"},
	{engine.ErrAlreadyDealt, "already_dealt"},
	{engine.ErrTooFewPlayers, "too_few_players"},
	{engine.ErrTooManyPlayers, "too_many_players"},
	{engine.ErrUnknownRole, "unknown_role"},
	{engine.ErrInvalidRoleCount, "invalid_role_count"},
	{engine.ErrRoleMustBeUnique, "role_must_be_unique"},
	{engine.ErrRoleCountMismatch, "role_count_mismatch"},
	{engine.ErrNoWolves, "no_wolves"},
	{engine.ErrTooManyWolves, "too_many_wolves"},
	{engine.ErrRoleNeedsMorePlayers, "role_needs_more_players"},
	{engine.ErrNotYourTurn, "not_your_turn"},
	{engine.ErrNotAlive, "not_alive"},
	{engine.ErrInvalidTarget, "invalid_target"},
	{engine.ErrAlreadyActed, "already_acted"},
	{engine.ErrSameGuardTarget, "same_guard_target"},
	{engine.ErrVisited, "visited"},
	{engine.ErrPotionUsed, "potion_used"},
	{engine.ErrNothingToHeal, "nothing_to_heal"},
	{engine.ErrSelfNomination, "self_nomination"},
	{engine.ErrAlreadySpared, "already_spared"},
	{engine.ErrOnTrial, "on_trial"},

	{room.ErrNotHost, "not_host"},
	{room.ErrForbidden, "forbidden"},
	{room.ErrDisconnected, "disconnected"},
	{room.ErrClosed, "room_closed"},
	{room.ErrBadCredentials, "bad_credentials"},
	{room.ErrPINLocked, "pin_locked"},
	{room.ErrRoomNotFound, "room_not_found"},
	{room.ErrTooManyRooms, "too_many_rooms"},
}

// errorMsg is an outgoing `error` message.
type errorMsg struct {
	Code       string
	Message    string
	RetryAfter time.Duration
}

// toErrorMsg turns any error into what a client is told. Errors the table does
// not know are internal: the client gets no detail.
func toErrorMsg(err error) errorMsg {
	var pe *protoError
	if errors.As(err, &pe) {
		return errorMsg{Code: pe.code, Message: pe.message}
	}
	var locked room.PINLockedError
	if errors.As(err, &locked) {
		return errorMsg{Code: "pin_locked", Message: room.ErrPINLocked.Error(), RetryAfter: locked.RetryAfter}
	}
	for _, e := range codeTable {
		if errors.Is(err, e.err) {
			if e.err == room.ErrBadCredentials {
				return errorMsg{Code: e.code, Message: e.err.Error()}
			}
			return errorMsg{Code: e.code, Message: err.Error()}
		}
	}
	return errorMsg{Code: codeInternal, Message: "internal error"}
}

func (e errorMsg) json() map[string]any {
	m := map[string]any{"type": "error", "code": e.Code, "message": e.Message}
	if e.RetryAfter > 0 {
		m["retryAfter"] = e.RetryAfter.Milliseconds()
	}
	return m
}

// isInternal says whether err was not one the client should know about.
func isInternal(err error) bool { return toErrorMsg(err).Code == codeInternal }
