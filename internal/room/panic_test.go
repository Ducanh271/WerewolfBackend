package room

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"masoi/internal/engine"
)

// A bug in the engine or in a role (a panic) must cost one command, not every
// game on the server: the room rejects that command with ErrInternal, keeps its
// state and its timers, and the other rooms never notice.

// panicRole is a role only these tests know. Its one command panics in Decide;
// it is never in a role list, so no game ever calls it at night.
type panicRole struct{}

type panicCmd struct {
	engine.CommandMarker
	By engine.PlayerID
}

const panicRoleID engine.RoleID = "PanicTestRole"

func init() { engine.Register(panicRole{}) }

func (panicRole) ID() engine.RoleID                       { return panicRoleID }
func (panicRole) BaseTeam() engine.Team                   { return engine.TeamVillage }
func (panicRole) NightCall() *engine.NightCall            { return nil }
func (panicRole) Evolve(*engine.State, engine.Event) bool { return false }
func (panicRole) Decide(_ engine.State, c engine.Command) ([]engine.Event, bool, error) {
	if _, ok := c.(panicCmd); ok {
		panic("a bug in a role")
	}
	return nil, false, nil
}

// syncBuffer is a log destination safe to read while the room writes.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestRoom_PanicInACommand_IsRejectedAsInternal_RoomAndOtherRoomsGoOn(t *testing.T) {
	logs := &syncBuffer{}
	m := NewManager(ManagerConfig{Clock: NewFakeClock(time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC)), Random: &fixedRandom{seed: 1},
		Logger: slog.New(slog.NewTextHandler(logs, nil))})
	t.Cleanup(m.Close)
	broken, _, err := m.Create(engine.InPerson)
	if err != nil {
		t.Fatal(err)
	}
	other, _, err := m.Create(engine.InPerson)
	if err != nil {
		t.Fatal(err)
	}

	an, _, err := broken.Join("An", "", &recSink{})
	if err != nil {
		t.Fatal(err)
	}
	var before engine.State
	_ = broken.do(func() error { before = broken.state; return nil })

	err = an.Submit(func(by engine.PlayerID) engine.Command { return panicCmd{By: by} })
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("err = %v, want ErrInternal", err)
	}
	if !strings.Contains(logs.String(), "panic in room") || !strings.Contains(logs.String(), broken.Code()) {
		t.Fatalf("the panic must be logged with the room code, logs:\n%s", logs.String())
	}

	// The room still works and the command changed nothing.
	var after engine.State
	_ = broken.do(func() error { after = broken.state; return nil })
	if after.Seq != before.Seq || len(after.Players) != len(before.Players) {
		t.Fatal("the command that panicked changed the state")
	}
	if _, _, err := broken.Join("Binh", "", &recSink{}); err != nil {
		t.Fatalf("the room stopped working after the panic: %v", err)
	}

	// The other room never noticed.
	if _, _, err := other.Join("Chi", "", &recSink{}); err != nil {
		t.Fatalf("another room was hit by the panic: %v", err)
	}
	if m.Len() != 2 {
		t.Fatalf("rooms open = %d, want 2", m.Len())
	}
}

// A panic in a job nobody waits for (a timer's) is recovered the same way, and
// the room's timers go on firing.
func TestRoom_PanicInATimerJob_RoomKeepsRunningItsTimers(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	g.r.post(func() { panic("a bug in a timer") })
	g.step(engine.DefaultSettings().RoleView)
	if s := g.snap(); s.Phase != engine.PhaseChiefElection {
		t.Fatalf("after the panic the role view timer did not fire: phase %s", s.Phase)
	}
}
