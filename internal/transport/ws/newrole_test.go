package ws

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/transport/ws/wsclient"
)

// The Mime is a throwaway role that exists only in this file. Everything a new
// role brings to the wire is declared here, beside it: its command, its rank
// in the role list and its own error. Nothing else in the repository names it.
type mime struct{}

const mimeID engine.RoleID = "Mime"

// ErrMimeSilent is the Mime's own error, with its own wire code.
var ErrMimeSilent = engine.NewError("mime_silent", "the mime cannot speak")

// MimeGesture is the Mime's command.
type MimeGesture struct {
	engine.CommandMarker
	By     engine.PlayerID
	Target engine.PlayerID
}

type mimeArgs struct {
	Target engine.PlayerID `json:"target"`
}

func (mime) ID() engine.RoleID            { return mimeID }
func (mime) BaseTeam() engine.Team        { return engine.TeamVillage }
func (mime) CardOrder() int               { return 55 } // between the Seer and the Guard
func (mime) NightCall() *engine.NightCall { return nil }
func (mime) Evolve(*engine.State, engine.Event) bool {
	return false
}

func (mime) Decide(_ engine.State, c engine.Command) ([]engine.Event, bool, error) {
	if _, ok := c.(MimeGesture); !ok {
		return nil, false, nil
	}
	return nil, true, ErrMimeSilent
}

func (mime) Commands() []engine.CommandSpec {
	return []engine.CommandSpec{
		engine.WireCommand("mimeGesture", func(a mimeArgs, by engine.PlayerID) engine.Command {
			return MimeGesture{By: by, Target: a.Target}
		}),
	}
}

// registerMime adds the Mime for one test only: the other tests (the golden
// command list first) see the real roles alone.
func registerMime(t *testing.T) {
	t.Helper()
	engine.Register(mime{})
	t.Cleanup(func() { delete(engine.Registry, mimeID) })
}

// Audit C1: a role registered from its own file gets a wire command, a slot
// in the public role list and an error code, with no other file edited.
func TestNewRole_GetsACommandACardSlotAndAnErrorCode_FromItsOwnFile(t *testing.T) {
	registerMime(t)
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
	tb.do(host, "configureRoles", map[string]any{"roles": map[string]int{
		"Guard": 1, "Mime": 1, "Seer": 1, "Villager": 1, "Werewolf": 1,
	}})

	// Card slot: the public role list puts the Mime by its own rank.
	want := []string{"Werewolf", "Villager", "Seer", "Mime", "Guard"}
	tb.waitAll("the role list with the Mime", func(_ *phone, s wsclient.Snapshot) bool {
		return s.View != nil && reflect.DeepEqual(roleListIDs(s), want)
	})

	// Command and error: the Mime's command reaches the engine, and its own
	// error reaches the phone with its own code.
	tb.do(tb.p[2], "mimeGesture", map[string]any{"target": tb.p[1].id})
	if got := e.waitError(tb.p[2].c, 1); got != "mime_silent" {
		t.Fatalf("mimeGesture: code %q, want mime_silent", got)
	}
	if msg := tb.p[2].c.Snapshot().Errors[0].Str("message"); msg != "the mime cannot speak" {
		t.Fatalf("mimeGesture: message %q", msg)
	}
	// The args are checked like any other command's.
	tb.do(tb.p[2], "mimeGesture", map[string]any{"target": tb.p[1].id, "by": tb.p[3].id})
	if got := e.waitError(tb.p[2].c, 2); got != "bad_message" {
		t.Fatalf("mimeGesture with a sender in the args: code %q, want bad_message", got)
	}
}

func roleListIDs(s wsclient.Snapshot) []string {
	var ids []string
	for _, c := range asList(s.View.Get("view", "roleCards")) {
		id, _ := c["role"].(string)
		ids = append(ids, id)
	}
	return ids
}
