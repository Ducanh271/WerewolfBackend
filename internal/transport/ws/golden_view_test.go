package ws

import (
	"encoding/json"
	"testing"
	"time"

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

func TestGolden_Welcome(t *testing.T) {
	tests := []struct {
		name  string
		creds *room.Credentials
		want  string
	}{
		{"join", &room.Credentials{Token: "tok", PIN: "4821"}, `{"type": "welcome", "protocolVersion": 1, "player": "p3", "token": "tok", "pin": "4821"}`},
		{"rejoin by token", nil, `{"type": "welcome", "protocolVersion": 1, "player": "p3"}`},
	}
	for _, tc := range tests {
		assertJSON(t, tc.name, welcome("p3", tc.creds), tc.want)
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

func TestGolden_ViewNoticeTexts(t *testing.T) {
	notCursed, inLove := "Bạn không bị nguyền.", "Bạn yêu Chi, vai Dân làng."
	msg := viewFrame(t, room.Message{View: &room.ViewMsg{View: engine.View{
		Phase: engine.PhaseNight,
		Me: &engine.MyView{
			ID:   "p1",
			Turn: &engine.TurnView{Step: "Werewolf", Doomed: true, Notice: "Bạn đã trúng độc và chết từ lúc này, bạn không được vote."},
			PhoneCheck: &engine.PhoneCheckView{
				CursedText: &notCursed,
				Love:       engine.LoveNotice{InLove: true, Partner: "p3", PartnerRole: "Villager", Colour: engine.LinkWhite},
				LoveText:   &inLove,
			},
		},
		Pack: &engine.PackView{Members: []engine.PlayerID{"p1", "p2"}, Doomed: []engine.PlayerID{"p1"}, DoomedNotices: []string{"An không thể vote đêm nay."}},
	}}})
	var view map[string]json.RawMessage
	if err := json.Unmarshal(msg["view"], &view); err != nil {
		t.Fatal(err)
	}
	var me map[string]json.RawMessage
	if err := json.Unmarshal(view["me"], &me); err != nil {
		t.Fatal(err)
	}

	assertJSON(t, "me.turn", me["turn"], `{"step": "Werewolf", "blocked": false, "doomed": true, "grief": false, "biteShown": false,
		"notice": "Bạn đã trúng độc và chết từ lúc này, bạn không được vote."}`)
	assertJSON(t, "me.phoneCheck", me["phoneCheck"], `{"cursed": false, "cursedText": "Bạn không bị nguyền.",
		"love": {"inLove": true, "partner": "p3", "partnerRole": "Villager", "colour": "white"},
		"loveText": "Bạn yêu Chi, vai Dân làng."}`)
	assertJSON(t, "pack", view["pack"], `{"members": ["p1", "p2"], "doomed": ["p1"], "doomedNotices": ["An không thể vote đêm nay."]}`)
}

// Without the Cursed wolf and Cupid in the role list, the phone check notice
// keeps exactly its old keys.
func TestGolden_ViewPhoneCheck_NoTextLinesWhenTheRolesAreNotInTheGame(t *testing.T) {
	tree := toTree(engine.PhoneCheckView{})
	b, _ := json.Marshal(tree)
	assertJSON(t, "phoneCheck", b, `{"cursed": false, "love": {"inLove": false, "partner": "", "partnerRole": "", "colour": ""}}`)
}

// SoundConfigured names the mode with the string configureSound takes.
func TestGolden_SoundConfigured_ModeIsItsName(t *testing.T) {
	tests := []struct {
		mode engine.SoundMode
		want string
	}{
		{engine.OneSpeaker, `{"mode": "oneSpeaker"}`},
		{engine.AllPhonesSpeak, `{"mode": "allPhones"}`},
	}
	for _, tc := range tests {
		b, _ := json.Marshal(toTree(engine.SoundConfigured{Mode: tc.mode}))
		assertJSON(t, tc.mode.String(), b, tc.want)
	}
}

// The lobby (and end) view carries every timer in ms and the sound mode.
func TestGolden_ViewTimersAndSound(t *testing.T) {
	timers, sound := engine.Timers{
		RoleView: 15 * time.Second, ChiefElection: 60 * time.Second, NightStep: 20 * time.Second,
		PhoneCheck: 10 * time.Second, LoversMeet: 10 * time.Second, Announce: 10 * time.Second,
		Day: 7 * time.Minute, NominationLock: 10 * time.Second, Defense: 60 * time.Second,
		Verdict: 20 * time.Second, HunterShot: 15 * time.Second, ChiefSuccession: 20 * time.Second,
	}, engine.AllPhonesSpeak
	msg := viewFrame(t, room.Message{View: &room.ViewMsg{View: engine.View{Phase: engine.PhaseLobby, Timers: &timers, Sound: &sound}}})
	var view map[string]json.RawMessage
	if err := json.Unmarshal(msg["view"], &view); err != nil {
		t.Fatal(err)
	}

	assertJSON(t, "timers", view["timers"], `{"roleView": 15000, "chiefElection": 60000, "nightStep": 20000,
		"phoneCheck": 10000, "loversMeet": 10000, "announce": 10000, "day": 420000, "nominationLock": 10000,
		"defense": 60000, "verdict": 20000, "hunterShot": 15000, "chiefSuccession": 20000}`)
	assertJSON(t, "sound", view["sound"], `"allPhones"`)
}
