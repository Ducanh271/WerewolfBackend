package room

import (
	"fmt"
	"testing"

	"masoi/internal/engine"
)

// BenchmarkBroadcast measures one full broadcast (audit R5): every connected
// phone gets its own view and events, as after any valid command. Run with
//
//	go test ./internal/room -run '^$' -bench Broadcast
func BenchmarkBroadcast(b *testing.B) {
	for _, table := range benchTables {
		g := playBenchGame(b, table.players, table.cast)
		for _, cut := range g.cuts() {
			b.Run(fmt.Sprintf("%dp/%s", table.players, cut.name), func(b *testing.B) {
				r := benchRoom(g, cut.events)
				r.changed() // the first send to a phone is a full reset; measure the usual ones
				for b.Loop() {
					r.changed()
				}
				ms := float64(b.Elapsed().Nanoseconds()) / 1e6 / float64(b.N)
				b.ReportMetric(ms, "ms/broadcast")
				b.ReportMetric(ms/float64(len(r.conns)), "ms/view")
				b.ReportMetric(float64(cut.events), "events")
			})
		}
	}
}

// benchRoom is a room at the point of the game where the log has events
// events, with every seat connected. It runs no goroutine: the benchmark is
// the only one touching it.
func benchRoom(g *benchGame, events int) *Room {
	log := g.log[:events:events]
	s := engine.Fold(engine.NewState(), log...)
	r := &Room{pol: InPersonSpeaker{}, state: s, log: log, host: s.Seats[0], conns: map[engine.PlayerID]*connection{}}
	for _, id := range s.Seats {
		r.joined = append(r.joined, id)
		r.conns[id] = &connection{player: id, sink: discardSink{}, fresh: true}
	}
	return r
}

type discardSink struct{}

func (discardSink) Deliver(Message) {}
func (discardSink) Evict(error)     {}
