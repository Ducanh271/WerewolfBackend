package ws

import (
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

// TestLobby_InvalidNicknameIsRejected is audit S4 end to end: the limits
// live in the engine (internal/engine/nickname.go), this only checks they
// reach the client with the right error code.
func TestLobby_InvalidNicknameIsRejected(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()

	rows := []struct {
		name     string
		nickname string
	}{
		{"too long", "123456789012345678901"},
		{"look-alike letter", "Аn"},    // Cyrillic А, not Latin A
		{"invisible character", "An​"}, // zero-width space
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			c := e.dial(nil)
			e.send(c, map[string]any{"type": "join", "room": code, "nickname": row.nickname})
			if got := e.waitError(c, 1); got != "invalid_nickname" {
				t.Fatalf("code = %q, want invalid_nickname", got)
			}
		})
	}
}

// TestLobby_KickPlayer is audit K2 end to end: the host removes a stranger
// from the lobby, who is told why and disconnected, without disturbing the
// seating or role list (decided when T27a was planned).
func TestLobby_KickPlayer(t *testing.T) {
	e := newEnv(t, nil)
	code, hostKey := e.createRoomWithHost()
	host := e.joinHost(code, hostKey, "An")
	stranger := e.join(code, "Binh")
	strangerID := stranger.Snapshot().Welcome.Str("player")

	e.send(host, map[string]any{"type": "kick", "target": strangerID})

	if !stranger.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Closed }) {
		t.Fatalf("the kicked phone was never closed: %v", stranger.Snapshot())
	}
	if got := stranger.Snapshot().Errors[len(stranger.Snapshot().Errors)-1].Str("code"); got != "disconnected" {
		t.Fatalf("code = %q, want disconnected", got)
	}
}

func TestLobby_KickPlayer_OnlyTheHostMay(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	other := e.join(code, "An")
	target := e.join(code, "Binh")
	targetID := target.Snapshot().Welcome.Str("player")

	e.send(other, map[string]any{"type": "kick", "target": targetID})
	if got := e.waitError(other, 1); got != "not_host" {
		t.Fatalf("code = %q, want not_host", got)
	}
}

// TestLobby_TransferHost is the T27a addition: the host hands the title to
// anyone in the room, no consent needed.
func TestLobby_TransferHost(t *testing.T) {
	e := newEnv(t, nil)
	code, hostKey := e.createRoomWithHost()
	host := e.joinHost(code, hostKey, "An")
	other := e.join(code, "Binh")
	otherID := other.Snapshot().Welcome.Str("player")

	e.send(host, map[string]any{"type": "transferHost", "target": otherID})

	e.send(host, cmd("setSeats", map[string]any{"order": []string{"p1", "p2"}}))
	if got := e.waitError(host, 1); got != "not_host" {
		t.Fatalf("the old host must lose the powers: code = %q, want not_host", got)
	}
	e.send(other, cmd("setSeats", map[string]any{"order": []string{"p1", "p2"}}))
	if len(other.Snapshot().Errors) != 0 {
		t.Fatalf("the new host could not set seats: %v", other.Snapshot().Errors)
	}
}
