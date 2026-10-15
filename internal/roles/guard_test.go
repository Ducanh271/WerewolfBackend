package roles_test

import (
	"errors"
	"testing"
	"time"

	"masoi/internal/engine"
	. "masoi/internal/engine/enginetest"
	"masoi/internal/i18n"
	"masoi/internal/roles"
)

func TestNight_GuardProtectsTheBittenPlayer_AnnouncesProtected(t *testing.T) {
	g := newTable(t)
	n := g.Night()
	n.Guard("Dung").Protects("Phuc")
	n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
}

func TestNight_GuardProtectsSomeoneElse_BittenPlayerDies(t *testing.T) {
	g := newTable(t)
	n := g.Night()
	n.Guard("Dung").Protects("Giang")
	n.Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc").ExpectAnnouncement(i18n.DawnSomeoneDied)
}

func TestNight_GuardSleeps_BittenPlayerDies(t *testing.T) {
	g := newTable(t)
	g.Night().Wolf("An").Bites("Phuc").Wolf("Binh").Bites("Phuc")
	g.Dawn().ExpectDead("Phuc")
}

func TestNight_GuardProtectsHimself_SurvivesTheBite(t *testing.T) {
	g := newTable(t)
	n := g.Night()
	n.Guard("Dung").Protects("Dung")
	n.Wolf("An").Bites("Dung").Wolf("Chi").Bites("Dung")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnProtected)
}

func TestNight_GuardProtectsButNobodyIsBitten_AnnouncesNobodyDied(t *testing.T) {
	g := newTable(t)
	g.Night().Guard("Dung").Protects("Phuc")
	g.Dawn().ExpectNobodyDied().ExpectAnnouncement(i18n.DawnNobodyDied)
}

func TestNight_GuardChoosesTwice_SecondChoiceIsRejected(t *testing.T) {
	g := newTable(t)
	g.Night().Guard("Dung").Protects("Phuc").Guard("Dung").CannotProtect("Giang", engine.ErrAlreadyActed)
}

func TestNight_GuardChoiceIsRejectedWhenInvalid(t *testing.T) {
	tests := []struct {
		name string
		do   func(n *Night)
	}{
		{"not the guard", func(n *Night) { n.Guard("Phuc").CannotProtect("Giang", engine.ErrNotYourTurn) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) { tc.do(newTable(t).Night()) })
	}
}

// laterNight moves a finished game to the Guard's turn on the given night. The
// day phase machine arrives with T7, so the kit cannot play the days between
// nights yet; the scenario folds the night events itself.
func laterNight(s engine.State, night int) engine.State {
	return engine.Fold(s,
		engine.NightStarted{Night: night},
		engine.NightStepCalled{Role: roles.Guard, Duration: time.Second},
	)
}

func protect(t *testing.T, s engine.State, by, target string) (engine.State, error) {
	t.Helper()
	evs, err := engine.Decide(s, roles.GuardProtect{By: engine.PlayerID(by), Target: engine.PlayerID(target)})
	return engine.Fold(s, evs...), err
}

func TestNight_GuardCannotProtectTheSamePlayerTwoNightsInARow(t *testing.T) {
	tests := []struct {
		name   string
		second string
		want   error
	}{
		{"same player again", "Phuc", engine.ErrSameGuardTarget},
		{"another player", "Giang", nil},
		{"himself", "Dung", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			g.Night().Guard("Dung").Protects("Phuc")
			g.Dawn()
			_, err := protect(t, laterNight(g.State(), 2), "Dung", tc.second)
			if (tc.want == nil && err != nil) || (tc.want != nil && !errors.Is(err, tc.want)) {
				t.Fatalf("second night: err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestNight_GuardSkipsANight_CanProtectTheOldTargetAgain(t *testing.T) {
	g := newTable(t)
	g.Night().Guard("Dung").Protects("Phuc")
	g.Dawn()
	s := engine.Fold(g.State(), engine.NightStarted{Night: 2}) // night 2: the Guard sleeps
	s = laterNight(s, 3)
	if _, err := protect(t, s, "Dung", "Phuc"); err != nil {
		t.Fatalf("night 3 after a skipped night: %v", err)
	}
}
