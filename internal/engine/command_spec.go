package engine

// CommandSpec is one command a client may send (docs/protocol.md "Commands"):
// its wire name, the shape of its args and how it becomes a Command. The
// engine only holds the description; the transport decodes the JSON.
type CommandSpec struct {
	// Name is the command's name on the wire, e.g. "guardProtect".
	Name string
	// NewArgs returns a pointer to a fresh args struct. Its json tags are the
	// only keys the args may have.
	NewArgs func() any
	// Build makes the command from the decoded args (what NewArgs returned)
	// and the sender, who never comes from the message. It is only called
	// once the args passed decoding and their Validate, if any.
	Build func(args any, by PlayerID) Command
}

// ArgsValidator is implemented by an args struct that checks itself once
// decoded, e.g. wolfVote's "target" must be present. The error text goes to
// the client as a bad_message.
type ArgsValidator interface {
	Validate() error
}

// WireCommand makes a CommandSpec from the args struct A and a builder.
func WireCommand[A any](name string, build func(a A, by PlayerID) Command) CommandSpec {
	return CommandSpec{
		Name:    name,
		NewArgs: func() any { return new(A) },
		Build:   func(args any, by PlayerID) Command { return build(*args.(*A), by) },
	}
}

// CommandRole is implemented by a role whose holder sends commands of its
// own. The transport adds them to its command table at start-up, so a new
// role needs no edit outside its own file.
type CommandRole interface {
	Commands() []CommandSpec
}
