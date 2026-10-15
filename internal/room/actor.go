package room

import (
	"reflect"

	"masoi/internal/engine"
)

// actorOf finds the player a command claims to come from: its By, Voter or
// Player field. Commands with none (SetSeats, ConfigureRoles...) have no
// actor of their own.
func actorOf(c engine.Command) (engine.PlayerID, bool) {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Struct {
		return "", false
	}
	for _, name := range []string{"By", "Voter", "Player"} {
		f := v.FieldByName(name)
		if f.IsValid() && f.Type() == reflect.TypeOf(engine.PlayerID("")) {
			return engine.PlayerID(f.String()), true
		}
	}
	return "", false
}
