package roles_test

import (
	"testing"
	"time"

	"masoi/internal/engine"
	"masoi/internal/i18n"
	"masoi/internal/roles"

	. "masoi/internal/engine/enginetest"
)

// chiefBitten plays a night where the wolves bite the Chief, Phuc.
func chiefBitten(t *testing.T) *Game {
	g := newTable(t)
	g.Election().Vote("An", "Phuc").Close(1)
	g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc").ExpectAnnouncement(i18n.DawnSomeoneDied)
	return g
}

func TestChiefDies_DeadChiefHasFifteenSecondsToNameASuccessor(t *testing.T) {
	g := chiefBitten(t)
	g.ExpectPhase(engine.PhaseChiefSuccession).ExpectNoChief()
	if got := g.State().Settings.ChiefSuccession; got != 15*time.Second {
		t.Fatalf("succession lasts %v, want 15s", got)
	}
	g.NameSuccessor("Phuc", "Em")
	g.ExpectChief("Em").ExpectPhase(engine.PhaseDawn)
}

func TestChiefDies_NoAnswerInTime_TitleIsLost(t *testing.T) {
	g := chiefBitten(t)
	g.TimeOut()
	g.ExpectNoChief().ExpectPhase(engine.PhaseDawn)
}

func TestChiefDies_SuccessorMustBeLiving(t *testing.T) {
	g := chiefBitten(t)
	g.DoExpectErr(engine.NameSuccessor{By: "Phuc", Successor: "Phuc"}, engine.ErrInvalidTarget)
	g.DoExpectErr(engine.NameSuccessor{By: "Phuc", Successor: "Nobody"}, engine.ErrInvalidTarget)
}

func TestChiefDies_OnlyTheDeadChiefNamesTheSuccessor(t *testing.T) {
	g := chiefBitten(t)
	g.DoExpectErr(engine.NameSuccessor{By: "Em", Successor: "Em"}, engine.ErrNotYourTurn)
}

func TestChiefDies_NamingOutsideSuccessionIsRejected(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("An", "Phuc").Close(1)
	g.DoExpectErr(engine.NameSuccessor{By: "Phuc", Successor: "Em"}, engine.ErrWrongPhase)
}

func TestChiefDies_LateTimeoutAfterNamingIsIgnored(t *testing.T) {
	g := chiefBitten(t)
	stale := engine.PhaseTimeout{Phase: g.State().Phase, Seq: g.State().Seq}
	g.NameSuccessor("Phuc", "Em")
	g.Do(stale)
	g.ExpectChief("Em")
}

func TestChiefSurvivesTheNight_KeepsTheTitle(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("An", "Phuc").Close(1)
	g.Night().Wolf("An").Bites("Giang").Wolf("Binh").Bites("Giang")
	g.Dawn().ExpectDead("Giang")
	g.ExpectChief("Phuc").ExpectPhase(engine.PhaseDawn)
}

func TestChiefGuardedFromTheBite_KeepsTheTitle(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("An", "Phuc").Close(1)
	g.Night().Guard("Dung").Protects("Phuc").Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
	g.ExpectChief("Phuc")
}

func TestChiefMayBeAWolf_StaysAWolf(t *testing.T) {
	g := newTable(t)
	g.Election().Vote("An", "An").Close(1)
	g.ExpectChief("An").ExpectRole("An", roles.Werewolf)
}
