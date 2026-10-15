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
