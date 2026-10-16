package ws

import (
	"encoding/json"
	"testing"

	"masoi/internal/engine"
	"masoi/internal/roles"
)

// docs/protocol.md: wolfVote with target "" is the vote "Không cắn ai".
func TestCommandTable_WolfVoteWithEmptyTargetIsTheNoBiteVote(t *testing.T) {
	build, err := commandTable["wolfVote"](json.RawMessage(`{"target":""}`))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := build("p1").(roles.WolfVote)
	if !ok || got.By != "p1" || got.Target != engine.NoBite {
		t.Fatalf("got %+v, want a WolfVote by p1 with NoBite", build("p1"))
	}
}

// A missing target is invalid; only an explicit "" means "Không cắn ai".
func TestCommandTable_WolfVoteWithoutTargetIsRejected(t *testing.T) {
	for _, raw := range []string{`{}`, `{"target":null}`} {
		if _, err := commandTable["wolfVote"](json.RawMessage(raw)); err == nil {
			t.Errorf("wolfVote %s accepted, want a bad_message error", raw)
		}
	}
}

// docs/protocol.md: cupidPair carries the two players as "a" and "b".
func TestCommandTable_CupidPairCarriesTwoPlayers(t *testing.T) {
	build, err := commandTable["cupidPair"](json.RawMessage(`{"a":"p1","b":"p2"}`))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := build("p3").(roles.CupidPair)
	if !ok || got.By != "p3" || got.A != "p1" || got.B != "p2" {
		t.Fatalf("got %+v, want a CupidPair by p3 of p1 and p2", build("p3"))
	}
	if _, err := commandTable["cupidPair"](json.RawMessage(`{"a":"p1","seed":1}`)); err == nil {
		t.Fatal("an unknown field was accepted")
	}
}

// docs/protocol.md: configureSound maps "oneSpeaker"/"allPhones" to the
// engine's SoundMode; anything else is bad_message.
func TestCommandTable_ConfigureSound(t *testing.T) {
	build, err := commandTable["configureSound"](json.RawMessage(`{"mode":"allPhones"}`))
	if err != nil {
		t.Fatal(err)
	}
	got, ok := build("p1").(engine.ConfigureSound)
	if !ok || got.Mode != engine.AllPhonesSpeak {
		t.Fatalf("got %+v, want ConfigureSound{AllPhonesSpeak}", build("p1"))
	}
	if _, err := commandTable["configureSound"](json.RawMessage(`{"mode":"megaphone"}`)); err == nil {
		t.Fatal("an unknown sound mode was accepted")
	}
}
