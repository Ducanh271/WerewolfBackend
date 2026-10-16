package roles_test

import (
	"testing"
	"time"

	"masoi/internal/engine"
)

func TestDay_NominationCountdown_ChangingAChoiceRestartsIt(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	stale := engine.PhaseTimeout{Phase: g.State().Phase, Seq: g.State().Seq}

	d.Nominate("Binh", "Hoa").Nominate("Chi", "Hoa") // two changes: Hoa now leads
	d.ExpectPhase(engine.PhaseNominationLock)
	g.Do(stale) // the first countdown's timeout no longer matches
	d.ExpectPhase(engine.PhaseNominationLock)

	d.TimeOut().ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Hoa")
}

// Rules §4: "mỗi lần có người đổi thì đếm lại từ đầu" — resending the exact
// same pick is not a change, so it must not restart the countdown.
func TestDay_NominationCountdown_ResendingTheSameChoiceDoesNotRestartIt(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	first := engine.PhaseTimeout{Phase: g.State().Phase, Seq: g.State().Seq}

	for i := 0; i < 3; i++ {
		d.Nominate("An", "Giang") // the exact same pick, resent again and again
	}
	d.ExpectPhase(engine.PhaseNominationLock)

	g.Do(first) // the original countdown's timeout must still be the live one
	d.ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Giang")
}

func TestDay_NominationCountdown_IsTenSecondsByDefault(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	if got := g.State().Settings.NominationLock; got != 10*time.Second {
		t.Fatalf("countdown = %v, want 10s", got)
	}
}

func TestDay_NominationCountdown_NobodyOnTrialUntilItRunsOut(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock) // still nobody on trial
	d.TimeOut().ExpectOnTrial("Giang")
}

func TestDay_NominationCountdown_EveryoneBacksOut_DiscussionContinues(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	d.NominateNobody("An").ExpectPhase(engine.PhaseDay) // nobody is nominated any more
}

func TestDay_NominationCountdown_DayClockRunsOut_LeaderGoesOnTrialForTheLastTime(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	d.ExpectPhase(engine.PhaseNominationLock)
	s := g.State()
	g.Do(engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq, DayClock: true})
	d.ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Giang")
	d.TimeOut()
	for _, name := range all[:5] {
		d.Spare(name)
	}
	d.TimeOut().ExpectPhase(engine.PhaseDusk) // spared, but the clock is out: night
	g.ExpectAlive("Giang")
}

func TestDay_NominationCountdown_DayClockKeepsRunning_SoChangingVotesCannotStallTheDay(t *testing.T) {
	g, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	for i := 0; i < 5; i++ { // restarting the countdown does not pause the clock
		d.Nominate("Binh", "Hoa").Nominate("Binh", "Giang")
	}
	s := g.State()
	g.Do(engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq, DayClock: true})
	d.ExpectPhase(engine.PhaseTrial).ExpectOnTrial("Giang")
}

func TestDay_NominationCountdown_PlayersCanStillSkip(t *testing.T) {
	_, d := day(t)
	d.Nominate("An", "Giang")
	nobodyExcept(d, "An")
	for _, name := range all[1:8] {
		d.Skip(name)
	}
	d.ExpectPhase(engine.PhaseDusk)
}
