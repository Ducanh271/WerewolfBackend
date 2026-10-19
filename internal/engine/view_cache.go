package engine

import "slices"

// ViewCache answers ViewFor and Visible for one room's log without replaying
// the log once per viewer (audit R5). The state the log replays to is the
// same for every viewer; only the audience filter differs. So the cache
// replays each event once, when it first sees it, and keeps its audience as
// judged at that moment (the rule Visible follows); a viewer then only
// filters. The end-of-game summary is narrated once per log length.
//
// What is judged now, not at the time of the event, stays with the caller's
// state: a viewer who has died or a game that has ended sees the whole log
// (seesAll), exactly as Visible does.
//
// log must only grow between calls (the room's log does, across NewGame
// too); a shorter log than last time rebuilds the cache. The zero value is
// ready to use. A ViewCache is not safe for concurrent use: a room keeps it
// on its own goroutine.
type ViewCache struct {
	at   State         // the replay of the events in seen
	seen []audienceSet // the audience of each event of log, judged at the time

	summaryLen int // len(log) that summary narrates; valid when summaryOK
	summaryOK  bool
	summary    []string
}

// audienceSet is an Audience resolved against the state of its time: the
// living wolves are named.
type audienceSet struct {
	everyone bool
	players  []PlayerID
}

// Visible returns the same events as Visible(s, log, v).
func (c *ViewCache) Visible(s State, log []Event, v Viewer) []Event {
	c.sync(log)
	if seesAll(s, v) {
		return append([]Event(nil), log...)
	}
	var out []Event
	for i, e := range log {
		if c.seen[i].includes(v) {
			out = append(out, e)
		}
	}
	return out
}

// ViewFor returns the same view as ViewFor(s, log, v).
func (c *ViewCache) ViewFor(s State, log []Event, v Viewer) View {
	return project(s, v, func() []string {
		if !c.summaryOK || c.summaryLen != len(log) {
			c.summary, c.summaryLen, c.summaryOK = summaryFor(log), len(log), true
		}
		return slices.Clone(c.summary)
	})
}

// sync replays the events of log the cache has not seen yet.
func (c *ViewCache) sync(log []Event) {
	if len(log) < len(c.seen) || len(c.seen) == 0 {
		c.at, c.seen = NewState(), c.seen[:0]
	}
	for _, e := range log[len(c.seen):] {
		c.seen = append(c.seen, resolve(c.at, audienceOf(c.at, e)))
		c.at = Evolve(c.at, e)
	}
}

// resolve names everybody a includes in s (see Audience.includes).
func resolve(s State, a Audience) audienceSet {
	if a.Everyone {
		return audienceSet{everyone: true}
	}
	out := audienceSet{players: slices.Clone(a.Players)}
	if a.Wolves {
		for id, p := range s.Players {
			if p.Alive && VisibleTeam(p) == TeamWolves {
				out.players = append(out.players, id)
			}
		}
	}
	return out
}

func (a audienceSet) includes(v Viewer) bool {
	return a.everyone || slices.Contains(a.players, v.Player)
}
