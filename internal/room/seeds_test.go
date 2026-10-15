package room

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"masoi/internal/engine"
)

func dealtRoles(s engine.State) map[engine.PlayerID]engine.RoleID {
	out := map[engine.PlayerID]engine.RoleID{}
	for id, p := range s.Players {
		out[id] = p.Role
	}
	return out
}

func TestSeeds_DealSeedComesFromTheRoomNotTheClient(t *testing.T) {
	g := newGame(t, 5)
	before := func() engine.State { return g.snap() }
	host := g.cl["p1"]
	cast := map[engine.RoleID]int{"Werewolf": 1, "Guard": 1, "Seer": 1, "Villager": 2}
	order := []engine.PlayerID{"p1", "p2", "p3", "p4", "p5"}
	g.must(host.Submit(as(engine.SetSeats{Order: order})))
	g.must(host.Submit(as(engine.ConfigureRoles{Roles: cast})))

	room, client := int64(7), int64(999)
	g.rnd.setSeed(room)
	ready := before()
	want := dealtRoles(engine.Fold(ready, mustDecide(t, ready, engine.DealRoles{Seed: room})...))
	other := dealtRoles(engine.Fold(ready, mustDecide(t, ready, engine.DealRoles{Seed: client})...))
	if reflect.DeepEqual(want, other) {
		t.Fatal("pick seeds that deal differently, or the test proves nothing")
	}

	g.must(host.Submit(as(engine.DealRoles{Seed: client}))) // the phone tries to choose
	if got := dealtRoles(g.snap()); !reflect.DeepEqual(got, want) {
		t.Fatalf("deal used the client's seed:\n got %v\nwant %v", got, want)
	}
}

func TestSeeds_ChiefTieBreakAndRandomChiefComeFromTheRoom(t *testing.T) {
	tests := []struct {
		name  string
		votes map[engine.PlayerID]engine.PlayerID
	}{
		{"tie", map[engine.PlayerID]engine.PlayerID{"p1": "p2", "p2": "p1"}},
		{"nobody voted", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := newGame(t, 5)
			g.deal()
			g.step(engine.DefaultSettings().RoleView)
			for voter, cand := range tc.votes {
				g.must(g.cl[voter].Submit(func(by engine.PlayerID) engine.Command {
					return engine.VoteChief{Voter: by, Candidate: cand}
				}))
			}
			ready := g.snap()
			// Find two seeds that elect different Chiefs.
			chiefFor := func(seed int64) engine.PlayerID {
				cmd := engine.PhaseTimeout{Phase: ready.Phase, Seq: ready.Seq, Seed: seed}
				return engine.Fold(ready, mustDecide(t, ready, cmd)...).Chief
			}
			var seed int64
			for seed = 2; chiefFor(seed) == chiefFor(1); seed++ {
				if seed > 200 {
					t.Fatal("no seed gives a different Chief")
				}
			}
			g.rnd.setSeed(seed)
			g.step(engine.DefaultSettings().ChiefElection)
			if got := g.snap().Chief; got != chiefFor(seed) {
				t.Fatalf("Chief = %s, want %s (from the room's seed)", got, chiefFor(seed))
			}
		})
	}
}

func TestSeeds_ClientCannotSendTimeoutsOrJoinThroughSubmit(t *testing.T) {
	g := newGame(t, 5)
	s := g.snap()
	for _, c := range []engine.Command{
		engine.PhaseTimeout{Phase: s.Phase, Seq: s.Seq, Seed: 5},
		engine.JoinRoom{Player: "p9", Nickname: "Zed"},
		engine.LeaveRoom{Player: "p2"},
	} {
		if err := g.cl["p2"].Submit(as(c)); !errorsIs(err, ErrForbidden) {
			t.Fatalf("%T: err = %v, want ErrForbidden", c, err)
		}
	}
}

func TestSeeds_AlwaysDrawnFromCryptoRand(t *testing.T) {
	r, err := New(Config{Code: "X"})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, ok := r.rnd.(CryptoRandom); !ok {
		t.Fatalf("default Randomness is %T, want CryptoRandom", r.rnd)
	}
	c := CryptoRandom{}
	seen := map[int64]bool{}
	for i := 0; i < 20; i++ {
		seen[c.Seed()] = true
		if pin := c.PIN(); len(pin) != 4 || strings.Trim(pin, "0123456789") != "" {
			t.Fatalf("PIN %q is not 4 digits", pin)
		}
	}
	if len(seen) < 19 || len(c.Token()) < 32 || sameToken(c) {
		t.Fatal("seeds and tokens do not look random")
	}

	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, _ := os.ReadFile(f)
		parsed, err := parser.ParseFile(token.NewFileSet(), f, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range parsed.Imports {
			if strings.Contains(imp.Path.Value, "math/rand") {
				t.Fatalf("%s imports %s: the room may only use crypto/rand", f, imp.Path.Value)
			}
		}
	}
}

func mustDecide(t *testing.T, s engine.State, c engine.Command) []engine.Event {
	t.Helper()
	evs, err := engine.Decide(s, c)
	if err != nil {
		t.Fatal(err)
	}
	return evs
}

func sameToken(c CryptoRandom) bool {
	first, second := c.Token(), c.Token()
	return first == second
}
