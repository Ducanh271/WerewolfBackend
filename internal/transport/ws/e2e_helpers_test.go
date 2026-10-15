package ws

import (
	"fmt"
	"testing"
	"time"

	"masoi/internal/transport/ws/wsclient"
)

// phone is one fake player.
type phone struct {
	c    *wsclient.Client
	id   string
	nick string
	role string
}

type table struct {
	t *testing.T
	e *env
	p []*phone
}

func (tb *table) host() *phone { return tb.p[0] }

func (tb *table) waitAll(what string, cond func(p *phone, s wsclient.Snapshot) bool) {
	tb.t.Helper()
	for _, p := range tb.p {
		if !p.c.Wait(ctx(tb.t), func(s wsclient.Snapshot) bool { return cond(p, s) }) {
			tb.t.Fatalf("%s never saw: %s\nphase %q, errors %v", p.nick, what, p.c.Snapshot().View.Str("view", "phase"), p.c.Snapshot().Errors)
		}
	}
}

func (tb *table) waitPhase(phase string) {
	tb.t.Helper()
	tb.waitAll("phase "+phase, func(_ *phone, s wsclient.Snapshot) bool {
		return s.View != nil && s.View.Str("view", "phase") == phase
	})
}

func seqOf(s wsclient.Snapshot) float64 { n, _ := s.View.Get("view", "seq").(float64); return n }

// advance moves the room's clock and waits until every phone has a newer seq
// (the phase or the night step changed).
func (tb *table) advance(d time.Duration) {
	tb.t.Helper()
	before := seqOf(tb.host().c.Snapshot())
	tb.e.clk.Advance(d)
	tb.waitAll("a new phase or step", func(_ *phone, s wsclient.Snapshot) bool { return seqOf(s) > before })
}

func (tb *table) do(p *phone, name string, args map[string]any) {
	tb.t.Helper()
	tb.e.send(p.c, cmd(name, args))
}

func (tb *table) alive(p *phone) bool { return p.c.Snapshot().View.Get("view", "me", "alive") == true }

func (tb *table) byRole(role string) *phone {
	for _, p := range tb.p {
		if p.role == role {
			return p
		}
	}
	tb.t.Fatalf("nobody is %s", role)
	return nil
}

func hasEvent(s wsclient.Snapshot, typ string) bool {
	for _, ev := range s.Events {
		if ev.Str("type") == typ {
			return true
		}
	}
	return false
}

func sayOf(s wsclient.Snapshot, typ string) string {
	for _, ev := range s.Events {
		if ev.Str("type") == typ {
			return ev.Str("say")
		}
	}
	return ""
}

// checkOnlyItsOwnView goes through everything the phone was ever sent. While
// the player is alive and the game is on, nothing may name another player's
// role or team, show the pack to a non-wolf, or carry a private result that is
// not theirs. Dead players and the finished game see everything (rules §4).
func checkOnlyItsOwnView(t *testing.T, p *phone) {
	t.Helper()
	full := false
	for i, f := range p.c.Snapshot().Frames {
		where := fmt.Sprintf("%s frame %d (%s)", p.nick, i, f.Type())
		switch f.Type() {
		case "view":
			full = f.Get("view", "full") == true
			if got := f.Str("view", "me", "id"); got != p.id {
				t.Errorf("%s: view of %q", where, got)
			}
			if full {
				continue
			}
			for _, pl := range asList(f.Get("view", "players")) {
				if (pl["role"] != nil || pl["team"] != nil) && pl["id"] != p.id {
					t.Errorf("%s: shows the role or team of %v", where, pl["id"])
				}
			}
			if f.Get("view", "pack") != nil && p.role != "Werewolf" {
				t.Errorf("%s: shows the wolf pack to a %s", where, p.role)
			}
			if f.Get("view", "secrets") != nil || f.Get("view", "lover") != nil {
				t.Errorf("%s: carries secrets", where)
			}
		case "events":
			if full {
				continue
			}
			for _, ev := range asList(f["events"]) {
				typ, _ := ev["type"].(string)
				switch typ {
				case "PlayerDied", "RolesDealt":
					t.Errorf("%s: a living player was sent %s", where, typ)
				case "SeerChecked":
					if p.role != "Seer" {
						t.Errorf("%s: %s was sent the Seer's result", where, p.role)
					}
				case "PlayerProtected":
					if p.role != "Guard" {
						t.Errorf("%s: %s was sent the Guard's choice", where, p.role)
					}
				case "WolfVoted":
					if p.role != "Werewolf" {
						t.Errorf("%s: %s was sent a wolf vote", where, p.role)
					}
				}
			}
		}
	}
}

func asList(v any) []map[string]any {
	var out []map[string]any
	l, _ := v.([]any)
	for _, x := range l {
		if m, ok := x.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
