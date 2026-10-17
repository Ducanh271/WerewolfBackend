package room

import (
	"errors"
	"strings"
	"testing"
	"time"

	"masoi/internal/engine"
)

func newManager(t *testing.T, cfg ManagerConfig) (*Manager, *FakeClock) {
	t.Helper()
	clk := NewFakeClock(time.Date(2026, 10, 15, 20, 0, 0, 0, time.UTC))
	cfg.Clock = clk
	m := NewManager(cfg)
	t.Cleanup(m.Close)
	return m, clk
}

func TestManager_CodesAreFourCharactersWithoutConfusingOnes(t *testing.T) {
	for i := 0; i < 500; i++ {
		code := randomCode()
		if len(code) != CodeLength {
			t.Fatalf("code %q is not %d characters", code, CodeLength)
		}
		if strings.ContainsAny(code, "O0I1L") || strings.Trim(code, CodeAlphabet) != "" {
			t.Fatalf("code %q uses a character outside the alphabet", code)
		}
	}
}

func TestManager_CreateThenGetFindsTheRoom_IgnoringCase(t *testing.T) {
	m, _ := newManager(t, ManagerConfig{NewCode: func() string { return "K7QP" }})
	r, _, err := m.Create(engine.InPerson)
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.Get(" k7qp ")
	if err != nil || got != r {
		t.Fatalf("Get = %v, %v; want the created room", got, err)
	}
	if _, err := m.Get("ZZZZ"); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("unknown code: err = %v, want ErrRoomNotFound", err)
	}
}

func TestManager_CodeCollisionDrawsAnotherCode(t *testing.T) {
	codes := []string{"AAAA", "AAAA", "BBBB"}
	m, _ := newManager(t, ManagerConfig{NewCode: func() string {
		c := codes[0]
		codes = codes[1:]
		return c
	}})
	a, _, _ := m.Create(engine.InPerson)
	b, _, err := m.Create(engine.InPerson)
	if err != nil {
		t.Fatal(err)
	}
	if a.Code() != "AAAA" || b.Code() != "BBBB" {
		t.Fatalf("codes = %s, %s; want AAAA, BBBB", a.Code(), b.Code())
	}
}

func TestManager_OnlineRoomsAreRefused(t *testing.T) {
	m, _ := newManager(t, ManagerConfig{})
	if _, _, err := m.Create(engine.Online); !errors.Is(err, engine.ErrModeUnsupported) {
		t.Fatalf("err = %v, want ErrModeUnsupported", err)
	}
}

func TestManager_RoomLimit(t *testing.T) {
	m, _ := newManager(t, ManagerConfig{MaxRooms: 1})
	if _, _, err := m.Create(engine.InPerson); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.Create(engine.InPerson); !errors.Is(err, ErrTooManyRooms) {
		t.Fatalf("err = %v, want ErrTooManyRooms", err)
	}
}

func TestManager_RoomIdleForTwoHoursIsDeleted(t *testing.T) {
	m, clk := newManager(t, ManagerConfig{NewCode: func() string { return "K7QP" }})
	r, _, _ := m.Create(engine.InPerson)

	clk.Advance(IdleTimeout - 2*sweepEvery)
	if _, err := m.Get("K7QP"); err != nil {
		t.Fatalf("room deleted too early: %v", err)
	}

	// Somebody joins: the idle time starts again.
	if _, _, err := r.Join("An", "", &recSink{}); err != nil {
		t.Fatal(err)
	}
	clk.Advance(IdleTimeout - 2*sweepEvery)
	if _, err := m.Get("K7QP"); err != nil {
		t.Fatalf("a room with recent activity was deleted: %v", err)
	}

	clk.Advance(3 * sweepEvery)
	if _, err := m.Get("K7QP"); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("idle room still there: err = %v", err)
	}
	select {
	case <-r.Done():
	case <-time.After(time.Second):
		t.Fatal("the deleted room was not stopped")
	}
}
