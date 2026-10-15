package enginetest

import (
	"fmt"
	"strings"
	"testing"

	"masoi/internal/engine"
)

// The kit plays the real Werewolf, Villager, Seer and Guard (it imports
// internal/roles); the kit's own tests only register stand-ins for roles that
// are not built yet.

type fakeRole struct {
	id   engine.RoleID
	team engine.Team
}

func (r fakeRole) ID() engine.RoleID            { return r.id }
func (r fakeRole) BaseTeam() engine.Team        { return r.team }
func (r fakeRole) NightCall() *engine.NightCall { return nil }
func (r fakeRole) Decide(engine.State, engine.Command) ([]engine.Event, bool, error) {
	return nil, false, nil
}
func (r fakeRole) Evolve(*engine.State, engine.Event) bool { return false }

type minRole struct {
	fakeRole
	min int
}

func (r minRole) MinPlayers() int { return r.min }

func init() {
	engine.Register(minRole{fakeRole{Detective, engine.TeamVillage}, 9})
}

// failure is what the recording TB panics with, so a test can assert that the
// kit failed a scenario and with which message.
type failure struct{ msg string }

// recordingTB stands in for *testing.T and turns Fatalf into a panic.
type recordingTB struct {
	testing.TB
}

func (recordingTB) Helper() {}
func (recordingTB) Fatalf(format string, args ...any) {
	panic(failure{fmt.Sprintf(format, args...)})
}

// failsWith runs f and checks the kit failed the scenario with a message that
// contains want.
func failsWith(t *testing.T, want string, f func(tb testing.TB)) {
	t.Helper()
	defer func() {
		t.Helper()
		r := recover()
		fail, ok := r.(failure)
		if !ok {
			t.Fatalf("expected the kit to fail with %q, got %v", want, r)
		}
		if !strings.Contains(fail.msg, want) {
			t.Fatalf("kit failed with %q, want it to contain %q", fail.msg, want)
		}
	}()
	f(recordingTB{t})
}

// fakeDriver is a minimal night/day machine: a night ends by killing whoever
// the "bitten" field names, and a hanging kills the hanged player.
type fakeDriver struct{ bitten engine.PlayerID }

type kill struct {
	engine.CommandMarker
	who engine.PlayerID
}

func (fakeDriver) StartNight(engine.State) ([]engine.Command, error) { return nil, nil }
func (d fakeDriver) EndNight(engine.State) ([]engine.Command, error) {
	if d.bitten == "" {
		return nil, nil
	}
	return []engine.Command{kill{who: d.bitten}}, nil
}
func (fakeDriver) NobodyHanged(engine.State) ([]engine.Command, error) { return nil, nil }
func (fakeDriver) Hang(_ engine.State, who engine.PlayerID) ([]engine.Command, error) {
	return []engine.Command{kill{who: who}}, nil
}

// killRole handles the kill command and announces deaths, standing in for the
// real night resolution.
type killRole struct{ fakeRole }

type died struct {
	engine.EventMarker
	who engine.PlayerID
}

func (died) AnnouncementKey() string { return "dawn.someone_died" }

func (killRole) Decide(_ engine.State, c engine.Command) ([]engine.Event, bool, error) {
	k, ok := c.(kill)
	if !ok {
		return nil, false, nil
	}
	return []engine.Event{died{who: k.who}}, true, nil
}

func (killRole) Evolve(s *engine.State, e engine.Event) bool {
	d, ok := e.(died)
	if !ok {
		return false
	}
	s.Players[d.who].Alive = false
	return true
}

func init() { engine.Register(killRole{fakeRole{"KillStandIn", engine.TeamVillage}}) }
