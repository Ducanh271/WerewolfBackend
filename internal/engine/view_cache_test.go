package engine

import (
	"reflect"
	"testing"
)

// twoGamesLog is a room's log over two games (audit K3, NewGame): five
// players join, are dealt, the game ends in a draw, the host starts a new
// one and the village wins it: the two summaries differ.
func twoGamesLog(t *testing.T) []Event {
	t.Helper()
	s := NewState()
	var log []Event
	do := func(c Command) {
		evs, err := Decide(s, c)
		if err != nil {
			t.Fatalf("Decide(%T) = %v", c, err)
		}
		log, s = append(log, evs...), Fold(s, evs...)
	}
	inject := func(e Event) { log, s = append(log, e), Evolve(s, e) }
	for i := 0; i < 5; i++ {
		do(JoinRoom{Player: playerID(i), Nickname: string(rune('A' + i))})
	}
	do(SetSeats{Order: seatOrder(5)})
	do(ConfigureRoles{Roles: map[RoleID]int{"Villager": 4, "Werewolf": 1}})
	for seed := int64(1); seed <= 2; seed++ {
		do(DealRoles{Seed: seed})
		if seed == 1 {
			inject(GameEnded{Outcome: Outcome{Draw: true}})
			do(NewGame{})
		} else {
			inject(GameEnded{Outcome: Outcome{Winner: TeamVillage, Winners: seatOrder(4)}})
		}
	}
	return log
}

// sameAsReplay fails unless the cache agrees with ViewFor and Visible for
// every player and an outsider at this point of log.
func sameAsReplay(t *testing.T, c *ViewCache, log []Event) {
	t.Helper()
	s := Fold(NewState(), log...)
	for _, id := range append(seatOrder(5), "outsider") {
		v := PlayerViewer(id)
		if got, want := c.ViewFor(s, log, v), ViewFor(s, log, v); !reflect.DeepEqual(got, want) {
			t.Fatalf("view of %s after %d events:\n got %+v\nwant %+v", id, len(log), got, want)
		}
		if got, want := c.Visible(s, log, v), Visible(s, log, v); !reflect.DeepEqual(got, want) {
			t.Fatalf("events of %s after %d events: got %d, want %d", id, len(log), len(got), len(want))
		}
	}
}

func TestViewCache_TwoGamesInOneRoom_MatchesTheFullReplayAfterEveryEvent(t *testing.T) {
	log := twoGamesLog(t)
	var c ViewCache
	for n := 1; n <= len(log); n++ {
		sameAsReplay(t, &c, log[:n])
	}
}

func TestViewCache_GivenAShorterLog_StartsOver(t *testing.T) {
	log := twoGamesLog(t)
	var c ViewCache
	sameAsReplay(t, &c, log)
	sameAsReplay(t, &c, log[:7])
	sameAsReplay(t, &c, log)
}
