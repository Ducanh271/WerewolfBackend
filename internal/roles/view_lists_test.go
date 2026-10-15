package roles_test

import (
	"fmt"
	"reflect"
	"testing"

	"masoi/internal/engine"

	. "masoi/internal/engine/enginetest"
)

// The order of a list of players can say more than its content: the deaths of
// a night in pipeline order tell bitten from poisoned (rules §4). So every list
// of players inside an event a living player gets must be in seating order (as
// the seats were right after the event), or be a list already read out in an
// announcement.

var playerListType = reflect.TypeOf([]engine.PlayerID(nil))

// playerLists returns every []PlayerID inside v, nested structs included.
func playerLists(v reflect.Value) [][]engine.PlayerID {
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return playerLists(v.Elem())
	case reflect.Struct:
		var out [][]engine.PlayerID
		for i := 0; i < v.NumField(); i++ {
			out = append(out, playerLists(v.Field(i))...)
		}
		return out
	case reflect.Slice:
		if v.Type() == playerListType {
			return [][]engine.PlayerID{v.Interface().([]engine.PlayerID)}
		}
		var out [][]engine.PlayerID
		for i := 0; i < v.Len(); i++ {
			out = append(out, playerLists(v.Index(i))...)
		}
		return out
	}
	return nil
}

// inSeatingOrder: every id is seated and they come in seat order.
func inSeatingOrder(seats, ids []engine.PlayerID) bool {
	at := map[engine.PlayerID]int{}
	for i, id := range seats {
		at[id] = i
	}
	last := -1
	for _, id := range ids {
		i, ok := at[id]
		if !ok || i <= last {
			return false
		}
		last = i
	}
	return true
}

// playerListLeak checks the events of log the viewer may know about and
// returns a description of the first list that breaks the rule, "" if none.
func playerListLeak(s engine.State, log []engine.Event, viewer engine.PlayerID) string {
	visible := engine.Visible(s, log, engine.PlayerViewer(viewer))
	var announced [][]engine.PlayerID
	at := engine.NewState()
	next := 0
	for _, e := range log {
		at = engine.Evolve(at, e)
		if next >= len(visible) || !reflect.DeepEqual(e, visible[next]) {
			continue
		}
		next++
		for _, list := range playerLists(reflect.ValueOf(e)) {
			if inSeatingOrder(at.Seats, list) || containsList(announced, list) {
				continue
			}
			return fmt.Sprintf("%T has the players %v, not in seating order %v and never announced", e, list, at.Seats)
		}
		if _, ok := e.(interface{ AnnouncementKey() string }); ok {
			announced = append(announced, playerLists(reflect.ValueOf(e))...)
		}
	}
	if next != len(visible) {
		return fmt.Sprintf("could not place visible event %d (%T) in the log", next, visible[next])
	}
	return ""
}

func containsList(lists [][]engine.PlayerID, list []engine.PlayerID) bool {
	for _, l := range lists {
		if reflect.DeepEqual(l, list) {
			return true
		}
	}
	return false
}

// expectNoPlayerListLeak runs the check for every living player.
func expectNoPlayerListLeak(t *testing.T, g *Game) {
	t.Helper()
	for _, id := range g.State().Seats {
		if g.State().Players[id].Alive {
			if leak := playerListLeak(g.State(), g.Events(), id); leak != "" {
				t.Fatalf("secret leak to %s: %s", id, leak)
			}
		}
	}
}

// The Hunter bitten and another player poisoned the same night: the morning
// reads both names in seat order, and nothing the living get afterwards (the
// Hunter's reveal, his pause to shoot) gives the order in which they died.
func TestView_HunterDiesWithAPoisonedPlayer_TheLivingNeverLearnWhoWasBittenAndWhoPoisoned(t *testing.T) {
	g := hunterTable(t)
	packBites(g.Night(), "Phuc").Witch("Giang").Poisons("Chi") // Chi sits before Phuc
	g.Dawn().ExpectDead("Chi", "Phuc")
	g.ExpectPhase(engine.PhaseHunterShot)
	expectNoPlayerListLeak(t, g)
	g.Hunter("Phuc").Shoots("Hoa")
	expectNoPlayerListLeak(t, g)

	// The chain is still known to the engine (and to the dead): it goes on
	// after the shot exactly as before.
	g.ExpectDead("Chi", "Phuc", "Hoa").ExpectPhase(engine.PhaseDawn)
}

func TestView_HangedHunterShootsAWolf_ListsStayInSeatingOrder(t *testing.T) {
	g := hunterTable(t)
	g.Night()
	g.Dawn().ExpectNobodyDied()
	g.Day().Hanged("Phuc")
	expectNoPlayerListLeak(t, g)
	g.Hunter("Phuc").Shoots("Binh")
	expectNoPlayerListLeak(t, g)
}

// The check itself must catch a list out of seating order.
func TestView_PlayerListCheck_CatchesAListOutOfSeatingOrder(t *testing.T) {
	seats := []engine.PlayerID{"a", "b", "c"}
	if !inSeatingOrder(seats, []engine.PlayerID{"a", "c"}) || inSeatingOrder(seats, []engine.PlayerID{"c", "a"}) {
		t.Fatal("inSeatingOrder is wrong")
	}
	type chain struct {
		engine.EventMarker
		Dead []engine.PlayerID
	}
	lists := playerLists(reflect.ValueOf(engine.Event(chain{Dead: []engine.PlayerID{"c", "a"}})))
	if len(lists) != 1 || inSeatingOrder(seats, lists[0]) {
		t.Fatalf("playerLists = %v, want the one list [c a] found and refused", lists)
	}
}
