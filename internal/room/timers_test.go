package room

import (
	"errors"
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

func TestNight_EveryStepLastsItsFullTime_EvenIfTheRoleActsEarly(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toNight1()
	step := engine.DefaultSettings().NightStep
	guard := g.who(roles.Guard)

	first := g.snap()
	if role, _ := engine.CurrentNightRole(first); role != roles.Guard {
		t.Fatalf("first call = %s, want Guard", role)
	}
	g.must(g.cl[guard].Submit(func(by engine.PlayerID) engine.Command {
		return roles.GuardProtect{By: by, Target: by}
	}))

	g.step(step - time.Second)
	if got := g.snap(); got.Seq != first.Seq {
		t.Fatalf("the Guard acted early and the step ended %s early", time.Second)
	}
	g.step(time.Second)
	if role, _ := engine.CurrentNightRole(g.snap()); role != roles.Werewolf {
		t.Fatalf("after the full %s the call = %s, want the pack", step, role)
	}
}

func TestTimers_RoleViewAndChiefElectionNeverFireEarly(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	set := engine.DefaultSettings()
	g.step(set.RoleView - time.Second)
	if g.snap().Phase != engine.PhaseRoleView {
		t.Fatal("role view ended early")
	}
	g.step(time.Second)
	if g.snap().Phase != engine.PhaseChiefElection {
		t.Fatal("role view did not end on time")
	}
}

func TestDay_SevenMinuteClockEndsTheDayWhenNobodyIsNominated(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	day := engine.DefaultSettings().Day
	g.step(day - time.Second)
	if g.snap().Phase != engine.PhaseDay {
		t.Fatal("the day ended early")
	}
	g.step(time.Second)
	if got := g.snap().Phase; got != engine.PhaseDusk {
		t.Fatalf("phase = %s, want Dusk", got)
	}
}

// nominateAll puts p1 up for everybody; p1 puts up p2.
func (g *game) nominateAll() {
	g.t.Helper()
	for id, cl := range g.cl {
		target := engine.PlayerID("p1")
		if id == "p1" {
			target = "p2"
		}
		g.must(cl.Submit(func(by engine.PlayerID) engine.Command {
			return engine.Nominate{Voter: by, Target: target}
		}))
	}
}

func TestDay_ClockPausesOnlyDuringDefenseAndVerdict(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	set := engine.DefaultSettings()

	g.step(2 * time.Minute)
	g.nominateAll()
	if g.snap().Phase != engine.PhaseNominationLock {
		t.Fatalf("phase = %s, want the nomination countdown", g.snap().Phase)
	}
	g.step(set.NominationLock - time.Second)
	if g.snap().Phase != engine.PhaseNominationLock {
		t.Fatal("the 10 s lock ended early")
	}
	g.step(time.Second)
	if g.snap().Phase != engine.PhaseTrial {
		t.Fatalf("phase = %s, want Trial after the 10 s lock", g.snap().Phase)
	}
	left := set.Day - 2*time.Minute - set.NominationLock // the countdown still counts
	g.step(set.Defense)
	if g.snap().Phase != engine.PhaseVerdict {
		t.Fatalf("phase = %s, want Verdict", g.snap().Phase)
	}
	g.step(set.Verdict)
	if got := g.snap().Phase; got != engine.PhaseDay {
		t.Fatalf("phase = %s, want Day again after the spare", got)
	}
	var remaining time.Duration
	_ = g.r.do(func() error { remaining = g.r.day.remaining; return nil })
	if remaining != left {
		t.Fatalf("day clock = %s, want %s (paused during defense and verdict only)", remaining, left)
	}

	g.step(left - time.Second)
	if g.snap().Phase != engine.PhaseDay {
		t.Fatal("the day ended early")
	}
	g.step(time.Second)
	if got := g.snap().Phase; got == engine.PhaseDay {
		t.Fatal("the day clock ran out and nothing happened")
	}
}

func TestDay_ClockRunningOutDuringTheCountdownPutsTheTopNomineeOnTrial(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toDay()
	set := engine.DefaultSettings()

	g.step(set.Day - 5*time.Second)
	g.nominateAll() // the 10 s lock would run past the 7 minutes
	g.step(5 * time.Second)
	s := g.snap()
	if s.Phase != engine.PhaseTrial || !s.Day.ClockOut || s.Day.Defendant != "p1" {
		t.Fatalf("phase=%s clockOut=%v defendant=%s, want the clock to put p1 on trial", s.Phase, s.Day.ClockOut, s.Day.Defendant)
	}
	g.step(set.Defense)
	g.step(set.Verdict) // spared, and the clock is out: the day is over
	if got := g.snap().Phase; got != engine.PhaseDusk {
		t.Fatalf("phase = %s, want Dusk", got)
	}
}

func TestTimers_StaleTimeoutsAreIgnored(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.toNight1()
	before, logLen := g.snap(), len(g.eventLog())

	stale := []engine.PhaseTimeout{
		{Phase: engine.PhaseNight, Seq: before.Seq - 1},     // an earlier step
		{Phase: engine.PhaseDay, Seq: before.Seq},           // the wrong phase
		{Phase: engine.PhaseRoleView, Seq: before.Seq - 20}, // long gone
	}
	for _, to := range stale {
		_ = g.r.do(func() error {
			if err := g.r.execute(to); err != nil {
				t.Errorf("stale %+v: err = %v, want none", to, err)
			}
			return nil
		})
	}
	if after := g.snap(); after.Seq != before.Seq || after.Phase != before.Phase || len(g.eventLog()) != logLen {
		t.Fatal("a stale timeout changed the game")
	}
}

func TestTimers_OldTimersAreStoppedWhenThePhaseMovesOn(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.step(engine.DefaultSettings().RoleView)
	if n := g.clk.Pending(); n != 1 {
		t.Fatalf("%d timers pending in the Chief election, want exactly 1", n)
	}
}

func TestTimers_HostChangesTheTimesBeforeTheDealAndTheRoomUsesThem(t *testing.T) {
	g := newGame(t, 5)
	g.must(g.cl["p1"].Submit(as(engine.ConfigureTimers{Timers: engine.Timers{RoleView: 12 * time.Second, Announce: 6 * time.Second}})))
	if err := g.cl["p1"].Submit(as(engine.ConfigureTimers{Timers: engine.Timers{Day: time.Second}})); !errors.Is(err, engine.ErrTimerOutOfRange) {
		t.Fatalf("bad value: %v, want ErrTimerOutOfRange", err)
	}
	g.deal()
	g.step(12*time.Second - time.Second)
	if g.snap().Phase != engine.PhaseRoleView {
		t.Fatal("role view ended before the host's 12 s")
	}
	g.step(time.Second)
	if g.snap().Phase != engine.PhaseChiefElection {
		t.Fatal("role view did not end after the host's 12 s")
	}
	if got := g.sink["p2"].last().View.Deadline.Sub(g.clk.Now()); got != engine.DefaultSettings().ChiefElection {
		t.Fatalf("deadline in %s", got)
	}
}
