package ws

import (
	"strings"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/room"
)

// chatExamples has p write until the limit refuses them, with the refusals of
// a message on the way.
func (g *exGame) chatExamples(p *exPhone) {
	g.t.Helper()
	g.ok(p, "sendChat", map[string]any{"text": "Tôi nghi Phúc."})
	g.fails(p, "chat_empty", "sendChat", map[string]any{"text": "  "})
	g.fails(p, "chat_too_long", "sendChat", map[string]any{"text": strings.Repeat("a", engine.MaxChatLen+1)})
	for range room.ChatBurst - 1 {
		g.ok(p, "sendChat", map[string]any{"text": "Phúc là sói!"})
	}
	g.fails(p, "chat_too_fast", "sendChat", map[string]any{"text": "Phúc là sói!"})
}

// chatFullExample gets server/error.chat_full.json from a table of its own,
// where five players write until the game holds its last message. Only that
// refusal joins x: the messages before it would add thousands of frames to
// validate and say nothing new.
func chatFullExample(t *testing.T, x *exampleSet) {
	t.Helper()
	g := &exGame{t: t, e: newExampleEnv(t), x: newExampleSet(t)}
	_, body := g.httpJSON("POST", "/rooms", []byte(`{"mode":"inPerson"}`))
	var created struct{ HostKey string }
	mustUnmarshal(t, body, &created)
	order := make([]any, 0, engine.MinPlayers)
	for i, nick := range exampleTable[:engine.MinPlayers] {
		msg := map[string]any{"type": "join", "room": exampleRoom, "nickname": nick}
		if i == 0 {
			msg["hostKey"] = created.HostKey
		}
		order = append(order, g.join(nick, "", msg).id)
	}
	host := g.host()
	g.ok(host, "setSeats", map[string]any{"order": order})
	g.ok(host, "configureRoles", map[string]any{"roles": map[string]any{"Werewolf": 1, "Villager": engine.MinPlayers - 1}})
	g.ok(host, "configureTimers", map[string]any{"timers": map[string]any{"day": 1800000}})
	g.ok(host, "dealRoles", nil)
	g.untilPhase("Day")
	for sent := 0; sent < room.ChatMaxPerGame; g.advance(room.ChatBurst * room.ChatInterval) {
		for _, p := range g.phones {
			for i := 0; i < room.ChatBurst && sent < room.ChatMaxPerGame; i++ {
				g.do(p, "sendChat", map[string]any{"text": "Bàn tán."})
				sent++
			}
		}
		g.sync() // the room must take them before the clock moves, or the credit is off
	}
	g.expectPhase("Day")
	g.fails(host, "chat_full", "sendChat", map[string]any{"text": "Bàn tán."})
	x.files["server/error.chat_full.json"] = g.x.files["server/error.chat_full.json"]
}

// secondGame: a new game in the same room, every phone reading aloud. The
// Hunter, elected Chief, is hanged on day 1 and lets both his shot and his
// succession run out; then the table votes to call the game off.
func (g *exGame) secondGame() {
	an := g.phone("An")
	g.ok(an, "newGame", nil)
	g.expectPhase("Lobby")
	g.ok(an, "configureSound", map[string]any{"mode": "allPhones"})
	g.ok(an, "dealRoles", nil)
	for _, p := range g.phones {
		p.role = p.view().Str("view", "me", "role")
	}
	hunter := g.byRole("Hunter")
	g.toDeadline()
	g.expectPhase("ChiefElection")
	for _, p := range g.phones {
		g.ok(p, "voteChief", map[string]any{"candidate": hunter.id})
	}
	g.untilPhase("Day")
	g.nominateAll(func(p *exPhone) *exPhone {
		if p == hunter {
			return g.phones[0]
		}
		return hunter
	})
	g.untilPhase("Verdict")
	for _, p := range g.livingPhones() {
		if p != hunter {
			g.ok(p, "verdict", map[string]any{"kill": true})
		}
	}
	g.expectPhase("HunterShot")
	g.toDeadline()
	g.expectPhase("ChiefSuccession")
	g.toDeadline()
	for _, p := range g.livingPhones() {
		if g.phase() == "Ended" {
			break
		}
		g.ok(p, "abortVote", map[string]any{"on": true})
	}
	g.expectPhase("Ended")
	g.shot(an, "ended.aborted")
}

func (g *exGame) byRole(role string) *exPhone {
	g.t.Helper()
	for _, p := range g.phones {
		if p.role == role {
			return p
		}
	}
	g.t.Fatalf("nobody is %s", role)
	return nil
}
