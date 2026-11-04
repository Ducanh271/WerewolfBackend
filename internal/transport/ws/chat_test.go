package ws

import (
	"strings"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/room"
	"masoi/internal/transport/ws/wsclient"
)

// chatTable is five phones in the Chief election, when everybody may write.
func chatTable(t *testing.T) (*env, *table, string) {
	t.Helper()
	e := newEnv(t, func(c *Config) { c.Burst = 100 })
	code, hostKey := e.createRoomWithHost()
	tb := &table{t: t, e: e}
	for _, nick := range []string{"An", "Binh", "Chi", "Dung", "Em"} {
		var c *wsclient.Client
		if nick == "An" {
			c = e.joinHost(code, hostKey, nick)
		} else {
			c = e.join(code, nick)
		}
		tb.p = append(tb.p, &phone{c: c, id: c.Snapshot().Welcome.Str("player"), nick: nick})
	}
	order := make([]string, len(tb.p))
	for i, p := range tb.p {
		order[i] = p.id
	}
	host := tb.host()
	tb.do(host, "setSeats", map[string]any{"order": order})
	tb.do(host, "configureRoles", map[string]any{"roles": map[string]int{"Werewolf": 1, "Guard": 1, "Seer": 1, "Villager": 2}})
	tb.do(host, "dealRoles", nil)
	tb.waitPhase("RoleView")
	tb.advance(engine.DefaultSettings().RoleView)
	tb.waitPhase("ChiefElection")
	return e, tb, code
}

func chatMsg(text string) map[string]any { return cmd("sendChat", map[string]any{"text": text}) }

// chatLines is the phone's chat as "player: text".
func chatLines(c *wsclient.Client) []string {
	var out []string
	for _, ev := range c.Snapshot().Events {
		if ev.Str("type") == "ChatSent" {
			out = append(out, ev.Str("data", "by")+": "+ev.Str("data", "text"))
		}
	}
	return out
}

func waitChat(t *testing.T, c *wsclient.Client, want []string) {
	t.Helper()
	ok := c.Wait(ctx(t), func(wsclient.Snapshot) bool { return len(chatLines(c)) >= len(want) })
	got := chatLines(c)
	if !ok || strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the chat reads %q, want %q", got, want)
	}
}

func TestE2E_Chat_EveryPhoneGetsItInOrderAndARejoiningPhoneGetsTheHistory(t *testing.T) {
	e, tb, code := chatTable(t)
	an, binh, chi := tb.p[0], tb.p[1], tb.p[2]

	e.send(binh.c, chatMsg("  Tôi nghi\nChi "))
	waitChat(t, binh.c, []string{binh.id + ": Tôi nghi Chi"})
	// Each message waits for the one before: two sockets have no order of their own.
	e.send(chi.c, chatMsg("Không phải tôi"))
	waitChatCount(t, an.c, 2)
	e.send(an.c, chatMsg("Bầu Bình đi"))
	want := []string{binh.id + ": Tôi nghi Chi", chi.id + ": Không phải tôi", an.id + ": Bầu Bình đi"}
	for _, p := range tb.p {
		waitChat(t, p.c, want)
	}

	// Nothing but the sender and the text, and the speaker has nothing to read.
	for _, ev := range an.c.Snapshot().Events {
		if ev.Str("type") == "ChatSent" && ev.Str("say") != "" {
			t.Fatalf("the speaker would read a message aloud: %v", ev)
		}
	}

	// A phone that comes back gets the history, then what follows.
	token := chi.c.Snapshot().Welcome.Str("token")
	chi.c.Close()
	<-chi.c.Done()
	e.send(binh.c, chatMsg("Chi đi đâu rồi?"))
	want = append(want, binh.id+": Chi đi đâu rồi?")
	waitChat(t, an.c, want)

	back := e.dial(nil)
	e.send(back, map[string]any{"type": "rejoin", "room": code, "token": token})
	waitChat(t, back, want)
	e.send(an.c, chatMsg("Chi về rồi"))
	waitChat(t, back, append(want, an.id+": Chi về rồi"))
}

func TestChat_Refusals_LeaveTheSocketUsable(t *testing.T) {
	big := `{"type":"command","name":"sendChat","args":{"text":"` + strings.Repeat("x", DefaultMaxMessageBytes) + `"}}`
	tests := []struct {
		name string
		msg  any
		code string
	}{
		{"a message over the size limit", big, "message_too_large"},
		{"over 200 characters", chatMsg(strings.Repeat("ữ", engine.MaxChatLen+1)), "chat_too_long"},
		{"nothing to read", chatMsg(" \n​ "), "chat_empty"},
		{"no text at all", cmd("sendChat", map[string]any{}), "chat_empty"},
		{"text of the wrong type", cmd("sendChat", map[string]any{"text": 5}), "bad_message"},
		{"writing as someone else", cmd("sendChat", map[string]any{"text": "chào", "by": "p1"}), "bad_message"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e, tb, _ := chatTable(t)
			binh := tb.p[1]
			before := len(binh.c.Snapshot().Errors)
			if raw, ok := tc.msg.(string); ok {
				if err := binh.c.SendRaw(ctx(t), []byte(raw)); err != nil {
					t.Fatal(err)
				}
			} else {
				e.send(binh.c, tc.msg)
			}
			if got := e.waitError(binh.c, before+1); got != tc.code {
				t.Fatalf("code = %q, want %q", got, tc.code)
			}
			e.send(binh.c, chatMsg("vẫn nhắn được"))
			waitChat(t, tb.p[0].c, []string{binh.id + ": vẫn nhắn được"})
		})
	}
}

// The longest message allowed fits in a message however the phone encodes it.
func TestChat_TheLongestMessageFitsTheSizeLimit(t *testing.T) {
	_, tb, _ := chatTable(t)
	binh := tb.p[1]
	emoji := strings.Repeat(`🐺`, engine.MaxChatLen) // JSON's longest form of a character
	raw := `{"type":"command","name":"sendChat","args":{"text":"` + emoji + `"}}`
	if len(raw) > DefaultMaxMessageBytes {
		t.Fatalf("%d bytes is over the %d limit", len(raw), DefaultMaxMessageBytes)
	}
	if err := binh.c.SendRaw(ctx(t), []byte(raw)); err != nil {
		t.Fatal(err)
	}
	waitChat(t, tb.p[0].c, []string{binh.id + ": " + strings.Repeat("🐺", engine.MaxChatLen)})
}

func TestChat_OverTheLimit_IsRefusedUntilTheClockAllowsIt(t *testing.T) {
	e, tb, _ := chatTable(t)
	binh := tb.p[1]
	for range room.ChatBurst {
		e.send(binh.c, chatMsg("chào"))
	}
	waitChatCount(t, tb.p[0].c, room.ChatBurst)
	before := len(binh.c.Snapshot().Errors)
	e.send(binh.c, chatMsg("chào"))
	if got := e.waitError(binh.c, before+1); got != "chat_too_fast" {
		t.Fatalf("code = %q, want chat_too_fast", got)
	}
	e.clk.Advance(room.ChatInterval)
	e.send(binh.c, chatMsg("chào"))
	waitChatCount(t, tb.p[0].c, room.ChatBurst+1)
}

func waitChatCount(t *testing.T, c *wsclient.Client, n int) {
	t.Helper()
	if !c.Wait(ctx(t), func(wsclient.Snapshot) bool { return len(chatLines(c)) >= n }) {
		t.Fatalf("the chat has %d messages, want %d", len(chatLines(c)), n)
	}
}
