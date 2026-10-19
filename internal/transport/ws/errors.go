package ws

import (
	"errors"
	"time"

	"masoi/internal/engine"
	"masoi/internal/room"
)

// Codes of errors the transport itself raises.
const (
	codeBadMessage         = "bad_message"
	codeUnknownType        = "unknown_type"
	codeUnknownCommand     = "unknown_command"
	codeTooLarge           = "message_too_large"
	codeRateLimited        = "rate_limited"
	codeNotJoined          = "not_joined"
	codeAlreadyJoined      = "already_joined"
	codeInternal           = "internal"
	codeSecretInURL        = "secret_in_url"
	codeBadOrigin          = "bad_origin"
	codeTimerOutOfRange    = "timer_out_of_range" // engine.ErrTimerOutOfRange, raised here when a timer overflows
	codeTooManyConnections = "too_many_connections"
)

// protoError is a refusal the transport makes before the room hears of the
// message.
type protoError struct{ code, message string }

func (e *protoError) Error() string { return e.message }

func badMessage(msg string) error { return &protoError{codeBadMessage, msg} }

// errorMsg is an outgoing `error` message.
type errorMsg struct {
	Code       string
	Message    string
	RetryAfter time.Duration
}

// toErrorMsg turns any error into what a client is told. Errors without a
// code are internal: the client gets no detail.
func toErrorMsg(err error) errorMsg {
	var pe *protoError
	if errors.As(err, &pe) {
		return errorMsg{Code: pe.code, Message: pe.message}
	}
	var locked room.PINLockedError
	if errors.As(err, &locked) {
		return errorMsg{Code: "pin_locked", Message: room.ErrPINLocked.Error(), RetryAfter: locked.RetryAfter}
	}
	if errors.Is(err, room.ErrBadCredentials) {
		// Never says more than "bad credentials", whatever was wrapped.
		return errorMsg{Code: "bad_credentials", Message: room.ErrBadCredentials.Error()}
	}
	// Every engine and room error a client may see carries its own code
	// (engine.NewError).
	if code, ok := engine.ErrorCode(err); ok {
		return errorMsg{Code: code, Message: err.Error()}
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
