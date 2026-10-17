package room

import (
	"reflect"
	"time"

	"masoi/internal/engine"
)

// Message is what one phone is sent. It carries only that phone's own
// projection of the game (engine.ViewFor and engine.Visible), never State.
type Message struct {
	View   *ViewMsg
	Events *EventsMsg
}

// ViewMsg is the screen now, plus what only the room knows.
type ViewMsg struct {
	Deadline time.Time // when the running phase or night step ends, zero if untimed
	Speaker  bool      // this phone reads announcements aloud
	View     engine.View
}

// EventsMsg is the events this phone may know about. Reset means Events is the
// whole list; otherwise they continue at index From.
type EventsMsg struct {
	Reset  bool
	From   int
	Events []engine.Event
}

// Sink receives messages on the room goroutine; it must not block.
type Sink interface {
	Deliver(Message)
	// Evict tells this phone that a rejoin elsewhere has taken over its
	// player: the sink must close the connection, after telling it why if
	// it can (the message is reason.Error()).
	Evict(reason error)
}

type connection struct {
	player engine.PlayerID
	sink   Sink
	sent   int      // visible events already sent
	full   bool     // the last view was a full one
	fresh  bool     // nothing sent yet
	last   *ViewMsg // the last view sent
}

func (r *Room) connected(p engine.PlayerID) bool { return r.conns[p] != nil }

func (r *Room) table() Table {
	order := r.joined
	if len(r.state.Seats) > 0 && len(r.state.Seats) == len(r.joined) {
		order = r.state.Seats
	}
	return Table{Host: r.host, Order: order, Connected: r.connected, Current: r.speaker}
}

// changed sends every connected phone whose screen changed its own view and
// any new events. A phone whose screen is the same as before gets nothing: a
// message arriving on every phone when one player acts in secret would tell
// the table that the role being called is alive (rules §3).
func (r *Room) changed() {
	speakers := r.pol.Speakers(r.table())
	speaking := map[engine.PlayerID]bool{}
	for _, p := range speakers {
		speaking[p] = true
	}
	// The room remembers who is speaking (rules §1) only when there is a
	// single one to remember; in AllPhones mode there is nothing to stick to.
	r.speaker = ""
	if len(speakers) == 1 {
		r.speaker = speakers[0]
	}
	for _, c := range r.conns {
		r.send(c, speaking[c.player])
	}
}

func (r *Room) send(c *connection, speaker bool) {
	viewer := engine.PlayerViewer(c.player)
	view := engine.ViewFor(r.state, r.log, viewer)
	visible := engine.Visible(r.state, r.log, viewer)

	vm := &ViewMsg{Deadline: r.deadline, Speaker: speaker, View: view}
	msg := Message{View: vm}
	switch {
	case c.fresh || view.Full != c.full || len(visible) < c.sent:
		msg.Events = &EventsMsg{Reset: true, Events: visible}
	case len(visible) > c.sent:
		msg.Events = &EventsMsg{From: c.sent, Events: visible[c.sent:]}
	}
	if msg.Events == nil && c.last != nil && reflect.DeepEqual(*c.last, *vm) {
		return // nothing new for this screen
	}
	c.sent, c.full, c.fresh, c.last = len(visible), view.Full, false, vm
	c.sink.Deliver(msg)
}
