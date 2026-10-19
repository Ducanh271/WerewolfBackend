package ws

import (
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

// A phone replaced by a later rejoin, or kicked, is told "disconnected" with
// the reason as the message.
func TestErrors_DisconnectReasonsKeepTheDisconnectedCode(t *testing.T) {
	for _, reason := range []error{room.ReasonAnotherDevice, room.ReasonPINTakenOver, room.ReasonKicked} {
		if m := toErrorMsg(reason); m.Code != "disconnected" || m.Message != reason.Error() {
			t.Errorf("%v: got %+v, want code disconnected and the reason", reason, m)
		}
	}
}
