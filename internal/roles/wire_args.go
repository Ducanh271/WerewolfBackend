package roles

import "masoi/internal/engine"

// Shapes of the args of the roles' wire commands (docs/protocol.md). The
// sender is never in the args: the transport fills it in. The type names show
// in the "wrong field type" refusals, so they are part of the wire text.
type (
	noArgs     struct{}
	targetArgs struct {
		Target engine.PlayerID `json:"target"`
	}
	pairArgs struct {
		A engine.PlayerID `json:"a"`
		B engine.PlayerID `json:"b"`
	}
)
