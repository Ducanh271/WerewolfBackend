package room

import (
	"errors"
	"fmt"
	"testing"

	"masoi/internal/engine"
)

// seatOrder is p1..pn, the join order newGame uses.
func seatOrder(n int) []engine.PlayerID {
	order := make([]engine.PlayerID, n)
	for i := range order {
		order[i] = engine.PlayerID(fmt.Sprintf("p%d", i+1))
	}
	return order
}

// TestJoin_HostKeyClaimsHost_RegardlessOfJoinOrder is the S8 regression test:
// the room's creator is the host even if someone else's phone connects
// first.
func TestJoin_HostKeyClaimsHost_RegardlessOfJoinOrder(t *testing.T) {
	r, err := New(Config{Code: "K7QP", HostKey: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)

	first, _, err := r.Join("Binh", "", &recSink{}) // connects first, has no key
	if err != nil {
		t.Fatal(err)
	}
	creator, _, err := r.Join("An", "secret", &recSink{}) // the actual creator, connects second
	if err != nil {
		t.Fatal(err)
	}

	order := []engine.PlayerID{first.Player(), creator.Player()}
	if err := first.Submit(as(engine.SetSeats{Order: order})); !errorsIs(err, ErrNotHost) {
		t.Fatalf("the first connection must not be host: %v", err)
	}
	if err := creator.Submit(as(engine.SetSeats{Order: order})); err != nil {
		t.Fatalf("the creator must be host: %v", err)
	}
}

func TestJoin_WrongOrMissingHostKey_ClaimsNothing(t *testing.T) {
	r, err := New(Config{Code: "K7QP", HostKey: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)

	wrong, _, err := r.Join("An", "not-it", &recSink{})
	if err != nil {
		t.Fatal(err)
	}
	none, _, err := r.Join("Binh", "", &recSink{})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []*Client{wrong, none} {
		if err := c.Submit(as(engine.SetSeats{Order: []engine.PlayerID{wrong.Player(), none.Player()}})); !errorsIs(err, ErrNotHost) {
			t.Fatalf("%s: err = %v, want ErrNotHost", c.Player(), err)
		}
	}
}

// TestJoin_HostKeyIsSingleUse: once the title has moved on, presenting the
// creator's key again does not take it back (decided when T27a was planned).
func TestJoin_HostKeyIsSingleUse(t *testing.T) {
	g := newGame(t, 2) // p1 joined with the room's hostKey in newGame
	g.must(g.cl["p1"].TransferHost(g.cl["p2"].Player()))

	// A third phone presents the same key the creator used: it must not
	// become host, now or ever.
	late, _, err := g.r.Join("Chi", g.hostKey, &recSink{})
	if err != nil {
		t.Fatal(err)
	}
	if err := late.Submit(as(engine.SetSeats{Order: []engine.PlayerID{"p1", "p2", late.Player()}})); !errorsIs(err, ErrNotHost) {
		t.Fatalf("a spent hostKey must not grant host: %v", err)
	}
	if got := g.host(); got != "p2" {
		t.Fatalf("host = %s, want p2 (unchanged)", got)
	}
}

func TestClient_KickPlayer_RemovesTargetKeepsTheRest(t *testing.T) {
	g := newGame(t, 5)
	g.must(g.cl["p1"].Submit(as(engine.SetSeats{Order: seatOrder(5)})))
	cast := map[engine.RoleID]int{"Werewolf": 1, "Villager": 4}
	g.must(g.cl["p1"].Submit(as(engine.ConfigureRoles{Roles: cast})))

	g.must(g.cl["p1"].KickPlayer("p3"))

	if _, ok := g.snap().Players["p3"]; ok {
		t.Fatal("p3 should be gone")
	}
	if got := g.snap().Seats; len(got) != 4 {
		t.Fatalf("seats = %v, want the other 4 kept", got)
	}
	if len(g.snap().Settings.Roles) != 2 {
		t.Fatal("the role list must survive a kick")
	}
	if n := len(g.sink["p3"].evicted); n != 1 {
		t.Fatalf("p3's connection must be evicted once, got %d", n)
	}
	if _, err := g.r.RejoinToken(g.cred["p3"].Token, &recSink{}); !errorsIs(err, ErrBadCredentials) {
		t.Fatalf("a kicked player's token must be revoked: %v", err)
	}
}

func TestClient_KickPlayer_Rejected(t *testing.T) {
	g := newGame(t, 5)

	if err := g.cl["p2"].KickPlayer("p3"); !errorsIs(err, ErrNotHost) {
		t.Fatalf("a non-host kicking: %v, want ErrNotHost", err)
	}
	if err := g.cl["p1"].KickPlayer("p1"); !errorsIs(err, ErrForbidden) {
		t.Fatalf("the host kicking themselves: %v, want ErrForbidden", err)
	}
	if err := g.cl["p1"].KickPlayer("p9"); !errors.Is(err, engine.ErrUnknownPlayer) {
		t.Fatalf("kicking a stranger: %v, want ErrUnknownPlayer", err)
	}

	g.deal()
	if err := g.cl["p1"].KickPlayer("p2"); !errors.Is(err, engine.ErrWrongPhase) {
		t.Fatalf("kicking after the deal: %v, want ErrWrongPhase", err)
	}
}

func TestClient_TransferHost_GivesTheTitleAwayNoConsentNeeded(t *testing.T) {
	g := newGame(t, 3)

	g.must(g.cl["p1"].TransferHost("p2"))

	if got := g.host(); got != "p2" {
		t.Fatalf("host = %s, want p2", got)
	}
	if err := g.cl["p1"].Submit(as(engine.SetSeats{Order: seatOrder(3)})); !errorsIs(err, ErrNotHost) {
		t.Fatalf("the old host must lose the powers: %v", err)
	}
	g.must(g.cl["p2"].Submit(as(engine.SetSeats{Order: seatOrder(3)})))
}

func TestClient_TransferHost_WorksMidGame(t *testing.T) {
	g := newGame(t, 5)
	g.deal() // host powers are over per rules §2, but the title itself can move

	g.must(g.cl["p1"].TransferHost("p3"))

	if got := g.host(); got != "p3" {
		t.Fatalf("host = %s, want p3", got)
	}
}

func TestClient_TransferHost_Rejected(t *testing.T) {
	g := newGame(t, 3)

	if err := g.cl["p2"].TransferHost("p3"); !errorsIs(err, ErrNotHost) {
		t.Fatalf("a non-host transferring: %v, want ErrNotHost", err)
	}
	if err := g.cl["p1"].TransferHost("p1"); !errorsIs(err, ErrForbidden) {
		t.Fatalf("transferring to self: %v, want ErrForbidden", err)
	}
	if err := g.cl["p1"].TransferHost("p9"); !errors.Is(err, engine.ErrUnknownPlayer) {
		t.Fatalf("transferring to a stranger: %v, want ErrUnknownPlayer", err)
	}
}
