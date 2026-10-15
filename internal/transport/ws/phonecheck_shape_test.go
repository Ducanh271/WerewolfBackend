package ws

import (
	"reflect"
	"sort"
	"testing"

	"masoi/internal/engine"
)

func keySet(m map[string]any, prefix string) []string {
	var keys []string
	for k, v := range m {
		keys = append(keys, prefix+k)
		if sub, ok := v.(map[string]any); ok {
			keys = append(keys, keySet(sub, prefix+k+".")...)
		}
	}
	sort.Strings(keys)
	return keys
}

// Rules §3: every phone check notice has the same JSON keys whatever the
// answer; empty values are sent, not omitted.
func TestEncode_PhoneCheckNoticeHasTheSameKeysForEveryone(t *testing.T) {
	notices := map[string]engine.PhoneChecked{
		"nothing":       {Player: "p1"},
		"cursed":        {Player: "p2", Cursed: true},
		"in love":       {Player: "p3", Love: engine.LoveNotice{InLove: true, Partner: "p4", PartnerRole: "Seer", Colour: engine.LinkStriped}},
		"cursed + love": {Player: "p5", Cursed: true, Love: engine.LoveNotice{InLove: true, Partner: "p6", PartnerRole: "Villager", Colour: engine.LinkWhite}},
	}
	var want []string
	for name, n := range notices {
		got := keySet(toTree(n).(map[string]any), "")
		if want == nil {
			want = got
			if len(want) != 7 { // player, cursed, love and its four parts
				t.Fatalf("%s: keys %v", name, got)
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s has keys %v, others have %v", name, got, want)
		}
	}
	empty := toTree(notices["nothing"]).(map[string]any)["love"].(map[string]any)
	if empty["inLove"] != false || empty["partner"] != "" || empty["partnerRole"] != "" || empty["colour"] != "" {
		t.Errorf("empty love part = %v, want inLove false and empty strings", empty)
	}

	views := []engine.PhoneCheckView{{}, {Cursed: true, Love: engine.LoveNotice{InLove: true, Partner: "p4", PartnerRole: "Seer", Colour: engine.LinkBlack}}}
	a, b := keySet(toTree(views[0]).(map[string]any), ""), keySet(toTree(views[1]).(map[string]any), "")
	if !reflect.DeepEqual(a, b) {
		t.Errorf("phoneCheck view keys differ: %v and %v", a, b)
	}
}
