package roles_test

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/store"
)

// A stored log must give back the same game (docs/database.md "Event
// encoding"): every event of whole simulated games goes through the store's
// encoding and back, and folding the decoded log gives the same State.
func TestCodec_RoundTripsEverySimulatedGame(t *testing.T) {
	games := 40
	if testing.Short() {
		games = 8
	}
	seen := map[string]bool{}
	for i := 0; i < games; i++ {
		seed := uint64(simBaseSeed + 7000 + i)
		g, err := runSimGame(seed, false)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}

		decoded := make([]engine.Event, 0, len(g.log))
		for n, e := range g.log {
			back, name, data := roundTrip(t, e)
			if !reflect.DeepEqual(e, back) {
				t.Fatalf("seed %d, event %d %s: %s\n before %#v\n after  %#v", seed, n, name, data, e, back)
			}
			decoded = append(decoded, back)
			seen[name] = true
		}
		if !reflect.DeepEqual(g.s, engine.Fold(engine.NewState(), decoded...)) {
			t.Fatalf("seed %d: the decoded log folds to another state", seed)
		}
	}
	t.Logf("%d games, %d of %d event types seen", games, len(seen), len(engine.EventTypes()))
}

// Events the simulation may not reach still encode and decode.
func TestCodec_EveryRegisteredEventRoundTrips(t *testing.T) {
	for _, e := range engine.EventTypes() {
		if back, name, data := roundTrip(t, e); !reflect.DeepEqual(e, back) {
			t.Errorf("%s: %s\n before %#v\n after  %#v", name, data, e, back)
		}
	}
}

func roundTrip(t *testing.T, e engine.Event) (back engine.Event, name string, data []byte) {
	t.Helper()
	name, data, err := store.EncodeEvent(e)
	if err != nil {
		t.Fatalf("encode %T: %v", e, err)
	}
	back, err = store.DecodeEvent(name, data)
	if err != nil {
		t.Fatalf("decode %s %s: %v", name, data, err)
	}
	return back, name, data
}

func TestCodec_RejectsUnknownAndUnregistered(t *testing.T) {
	type notRegistered struct{ engine.EventMarker }
	tests := []struct {
		name string
		run  func() error
	}{
		{"unknown type", func() error { _, err := store.DecodeEvent("NoSuchEvent", []byte(`{}`)); return err }},
		{"unknown field", func() error {
			_, err := store.DecodeEvent("PlayerLeft", []byte(`{"Player":"p1","Extra":1}`))
			return err
		}},
		{"bad json", func() error { _, err := store.DecodeEvent("PlayerLeft", []byte(`{`)); return err }},
		{"unregistered event", func() error { _, _, err := store.EncodeEvent(notRegistered{}); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.run() == nil {
				t.Fatal("no error")
			}
		})
	}
}
