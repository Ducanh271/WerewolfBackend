package ws

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"masoi/internal/room"
	"masoi/internal/transport/ws/wsclient"
)

func (e *env) do(method, path, body string) (int, map[string]any) {
	e.t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, strings.NewReader(body))
	if err != nil {
		e.t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func TestRooms_CreateReturnsAFourCharacterCode(t *testing.T) {
	e := newEnv(t, nil)
	status, body := e.do("POST", "/rooms", "")
	code, _ := body["room"].(string)
	if status != http.StatusCreated || len(code) != room.CodeLength || strings.Trim(code, room.CodeAlphabet) != "" {
		t.Fatalf("POST /rooms = %d %v", status, body)
	}
	if status, _ = e.do("POST", "/rooms", `{"mode":"inPerson"}`); status != http.StatusCreated {
		t.Fatalf("explicit inPerson: status %d", status)
	}
}

func TestRooms_OnlineModeIsRefused(t *testing.T) {
	e := newEnv(t, nil)
	status, body := e.do("POST", "/rooms", `{"mode":"online"}`)
	if status != http.StatusBadRequest || body["code"] != "mode_unsupported" {
		t.Fatalf("POST /rooms online = %d %v", status, body)
	}
	for _, bad := range []string{`{"mode":"cards"}`, `{"mode":"inPerson","seed":1}`, `nope`} {
		if status, body = e.do("POST", "/rooms", bad); status != http.StatusBadRequest || body["code"] != "bad_message" {
			t.Errorf("POST /rooms %s = %d %v", bad, status, body)
		}
	}
}

func TestRooms_JoinLinkTellsWhetherTheRoomExists(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	if status, body := e.do("GET", "/join/"+code, ""); status != http.StatusOK || body["room"] != code {
		t.Fatalf("GET /join/%s = %d %v", code, status, body)
	}
	if status, body := e.do("GET", "/join/"+strings.ToLower(code), ""); status != http.StatusOK || body["room"] != code {
		t.Fatalf("lower-case code = %d %v", status, body)
	}
	if status, body := e.do("GET", "/join/ZZZZ", ""); status != http.StatusNotFound || body["code"] != "room_not_found" {
		t.Fatalf("unknown room = %d %v", status, body)
	}
}

func TestRooms_JoiningAnUnknownRoomIsAnError(t *testing.T) {
	e := newEnv(t, nil)
	c := e.dial(nil)
	e.send(c, map[string]any{"type": "join", "room": "ZZZZ", "nickname": "An"})
	if got := e.waitError(c, 1); got != "room_not_found" {
		t.Fatalf("code = %q", got)
	}
}

func TestRooms_OnlyTheHostPhoneIsTheSpeaker(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	host := e.join(code, "An")
	other := e.join(code, "Binh")
	two := func(s wsclient.Snapshot) bool { return len(s.Events) == 2 }
	host.Wait(ctx(t), two)
	other.Wait(ctx(t), two)
	if host.Snapshot().View["speaker"] != true {
		t.Errorf("the host phone is not the speaker: %v", host.Snapshot().View)
	}
	if _, has := other.Snapshot().View["speaker"]; has {
		t.Errorf("a second phone is told to speak: %v", other.Snapshot().View)
	}
}
