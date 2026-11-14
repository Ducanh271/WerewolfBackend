package room

import (
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// Every view carries the room's clock (ViewMsg.Now), the clock its deadline
// is on, so phones with a wrong clock still end a step together.
func TestRoom_View_CarriesTheRoomClock(t *testing.T) {
	g := newGame(t, 5)
	g.step(3 * time.Second) // the clock alone sends nothing
	g.deal()

	v := g.sink["p2"].last().View
	if v == nil || !v.Now.Equal(g.clk.Now()) {
		t.Fatalf("view sent at %v, want the room clock %v", v.Now, g.clk.Now())
	}
	if v.Deadline.IsZero() || !v.Deadline.After(v.Now) {
		t.Errorf("deadline %v is not after now %v", v.Deadline, v.Now)
	}
}

// Rules §3: time passing is not a new screen. A secret night action seconds
// into the step still reaches nobody but the actor, though the clock in
// every view would be newer.
func TestRoom_SecretNightAction_LaterInTheStep_SendsNothingToAnyOtherPhone(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toNight1()
	for {
		step, ok := engine.CurrentNightRole(g.snap())
		if !ok {
			t.Fatal("night 1 had no Guard step")
		}
		if step == roles.Guard {
			break
		}
		g.step(engine.NightStepDuration(g.snap(), step))
	}
	guard, seer := g.who(roles.Guard), g.who(roles.Seer)
	g.step(5 * time.Second)
	g.expectSilent(roles.Guard, map[engine.PlayerID]bool{}, secretAction{guard, roles.GuardProtect{By: guard, Target: seer}})
}
