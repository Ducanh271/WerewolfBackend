package room

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// The room's side of the differential test of the fast view path (audit
// R5): with newCheckedGame every message any phone gets is compared, as it
// is delivered, with the reference path, engine.ViewFor and engine.Visible
// replaying the whole log.

// refSink is a phone that checks each message against the full replay. Deliver
// runs on the room goroutine (Sink), so it may read the room's state and log.
type refSink struct {
	recSink
	t      *testing.T
	r      *Room
	player engine.PlayerID // "" until the first view names it
}

func (s *refSink) Deliver(m Message) {
	s.recSink.Deliver(m)
	if s.player == "" && m.View.View.Me != nil {
		s.player = m.View.View.Me.ID
	}
	v := engine.PlayerViewer(s.player)
	if want := engine.ViewFor(s.r.state, s.r.log, v); !reflect.DeepEqual(m.View.View, want) {
		s.t.Errorf("%s was sent a view that differs from ViewFor after %d events:\n got %+v\nwant %+v", s.player, len(s.r.log), m.View.View, want)
	}
	if got, want := s.events(), engine.Visible(s.r.state, s.r.log, v); !reflect.DeepEqual(got, want) {
		s.t.Errorf("%s knows %d events after %d, Visible gives %d", s.player, len(got), len(s.r.log), len(want))
	}
}

func newCheckedGame(t *testing.T, players int) *game {
	t.Helper()
	return newGameOf(t, players, true)
}

// newSink is a new phone for player ("" when not known yet): a refSink in a
// checked game. sink is what the test reads, phone what the room is given.
func (g *game) newSink(player engine.PlayerID) (sink *recSink, phone Sink) {
	if !g.checked {
		s := &recSink{}
		return s, s
	}
	s := &refSink{t: g.t, r: g.r, player: player}
	return &s.recSink, s
}

// nightWithABite plays night 1 with the wolf biting a villager, and the dawn,
// as in TestBroadcast_EachPhoneGetsOnlyItsOwnProjection. It returns the dead.
func (g *game) nightWithABite() engine.PlayerID {
	g.t.Helper()
	g.toNight1()
	wolf, victim := g.who(roles.Werewolf), g.who(roles.Villager)
	g.step(engine.DefaultSettings().NightStep) // the Guard's call passes
	g.must(g.cl[wolf].Submit(func(by engine.PlayerID) engine.Command {
		return roles.WolfVote{By: by, Target: victim}
	}))
	for i := 0; i < 2; i++ { // the pack's call, then the Seer's
		g.step(engine.DefaultSettings().NightStep)
	}
	g.step(engine.DefaultSettings().Announce)
	if g.snap().Players[victim].Alive {
		g.t.Fatalf("%s survived the bite", victim)
	}
	return victim
}

// rejoinOnNewPhone rejoins id with nickname and PIN on another phone.
func (g *game) rejoinOnNewPhone(id engine.PlayerID) *recSink {
	g.t.Helper()
	sink, phone := g.newSink(id)
	cl, creds, err := g.r.RejoinPIN(g.nick[id], g.cred[id].PIN, phone)
	if err != nil || cl.Player() != id {
		g.t.Fatalf("RejoinPIN(%s) = %v, %v", id, cl, err)
	}
	g.cl[id], g.sink[id], g.cred[id] = cl, sink, creds
	return sink
}

func TestViews_EveryMessageMatchesTheFullReplay_ThroughADeathTwoGamesAndRejoins(t *testing.T) {
	g := newCheckedGame(t, 5)
	g.deal()
	dead := g.nightWithABite()
	wolf, seer := g.who(roles.Werewolf), g.who(roles.Seer)
	for _, id := range g.snap().Seats {
		target := wolf
		if id == wolf {
			target = seer
		}
		if id != dead {
			g.must(g.cl[id].Submit(as(engine.Nominate{Voter: id, Target: target})))
		}
	}
	g.cl["p3"].Disconnect()
	g.rejoinOnNewPhone("p3")
	g.rejoinOnNewPhone(dead)
	g.forceEnded()
	g.must(g.cl["p1"].Submit(as(engine.NewGame{})))
	g.deal()
	g.toNight1()
	g.rejoinOnNewPhone("p2")
	if _, err := g.r.RejoinToken(g.cred["p4"].Token, &refSink{t: t, r: g.r, player: "p4"}); err != nil {
		t.Fatal(err)
	}
}

// The new phone of a player who comes back mid-game with nickname and PIN
// starts from exactly what the full replay gives: the view and every event the
// player may know, from the start of the log.
func TestRejoin_OtherPhoneMidGame_NewPhoneGetsExactlyTheFullReplay(t *testing.T) {
	for _, who := range []string{"alive", "dead"} {
		t.Run(who, func(t *testing.T) {
			g := newGame(t, 5)
			g.deal()
			dead := g.nightWithABite()
			id := dead
			if who == "alive" {
				id = g.who(roles.Werewolf)
			}
			g.cl[id].Disconnect()
			m := g.rejoinOnNewPhone(id).last()

			var view engine.View
			var events []engine.Event
			_ = g.r.do(func() error {
				v := engine.PlayerViewer(id)
				view, events = engine.ViewFor(g.r.state, g.r.log, v), engine.Visible(g.r.state, g.r.log, v)
				return nil
			})
			if !reflect.DeepEqual(m.View.View, view) {
				t.Fatalf("the new phone's view differs from ViewFor:\n got %+v\nwant %+v", m.View.View, view)
			}
			if m.Events == nil || !m.Events.Reset || !reflect.DeepEqual(m.Events.Events, events) {
				t.Fatalf("the new phone got events %+v, want a reset with the %d of Visible", m.Events, len(events))
			}
			if who == "dead" && !view.Full {
				t.Fatal("the dead player's new phone must see everything")
			}
		})
	}
}
