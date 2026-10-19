package ws

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"masoi/internal/room"
	"masoi/internal/transport/ws/wsclient"
)

// exPhone is one phone of an example game.
type exPhone struct {
	nick, id, role string
	c              *wsclient.Client
	token, pin     string
}

// exGame drives one room.
type exGame struct {
	t      *testing.T
	e      *env
	x      *exampleSet
	phones []*exPhone // joined and connected, in join order
}

func newExampleEnv(t *testing.T) *env {
	t.Helper()
	e := &env{
		t:    t,
		clk:  room.NewFakeClock(time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)),
		now:  &fakeNow{t: time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)},
		logs: &syncBuffer{},
	}
	e.rooms = room.NewManager(room.ManagerConfig{
		Clock: e.clk, Random: &exampleRandom{},
		NewCode: func() string { return exampleRoom },
	})
	handler, err := New(Config{Rooms: e.rooms, AllowAnyOrigin: true, Now: e.now.Now, RatePerSecond: 1e6, Burst: 1e6})
	if err != nil {
		t.Fatal(err)
	}
	e.srv = httptest.NewServer(handler)
	t.Cleanup(func() {
		e.srv.Close()
		e.rooms.Close()
	})
	return e
}

// httpJSON makes a request and returns the status and the body.
func (g *exGame) httpJSON(method, path string, body []byte) (int, []byte) {
	g.t.Helper()
	req, err := http.NewRequest(method, g.e.srv.URL+path, bytes.NewReader(body))
	if err != nil {
		g.t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		g.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		g.t.Fatal(err)
	}
	return resp.StatusCode, b
}

// recordHTTP stores a body as an example and for validation.
func (g *exGame) recordHTTP(name, def string, body []byte) {
	g.t.Helper()
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		g.t.Fatalf("%s: %v", name, err)
	}
	g.x.put("http/"+name+".json", v)
	g.x.http = append(g.x.http, sentMsg{"http/" + name, "schema/http.schema.json#/$defs/" + def, body})
}

// dial opens a phone that has not joined yet.
func (g *exGame) dial() *wsclient.Client {
	g.t.Helper()
	c := g.e.dial(nil)
	g.x.clients = append(g.x.clients, c)
	return c
}

// send sends msg; a non-empty example name stores it as an example.
func (g *exGame) send(c *wsclient.Client, example string, msg map[string]any) {
	g.t.Helper()
	b, _ := json.Marshal(msg)
	g.x.sent = append(g.x.sent, sentMsg{"sent " + string(b), "schema/client.schema.json", b})
	if example != "" {
		g.x.put("client/"+example+".json", msg)
	}
	g.e.send(c, msg)
}

// do sends a command; the first time a command name is used it becomes its
// example.
func (g *exGame) do(p *exPhone, name string, args map[string]any) {
	g.t.Helper()
	msg := map[string]any{"type": "command", "name": name}
	if args != nil {
		msg["args"] = args
	}
	example := "command." + name
	if _, ok := g.x.files["client/"+example+".json"]; ok {
		example = ""
	}
	g.send(p.c, example, msg)
}

// sync waits until every phone has received everything the room sent so
// far: each phone sends a request the room always refuses (kicking itself),
// which the room answers after every earlier delivery to that phone.
func (g *exGame) sync() {
	g.t.Helper()
	for _, p := range g.phones {
		n := len(p.c.Snapshot().Errors)
		g.e.send(p.c, map[string]any{"type": "kick", "target": p.id})
		if !p.c.Wait(ctx(g.t), func(s wsclient.Snapshot) bool { return len(s.Errors) > n }) {
			g.t.Fatalf("%s: no answer to the sync probe", p.nick)
		}
	}
}

// expectError waits for c's next error after n and stores it as an example.
func (g *exGame) expectError(c *wsclient.Client, n int, code, example string) {
	g.t.Helper()
	if !c.Wait(ctx(g.t), func(s wsclient.Snapshot) bool { return len(s.Errors) > n }) {
		g.t.Fatalf("no error %s", code)
	}
	got := c.Snapshot().Errors[n]
	if got.Str("code") != code {
		g.t.Fatalf("error %v, want %s", got, code)
	}
	if example != "" {
		g.x.put("server/"+example+".json", got)
	}
}

// fails sends a command that must be refused with code, stored as the
// example server/error.<code>.json unless one exists.
func (g *exGame) fails(p *exPhone, code, name string, args map[string]any) {
	g.t.Helper()
	n := len(p.c.Snapshot().Errors)
	g.do(p, name, args)
	example := "error." + code
	if _, ok := g.x.files["server/"+example+".json"]; ok {
		example = ""
	}
	g.expectError(p.c, n, code, example)
}

func (g *exGame) phone(nick string) *exPhone {
	g.t.Helper()
	for _, p := range g.phones {
		if p.nick == nick {
			return p
		}
	}
	g.t.Fatalf("no phone %s", nick)
	return nil
}

func (g *exGame) drop(p *exPhone) {
	for i, q := range g.phones {
		if q == p {
			g.phones = append(g.phones[:i], g.phones[i+1:]...)
			return
		}
	}
}

// view is the phone's latest view message.
func (p *exPhone) view() wsclient.Frame { return p.c.Snapshot().View }

// lastFrame is the phone's latest message of a type.
func (p *exPhone) lastFrame(typ string) wsclient.Frame {
	frames := p.c.Snapshot().Frames
	for i := len(frames) - 1; i >= 0; i-- {
		if frames[i].Type() == typ {
			return frames[i]
		}
	}
	return nil
}

// shot stores the phone's latest view as server/view.<name>.json.
func (g *exGame) shot(p *exPhone, name string) {
	g.t.Helper()
	g.x.put("server/view."+name+".json", p.view())
}

// advance moves the room clock by d, then waits for the phones.
func (g *exGame) advance(d time.Duration) {
	g.t.Helper()
	g.e.clk.Advance(d)
	g.sync()
}

// ok sends a command that must be accepted, and waits for the phones.
func (g *exGame) ok(p *exPhone, name string, args map[string]any) {
	g.t.Helper()
	n := len(p.c.Snapshot().Errors)
	g.do(p, name, args)
	g.sync()
	if errs := p.c.Snapshot().Errors; len(errs) != n+1 {
		g.t.Fatalf("%s %s %v refused: %v", p.nick, name, args, errs[n])
	}
}
