package ws

import "testing"

func TestDecode_UnknownTypesAndUnknownFieldsAreRejected(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	c := e.join(code, "An")

	rows := []struct {
		name string
		msg  any
		code string
	}{
		{"unknown message type", map[string]any{"type": "teleport"}, "unknown_type"},
		{"no type", map[string]any{"room": code}, "bad_message"},
		{"unknown command", cmd("openTheTable", nil), "unknown_command"},
		{"seed in dealRoles", cmd("dealRoles", map[string]any{"seed": 1}), "bad_message"},
		{"seed in another command", cmd("nominate", map[string]any{"target": "p2", "seed": 7}), "bad_message"},
		{"voter in args", cmd("nominate", map[string]any{"target": "p2", "voter": "p3"}), "bad_message"},
		{"by in args", cmd("wolfVote", map[string]any{"target": "p2", "by": "p3"}), "bad_message"},
		{"seed beside the command", map[string]any{"type": "command", "name": "dealRoles", "seed": 1}, "bad_message"},
		{"seed in leave", map[string]any{"type": "leave", "seed": 1}, "bad_message"},
		{"wrong capitals", cmd("nominate", map[string]any{"Target": "p2"}), "bad_message"},
		{"unknown timer", cmd("configureTimers", map[string]any{"timers": map[string]any{"coffeeBreak": 5000}}), "bad_message"},
		{"timer out of range", cmd("configureTimers", map[string]any{"timers": map[string]any{"day": -5}}), "timer_out_of_range"},
		{"wrong field type", cmd("skipDay", map[string]any{"on": "yes"}), "bad_message"},
		{"timeout is not a command", cmd("phaseTimeout", nil), "unknown_command"},
	}
	for i, row := range rows {
		e.send(c, row.msg)
		if got := e.waitError(c, i+1); got != row.code {
			t.Errorf("%s: code = %q, want %q", row.name, got, row.code)
		}
	}
}

func TestDecode_NotJSONAndTrailingDataAreRejected(t *testing.T) {
	e := newEnv(t, nil)
	c := e.join(e.createRoom(), "An")
	for i, raw := range []string{`not json`, `[1,2]`, `{"type":"leave"} {"type":"leave"}`, `null`} {
		if err := c.SendRaw(ctx(t), []byte(raw)); err != nil {
			t.Fatal(err)
		}
		if got := e.waitError(c, i+1); got != "bad_message" {
			t.Errorf("%q: code = %q, want bad_message", raw, got)
		}
	}
}

func TestDecode_CommandsBeforeJoiningAreRefused(t *testing.T) {
	e := newEnv(t, nil)
	e.createRoom()
	c := e.dial(nil)
	e.send(c, cmd("dealRoles", nil))
	if got := e.waitError(c, 1); got != "not_joined" {
		t.Fatalf("code = %q, want not_joined", got)
	}
}

func TestDecode_ClientNeverChoosesWhoActs(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	host := e.join(code, "An")
	other := e.join(code, "Binh")
	// Only the host may set seats; a player cannot borrow the host's identity
	// because there is nowhere in a message to put one.
	e.send(other, cmd("setSeats", map[string]any{"order": []string{"p1", "p2"}}))
	if got := e.waitError(other, 1); got != "not_host" {
		t.Fatalf("code = %q, want not_host", got)
	}
	e.send(host, cmd("setSeats", map[string]any{"order": []string{"p1", "p2"}}))
	e.send(host, cmd("nominate", map[string]any{"target": "p2"})) // a marker: it fails in the lobby
	if got := e.waitError(host, 1); got != "wrong_phase" {
		t.Fatalf("first error of the host = %q, want only the marker's wrong_phase", got)
	}
}
