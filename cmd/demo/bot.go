package main

import (
	"context"
	"fmt"
	"math/rand/v2"

	"masoi/internal/transport/ws/wsclient"
)

// bot is a scripted player. It decides only from its own view, like a phone.
type bot struct {
	nick   string
	c      *wsclient.Client
	id     string
	role   string
	acted  map[string]bool // phase/seq already answered
	guard  string          // whom the Guard protected last night
	visit  string          // whom the Visitor visited last night
	cursed bool            // the Cursed wolf already used his curse
}

func newBot(nick string, c *wsclient.Client) *bot {
	return &bot{nick: nick, c: c, acted: map[string]bool{}}
}

// run plays until the game ends or ctx is done.
func (b *bot) run(ctx context.Context) {
	seen := 0
	for {
		s := b.c.Snapshot()
		if len(s.Frames) > seen {
			seen = len(s.Frames)
			if s.View != nil {
				if s.View.Str("view", "phase") == "Ended" {
					return
				}
				b.think(ctx, s.View)
			}
		}
		if !b.c.Wait(ctx, func(s wsclient.Snapshot) bool { return len(s.Frames) > seen }) {
			return
		}
	}
}

func (b *bot) think(ctx context.Context, f wsclient.Frame) {
	phase := f.Str("view", "phase")
	key := fmt.Sprintf("%s/%v", phase, f.Get("view", "seq"))
	if b.acted[key] {
		return
	}
	name, args := b.decide(f, phase)
	if name == "" {
		return
	}
	b.acted[key] = true
	_ = b.c.Send(ctx, map[string]any{"type": "command", "name": name, "args": args})
}

func (b *bot) decide(f wsclient.Frame, phase string) (string, map[string]any) {
	me, _ := f.Get("view", "me").(map[string]any)
	alive, _ := me["alive"].(bool)
	living := b.living(f)
	switch phase {
	case "ChiefElection":
		return "voteChief", map[string]any{"candidate": pick(living)}
	case "ChiefSuccession":
		if alive || len(living) == 0 || f.Str("view", "succession") != b.id {
			return "", nil
		}
		return "nameSuccessor", map[string]any{"successor": pick(living)}
	case "Night":
		return b.night(f, me, living)
	case "Day":
		if !alive {
			return "", nil
		}
		return "nominate", map[string]any{"target": pick(without(living, b.id))}
	case "Verdict":
		if !alive || f.Str("view", "votes", "defendant") == b.id {
			return "", nil
		}
		return "verdict", map[string]any{"kill": rand.IntN(10) < 7}
	}
	return "", nil
}

func (b *bot) night(f wsclient.Frame, me map[string]any, living []string) (string, map[string]any) {
	turn, _ := me["turn"].(map[string]any)
	if turn == nil || turn["blocked"] == true {
		return "", nil
	}
	others := without(living, b.id)
	if turn["step"] == "Werewolf" { // the pack step: any wolf, a cursed player too
		return "wolfVote", map[string]any{"target": pick(without(others, packMembers(f)...))}
	}
	switch b.role {
	case "Guard":
		target := pick(without(living, b.guard))
		b.guard = target
		return "guardProtect", map[string]any{"target": target}
	case "Visitor":
		target := pick(without(others, b.visit))
		b.visit = target
		return "visitorVisit", map[string]any{"target": target}
	case "CursedWolf":
		// His own step comes before the pack's: curse once, then bite as a wolf.
		if b.cursed {
			return "", nil
		}
		b.cursed = true
		return "cursedWolfCurse", map[string]any{"target": pick(without(others, packMembers(f)...))}
	case "WolfSeer":
		// His own step comes before the pack's; the pack step is handled above.
		return "wolfSeerCheck", map[string]any{"target": pick(living)}
	case "Seer":
		return "seerCheck", map[string]any{"target": pick(others)}
	}
	return "", nil
}

// living lists the players the view shows as alive.
func (b *bot) living(f wsclient.Frame) []string {
	var out []string
	players, _ := f.Get("view", "players").([]any)
	for _, p := range players {
		m, _ := p.(map[string]any)
		if m["alive"] == true {
			out = append(out, m["id"].(string))
		}
	}
	return out
}

func packMembers(f wsclient.Frame) []string {
	var out []string
	members, _ := f.Get("view", "pack", "members").([]any)
	for _, m := range members {
		out = append(out, m.(string))
	}
	return out
}

func without(list []string, drop ...string) []string {
	var out []string
	for _, s := range list {
		keep := true
		for _, d := range drop {
			if s == d {
				keep = false
			}
		}
		if keep {
			out = append(out, s)
		}
	}
	return out
}

func pick(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[rand.IntN(len(list))]
}
