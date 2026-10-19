package ws

import (
	"encoding/json"
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

// exampleTable is who sits at the example table, in join and seat order.
var exampleTable = []string{"An", "Bình", "Chi", "Dũng", "Giang", "Hà", "Khoa", "Lan", "Minh", "Nga", "Phúc", "Quân"}

// exampleRoles is the role list of the example games: every role but the
// plain wolf, which would break the wolf limit with the Cursed wolf (rules §2).
var exampleRoles = map[string]int{
	"CursedWolf": 1, "WolfSeer": 1, "Villager": 2, "Seer": 1, "Guard": 1,
	"Witch": 1, "Hunter": 1, "Visitor": 1, "Detective": 1, "RustyKnight": 1, "Cupid": 1,
}

// playExampleGames plays the example games and returns their examples.
func playExampleGames(t *testing.T) *exampleSet {
	x := newExampleSet(t)
	g := &exGame{t: t, e: newExampleEnv(t), x: x}
	g.lobby()
	g.firstGame()
	g.secondGame()
	return x
}

// lobby: the HTTP endpoints, joining and rejoining, the host's setup.
func (g *exGame) lobby() {
	t := g.t
	status, body := g.httpJSON("POST", "/rooms", []byte(`{"mode":"inPerson"}`))
	if status != 201 {
		t.Fatalf("POST /rooms: %d %s", status, body)
	}
	g.recordHTTP("create-room.request", "CreateRoomRequest", []byte(`{"mode":"inPerson"}`))
	var created struct{ Room, HostKey string }
	mustUnmarshal(t, body, &created)
	g.x.redact[created.HostKey] = exampleHostKey
	g.recordHTTP("create-room.reply", "CreateRoomReply", body)
	if status, body = g.httpJSON("GET", "/join/k7qp", nil); status != 200 {
		t.Fatalf("GET /join: %d %s", status, body)
	}
	g.recordHTTP("join-room.reply", "JoinRoomReply", body)
	status, body = g.httpJSON("GET", "/join/ZZZZ", nil)
	if status != 404 {
		t.Fatalf("GET /join/ZZZZ: %d", status)
	}
	g.recordHTTP("join-room.error.room_not_found", "HTTPError", body)
	if status, body = g.httpJSON("GET", "/guide", nil); status != 200 {
		t.Fatalf("GET /guide: %d", status)
	}
	g.recordHTTP("guide.reply", "GuideReply", body)

	// The creator joins with the hostKey, then everybody else; Tú and Vy
	// will not stay.
	for _, nick := range append(append([]string{}, exampleTable...), "Tú", "Vy") {
		msg := map[string]any{"type": "join", "room": exampleRoom, "nickname": nick}
		example := ""
		switch nick {
		case "An":
			msg["hostKey"] = created.HostKey
			example = "join.host"
		case "Bình":
			example = "join"
		}
		g.join(nick, example, msg)
	}
	an, binh := g.phone("An"), g.phone("Bình")
	g.x.put("server/welcome.join.json", an.c.Snapshot().Welcome)
	g.x.put("server/events.reset.json", binh.c.Snapshot().Frames[2])
	g.x.put("server/events.batch.json", binh.lastFrame("events"))

	// Refused joins, each from a phone of its own.
	for _, tc := range []struct{ nick, code string }{{"an", "nickname_taken"}, {"An!", "invalid_nickname"}} {
		c := g.dial()
		g.send(c, "", map[string]any{"type": "join", "room": exampleRoom, "nickname": tc.nick})
		g.expectError(c, 0, tc.code, "error."+tc.code)
		c.Close()
	}
	// Messages the transport refuses before the room sees them.
	c := g.dial()
	g.e.send(c, map[string]any{"type": "hello"})
	g.expectError(c, 0, "unknown_type", "error.unknown_type")
	g.e.send(c, map[string]any{"type": "command", "name": "dealRoles"})
	g.expectError(c, 1, "not_joined", "error.not_joined")
	c.Close()
	n := len(binh.c.Snapshot().Errors)
	g.e.send(binh.c, map[string]any{"type": "command", "name": "fly"})
	g.expectError(binh.c, n, "unknown_command", "error.unknown_command")
	g.e.send(binh.c, map[string]any{"type": "command", "name": "nominate", "args": map[string]any{"target": 3}})
	g.expectError(binh.c, n+1, "bad_message", "error.bad_message")

	// Tú leaves; the host kicks Vy, whose phone is told why.
	tu, vy := g.phone("Tú"), g.phone("Vy")
	g.send(tu.c, "leave", map[string]any{"type": "leave"})
	g.drop(tu)
	n = len(binh.c.Snapshot().Errors)
	g.send(binh.c, "", map[string]any{"type": "kick", "target": vy.id})
	g.expectError(binh.c, n, "not_host", "error.not_host")
	g.send(an.c, "kick", map[string]any{"type": "kick", "target": vy.id})
	g.expectError(vy.c, len(vy.c.Snapshot().Errors), "disconnected", "error.disconnected.kicked")
	g.drop(vy)
	g.sync()

	// The host hands the title to Bình and Bình hands it back.
	g.send(an.c, "transferHost", map[string]any{"type": "transferHost", "target": binh.id})
	g.sync()
	g.send(binh.c, "", map[string]any{"type": "transferHost", "target": an.id})
	g.sync()

	g.rejoins()

	// Setup.
	order := make([]any, 0, len(g.phones))
	for _, p := range g.phones {
		order = append(order, p.id)
	}
	g.fails(an, "roles_not_configured", "dealRoles", nil)
	g.do(an, "setSeats", map[string]any{"order": order})
	g.fails(an, "too_many_wolves", "configureRoles", map[string]any{"roles": map[string]any{"Werewolf": 4, "Villager": 8}})
	g.do(an, "configureRoles", map[string]any{"roles": exampleRoles})
	g.do(an, "configureTimers", map[string]any{"timers": map[string]any{"day": 300000, "defense": 30000}})
	g.fails(an, "timer_out_of_range", "configureTimers", map[string]any{"timers": map[string]any{"verdict": 1000}})
	g.do(an, "configureSound", map[string]any{"mode": "oneSpeaker"})
	g.sync()
	g.shot(an, "lobby.host")
	g.shot(binh, "lobby")
}

func (g *exGame) join(nick, example string, msg map[string]any) *exPhone {
	g.t.Helper()
	c := g.dial()
	g.send(c, example, msg)
	if !c.Wait(ctx(g.t), func(s wsclient.Snapshot) bool { return s.Welcome != nil && s.View != nil }) {
		g.t.Fatalf("%s: no welcome; errors: %v", nick, c.Snapshot().Errors)
	}
	w := c.Snapshot().Welcome
	p := &exPhone{nick: nick, id: w.Str("player"), c: c, token: w.Str("token"), pin: w.Str("pin")}
	g.phones = append(g.phones, p)
	g.sync()
	return p
}

// rejoins: Chi's phone reloads (token), Dũng moves to another phone
// (nickname + PIN), someone guesses Giang's PIN until it locks.
func (g *exGame) rejoins() {
	t := g.t
	chi := g.phone("Chi")
	chi.c.Close()
	g.drop(chi)
	c := g.dial()
	g.send(c, "rejoin.token", map[string]any{"type": "rejoin", "room": exampleRoom, "token": chi.token})
	waitWelcome(t, c)
	g.x.put("server/welcome.rejoin-token.json", c.Snapshot().Welcome)
	chi.c = c
	g.phones = insertPhone(g.phones, chi)

	dung := g.phone("Dũng")
	old := dung.c
	nOld := len(old.Snapshot().Errors)
	c = g.dial()
	g.send(c, "", map[string]any{"type": "rejoin", "room": exampleRoom, "nickname": "Dũng", "pin": wrongPIN(dung.pin)})
	g.expectError(c, 0, "bad_credentials", "error.bad_credentials")
	g.send(c, "rejoin.pin", map[string]any{"type": "rejoin", "room": exampleRoom, "nickname": "Dũng", "pin": dung.pin})
	waitWelcome(t, c)
	g.x.put("server/welcome.rejoin-pin.json", c.Snapshot().Welcome)
	g.expectError(old, nOld, "disconnected", "error.disconnected.takeover")
	dung.c, dung.token = c, c.Snapshot().Welcome.Str("token")

	giang := g.phone("Giang")
	c = g.dial()
	for i := 0; i < 6; i++ {
		g.e.send(c, map[string]any{"type": "rejoin", "room": exampleRoom, "nickname": "Giang", "pin": wrongPIN(giang.pin)})
		if i < 5 {
			g.expectError(c, i, "bad_credentials", "")
		}
	}
	g.expectError(c, 5, "pin_locked", "error.pin_locked")
	c.Close()
	g.sync()
}

func waitWelcome(t *testing.T, c *wsclient.Client) {
	t.Helper()
	if !c.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Welcome != nil && s.View != nil }) {
		t.Fatalf("no welcome: %v", c.Snapshot().Errors)
	}
}

func wrongPIN(pin string) string {
	if pin == "0000" {
		return "1111"
	}
	return "0000"
}

// insertPhone puts p back in join order (ids are p1, p2...).
func insertPhone(list []*exPhone, p *exPhone) []*exPhone {
	for i, q := range list {
		if len(q.id) > len(p.id) || (len(q.id) == len(p.id) && q.id > p.id) {
			return append(list[:i], append([]*exPhone{p}, list[i:]...)...)
		}
	}
	return append(list, p)
}

func mustUnmarshal(t *testing.T, b []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("%s: %v", b, err)
	}
}
