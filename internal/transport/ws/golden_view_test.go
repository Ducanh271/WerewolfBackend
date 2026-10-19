package ws

import (
	"encoding/json"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/room"
)

// The golden view tests pin the parts of the `view` message added in T30e
// (docs/protocol.md): new keys only, next to the old ones.

// viewFrame encodes m and returns its `view` message as raw JSON objects.
func viewFrame(t *testing.T, m room.Message) map[string]json.RawMessage {
	t.Helper()
	out, err := frames(m)
	if err != nil || len(out) == 0 {
		t.Fatalf("frames = %v, %v", out, err)
	}
	var msg map[string]json.RawMessage
	if err := json.Unmarshal(out[0], &msg); err != nil {
		t.Fatal(err)
	}
	return msg
}

func assertJSON(t *testing.T, what string, got json.RawMessage, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("%s: %v", what, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("%s: bad want: %v", what, err)
	}
	gb, _ := json.Marshal(g)
	wb, _ := json.Marshal(w)
	if string(gb) != string(wb) {
		t.Errorf("%s =\n  %s\nwant\n  %s", what, gb, wb)
	}
}

func TestGolden_ViewPeople(t *testing.T) {
	msg := viewFrame(t, room.Message{View: &room.ViewMsg{
		Speaker: true,
		People: []room.Person{
			{ID: "p1", Nickname: "An", Connected: true, IsHost: true},
			{ID: "p2", Nickname: "Bình"},
		},
		View: engine.View{Phase: engine.PhaseLobby},
	}})

	assertJSON(t, "people", msg["people"], `[
		{"id": "p1", "nickname": "An", "connected": true, "isHost": true},
		{"id": "p2", "nickname": "Bình", "connected": false, "isHost": false}
	]`)
	assertJSON(t, "speaker", msg["speaker"], `true`)
	assertJSON(t, "type", msg["type"], `"view"`)
	if _, ok := msg["view"]; !ok {
		t.Error("the view itself is missing")
	}
}

func TestGolden_ViewPeople_AbsentDuringTheGame(t *testing.T) {
	msg := viewFrame(t, room.Message{View: &room.ViewMsg{View: engine.View{Phase: engine.PhaseNight}}})
	if got, ok := msg["people"]; ok {
		t.Errorf("people = %s, want no key during the game", got)
	}
}
