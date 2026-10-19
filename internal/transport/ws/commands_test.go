package ws

import (
	"strings"
	"testing"

	"masoi/internal/engine"
)

// commandsRole is a role that only declares commands, for building tables
// from a hand-made registry.
type commandsRole struct {
	id       engine.RoleID
	commands []engine.CommandSpec
}

func (r commandsRole) ID() engine.RoleID          { return r.id }
func (commandsRole) BaseTeam() engine.Team        { return engine.TeamVillage }
func (commandsRole) NightCall() *engine.NightCall { return nil }
func (commandsRole) Decide(engine.State, engine.Command) ([]engine.Event, bool, error) {
	return nil, false, nil
}
func (commandsRole) Evolve(*engine.State, engine.Event) bool { return false }
func (r commandsRole) Commands() []engine.CommandSpec        { return r.commands }

type pingArgs struct{}

func pingCommand(name string) engine.CommandSpec {
	return engine.WireCommand(name, func(pingArgs, engine.PlayerID) engine.Command { return engine.NewGame{} })
}

// A clash of command names stops the server from starting: New returns the
// error instead of a handler.
func TestCommandTable_ClashingNamesAreRefused(t *testing.T) {
	tests := []struct {
		name     string
		registry map[engine.RoleID]engine.Role
		want     string
	}{
		{"a role reuses a core command", map[engine.RoleID]engine.Role{
			"Mime": commandsRole{"Mime", []engine.CommandSpec{pingCommand("nominate")}},
		}, `command "nominate" is declared by both the engine and role Mime`},
		{"two roles share a name", map[engine.RoleID]engine.Role{
			"Bard": commandsRole{"Bard", []engine.CommandSpec{pingCommand("sing")}},
			"Mime": commandsRole{"Mime", []engine.CommandSpec{pingCommand("sing")}},
		}, `command "sing" is declared by both role Bard and role Mime`},
		{"one role twice", map[engine.RoleID]engine.Role{
			"Bard": commandsRole{"Bard", []engine.CommandSpec{pingCommand("sing"), pingCommand("sing")}},
		}, `command "sing" is declared by both role Bard and role Bard`},
		{"a command without a name", map[engine.RoleID]engine.Role{
			"Bard": commandsRole{"Bard", []engine.CommandSpec{pingCommand("")}},
		}, `role Bard declares an incomplete command ""`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newCommandTable(tc.registry)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCommandTable_RolesWithoutClashAreAdded(t *testing.T) {
	table, err := newCommandTable(map[engine.RoleID]engine.Role{
		"Bard": commandsRole{"Bard", []engine.CommandSpec{pingCommand("sing")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := table["sing"]; !ok {
		t.Fatal("the role's command is missing")
	}
	if len(table) != len(coreCommands)+1 {
		t.Fatalf("table has %d commands, want the %d core ones and sing", len(table), len(coreCommands))
	}
}
