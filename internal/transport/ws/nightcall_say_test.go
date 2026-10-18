package ws

import (
	"testing"
	"time"

	"masoi/internal/engine"
)

// Rules §3: the text read when a role's turn comes at night names the role,
// not a player, and follows the exact calling name of the table (docs/voice-
// lines.md), Werewolf included ("Bầy sói", never "Sói").
func TestEncode_NightStepCalled_Say(t *testing.T) {
	tests := []struct {
		role engine.RoleID
		say  string
	}{
		{"Cupid", "Cupid hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"Visitor", "Nguyệt Nữ hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"Guard", "Bảo vệ hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"CursedWolf", "Sói nguyền hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"WolfSeer", "Sói tiên tri hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"Werewolf", "Bầy sói hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"Witch", "Phù thủy hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"Seer", "Tiên tri hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
		{"Detective", "Thám tử hãy mở mắt, thực hiện xong thì nhắm mắt lại."},
	}
	for _, tc := range tests {
		t.Run(string(tc.role), func(t *testing.T) {
			tree := eventTree(engine.NightStepCalled{Role: tc.role, Duration: 30 * time.Second}, nil)
			if got := tree["say"]; got != tc.say {
				t.Errorf("say = %v, want %q", got, tc.say)
			}
		})
	}
}

// The phone check and the lovers looking at each other already have their own
// voice line (PhoneCheckAnnounced, LoversMeetAnnounced); NightStepCalled for
// those steps must add no second line, or the two would be read back to back.
func TestEncode_NightStepCalled_NoSayForEngineOwnSteps(t *testing.T) {
	for _, step := range []engine.RoleID{engine.StepPhoneCheck, engine.StepLoversMeet} {
		tree := eventTree(engine.NightStepCalled{Role: step, Duration: 5 * time.Second}, nil)
		if say, has := tree["say"]; has {
			t.Errorf("%s: say = %v, want no say field", step, say)
		}
	}
}

// Rules §3 "gọi giả": NightStepCalled carries nothing but the role and the
// duration, so the exact same event is produced whether the holder is alive,
// dead or absent; the rendered line can only depend on those two fields, and
// is therefore identical for a real call and a fake one.
func TestEncode_NightStepCalled_SameLineForFakeAndRealHolder(t *testing.T) {
	real := engine.NightStepCalled{Role: "Werewolf", Duration: 30 * time.Second}
	fake := engine.NightStepCalled{Role: "Werewolf", Duration: 30 * time.Second}
	if real != fake {
		t.Fatalf("a real and a fake call differ: %+v vs %+v", real, fake)
	}
	if got := eventTree(real, nil)["say"]; got != eventTree(fake, nil)["say"] {
		t.Fatalf("rendered text differs between a real and a fake call")
	}
}
