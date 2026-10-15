package ws

import (
	"reflect"
	"testing"

	"masoi/internal/engine"
)

// Rules §9b: a role card, an entry of the public role list and an extra card
// each have the same JSON keys whatever they hold, so a screen cannot tell one
// role or one state from another by the shape.
func TestEncode_CardsHaveTheSameKeysForEveryone(t *testing.T) {
	tests := []struct {
		name  string
		cards []any
		want  []string
	}{
		{"role card", []any{engine.RoleCard{}, engine.RoleCard{Name: "Sói", Team: "Sói", Goal: "g", Description: "d"}},
			[]string{"description", "goal", "name", "team"}},
		{"role list entry", []any{engine.RoleListCard{}, engine.RoleListCard{Role: "Seer", Name: "Tiên tri", Description: "d"}},
			[]string{"description", "name", "role"}},
		{"extra card", []any{engine.StatusCard{}, engine.StatusCard{Kind: "cursed", Name: "Bị nguyền", Text: "t"}},
			[]string{"kind", "name", "text"}},
	}
	for _, tc := range tests {
		for _, c := range tc.cards {
			if got := keySet(toTree(c).(map[string]any), ""); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%s %+v has keys %v, want %v", tc.name, c, got, tc.want)
			}
		}
	}
}
