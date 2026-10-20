package ws

import (
	"errors"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/room"
)

// A player who does not hold the role being called gets engine.ErrNotYourTurn
// with nothing added (internal/roles/not_your_turn_test.go). On the phone it
// must be one fixed message, so the screen tells nothing either.
func TestErrors_NotYourTurn_IsOneFixedCodeAndText(t *testing.T) {
	got := toErrorMsg(engine.ErrNotYourTurn).json()
	want := map[string]any{"type": "error", "code": "not_your_turn", "message": "not your turn"}
	if len(got) != len(want) {
		t.Fatalf("error message %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("error message %v, want %v", got, want)
		}
	}
}

// A command that hit a bug in the room is "internal", with no detail.
func TestErrors_RoomPanic_IsInternal(t *testing.T) {
	m := toErrorMsg(room.ErrInternal)
	if m.Code != codeInternal || m.Message != "internal error" {
		t.Fatalf("got %+v, want code %q and no detail", m, codeInternal)
	}
	if !isInternal(room.ErrInternal) {
		t.Fatal("a room panic must be logged as an internal error")
	}
}

// A phone replaced by a later rejoin, or kicked, is told why with its own
// code, and the reason still closes the connection like ErrDisconnected.
func TestErrors_DisconnectReasonsHaveTheirOwnCodes(t *testing.T) {
	for _, tc := range []struct {
		reason error
		code   string
	}{
		{room.ReasonAnotherDevice, "replaced_by_other_device"},
		{room.ReasonPINTakenOver, "replaced_by_other_device"},
		{room.ReasonKicked, "kicked_by_host"},
	} {
		if m := toErrorMsg(tc.reason); m.Code != tc.code || m.Message != tc.reason.Error() {
			t.Errorf("%v: got %+v, want code %s and the reason", tc.reason, m, tc.code)
		}
		if !errors.Is(tc.reason, room.ErrDisconnected) {
			t.Errorf("%v: must still match ErrDisconnected", tc.reason)
		}
	}
}
