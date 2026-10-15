package ws

import (
	"testing"
	"time"

	"masoi/internal/i18n"
	"masoi/internal/roles"
	"masoi/internal/transport/ws/wsclient"
)

func TestE2E_FivePhonesPlayANightAndADayToAHanging(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.Burst = 100 })
	code := e.createRoom()
	tb := &table{t: t, e: e}
	for _, nick := range []string{"An", "Binh", "Chi", "Dung", "Em"} {
		c := e.join(code, nick)
		tb.p = append(tb.p, &phone{c: c, id: c.Snapshot().Welcome.Str("player"), nick: nick})
	}

	// The host sets the table, the role list and short times, then deals.
	order := make([]string, len(tb.p))
	for i, p := range tb.p {
		order[i] = p.id
	}
	host := tb.host()
	tb.do(host, "setSeats", map[string]any{"order": order})
	tb.do(host, "configureRoles", map[string]any{"roles": map[string]int{string(roles.Werewolf): 1, "Guard": 1, "Seer": 1, "Villager": 2}})
	short := 5 * time.Second
	tb.do(host, "configureTimers", map[string]any{"timers": map[string]any{
		"roleView": 5000, "chiefElection": 5000, "nightStep": 5000, "announce": 5000, "day": 60000,
		"nominationLock": 5000, "defense": 5000, "verdict": 5000,
	}})
	tb.do(host, "dealRoles", nil)
	tb.waitPhase("RoleView")
	for _, p := range tb.p {
		p.role = p.c.Snapshot().View.Str("view", "me", "role")
		if p.role == "" {
			t.Fatalf("%s has no role after the deal", p.nick)
		}
	}
	wolf, guard := tb.byRole("Werewolf"), tb.byRole("Guard")
	var villagers []*phone
	for _, p := range tb.p {
		if p.role == "Villager" {
			villagers = append(villagers, p)
		}
	}
	victim, other := villagers[0], villagers[1]

	// Role view, then everybody votes for the host as Chief.
	tb.advance(short)
	tb.waitPhase("ChiefElection")
	for _, p := range tb.p {
		tb.do(p, "voteChief", map[string]any{"candidate": host.id})
	}
	tb.waitAll("every Chief vote", func(_ *phone, s wsclient.Snapshot) bool {
		m, _ := s.View.Get("view", "votes", "chief").(map[string]any)
		return len(m) == 5
	})
	tb.advance(short)

	// Night 1: the wolf bites a villager, the Guard protects the other one.
	for step := 0; tb.host().c.Snapshot().View.Str("view", "phase") == "Night"; step++ {
		if step > 10 {
			t.Fatal("the night never ends")
		}
		now := seqOf(tb.host().c.Snapshot())
		tb.waitAll("the same night step", func(_ *phone, s wsclient.Snapshot) bool { return seqOf(s) == now })
		for _, p := range tb.p {
			if p.c.Snapshot().View.Get("view", "me", "turn") == nil {
				continue
			}
			switch p.role {
			case "Werewolf":
				tb.do(p, "wolfVote", map[string]any{"target": victim.id})
			case "Guard":
				tb.do(p, "guardProtect", map[string]any{"target": other.id})
			case "Seer":
				tb.do(p, "seerCheck", map[string]any{"target": wolf.id})
			}
		}
		// The step must not end before the room has the command.
		for _, p := range tb.p {
			if p.c.Snapshot().View.Get("view", "me", "turn") == nil {
				continue
			}
			p.c.Wait(ctx(t), func(s wsclient.Snapshot) bool {
				switch p.role {
				case "Werewolf":
					return s.View.Get("view", "pack", "votes", p.id) == victim.id
				case "Guard":
					return hasEvent(s, "PlayerProtected")
				default:
					return hasEvent(s, "SeerChecked")
				}
			})
		}
		tb.advance(short)
	}

	// Dawn: the announcement names the dead player, in Vietnamese.
	for tb.host().c.Snapshot().View.Str("view", "phase") != "Day" {
		switch phase := tb.host().c.Snapshot().View.Str("view", "phase"); phase {
		case "Dawn":
			tb.waitAll("the dawn announcement", func(_ *phone, s wsclient.Snapshot) bool { return hasEvent(s, "DawnAnnounced") })
			tb.advance(short)
		case "ChiefSuccession":
			tb.advance(15 * time.Second)
		default:
			t.Fatalf("unexpected phase %q after the night", phase)
		}
		tb.waitAll("a settled phase", func(*phone, wsclient.Snapshot) bool { return true })
	}
	wantDawn := i18n.Vi(i18n.DawnSomeoneDied, victim.nick)
	tb.waitAll("the dawn line", func(_ *phone, s wsclient.Snapshot) bool { return sayOf(s, "DawnAnnounced") == wantDawn })
	if tb.alive(victim) {
		t.Fatalf("%s should be dead", victim.nick)
	}
	if got := victim.c.Snapshot().View.Get("view", "full"); got != true {
		t.Fatalf("a dead player must see everything, full = %v", got)
	}
	if got := guard.c.Snapshot().View.Get("view", "full"); got == true {
		t.Fatal("a living player got the full view")
	}

	// Day: everybody nominates the wolf, who nominates the Guard.
	for _, p := range tb.p {
		if !tb.alive(p) {
			continue
		}
		target := wolf.id
		if p == wolf {
			target = guard.id
		}
		tb.do(p, "nominate", map[string]any{"target": target})
	}
	tb.waitPhase("NominationLock")
	tb.advance(short)
	tb.waitPhase("Trial")
	tb.advance(short)
	tb.waitPhase("Verdict")
	for _, p := range tb.p {
		if tb.alive(p) && p != wolf {
			tb.do(p, "verdict", map[string]any{"kill": true})
		}
	}
	tb.waitPhase("Ended")

	// The wolf hangs, the village wins, and everybody now sees everything.
	tb.waitAll("the end of the game", func(_ *phone, s wsclient.Snapshot) bool {
		return sayOf(s, "Hanged") == i18n.Vi(i18n.DayHanged, wolf.nick) &&
			sayOf(s, "GameEnded") == i18n.Vi(i18n.WinVillage) && s.View.Get("view", "full") == true
	})
	for _, p := range tb.p {
		if got := p.c.Snapshot().View.Str("view", "winner"); got != "Village" {
			t.Errorf("%s: winner = %q", p.nick, got)
		}
		if len(p.c.Snapshot().Errors) != 0 {
			t.Errorf("%s got errors: %v", p.nick, p.c.Snapshot().Errors)
		}
	}
	for _, p := range tb.p {
		checkOnlyItsOwnView(t, p)
		resets := checkEventStream(t, p)
		if p == victim && resets == 0 {
			t.Errorf("%s died and now sees everything, but never got the full history", p.nick)
		}
	}
	if host.c.Snapshot().View["speaker"] != true {
		t.Error("the host phone is not the speaker")
	}
	for _, p := range tb.p[1:] {
		for _, f := range p.c.Snapshot().Frames {
			if _, has := f["speaker"]; has {
				t.Errorf("%s was told to speak", p.nick)
			}
		}
	}
}
