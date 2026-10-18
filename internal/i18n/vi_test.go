package i18n

import (
	"strings"
	"testing"
)

func TestVi_DawnAnnouncements(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		names []string
		want  string
	}{
		{"several dead, names as given", DawnSomeoneDied, []string{"An", "Binh"}, "Đêm qua, An, Binh đã chết."},
		{"protected", DawnProtected, nil, "Đêm qua không ai chết, có người đã được bảo vệ."},
		{"nobody died", DawnNobodyDied, nil, "Không có ai chết."},
		{"chief elected", ChiefElected, []string{"An"}, "An được bầu làm Trưởng làng."},
		{"chief succeeded", ChiefSucceeded, []string{"Binh"}, "Trưởng làng mới là Binh."},
		{"chief lost", ChiefLost, nil, "Trưởng làng không kịp chọn người kế nhiệm, danh hiệu đã mất."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Vi(tc.key, tc.names...); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVi_DawnNeverStatesCauseOrRole(t *testing.T) {
	for _, key := range []string{DawnSomeoneDied, DawnProtected, DawnNobodyDied} {
		text := strings.ToLower(Vi(key, "An"))
		for _, banned := range []string{"cắn", "độc", "sói", "phù thủy", "thợ săn", "vai"} {
			if strings.Contains(text, banned) {
				t.Errorf("%s mentions %q: %s", key, banned, text)
			}
		}
	}
}

// TestVi_NewVoiceLines (audit [L3], docs/voice-lines.md): the 9 new keys plus
// the 2 added on top of the audit, text as approved.
func TestVi_NewVoiceLines(t *testing.T) {
	tests := []struct{ key, want string }{
		{RoleViewStarted, "Mời mọi người xem vai của mình. Nhớ che màn hình nhé."},
		{ChiefElectionStarted, "Mời cả làng bầu Trưởng làng."},
		{NightStarted, "Trời đã tối, tất cả nhắm mắt lại."},
		{DawnStarted, "Trời đã sáng, mọi người mở mắt."},
		{DayStarted, "Cả làng bắt đầu thảo luận và đề cử."},
		{VerdictStarted, "Hết giờ bào chữa. Cả làng bỏ phiếu: treo hay tha."},
	}
	for _, tc := range tests {
		if got := Vi(tc.key); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.key, got, tc.want)
		}
	}
	if got := Vi(TrialStarted, "An"); got != "An lên giàn bào chữa." {
		t.Errorf("TrialStarted = %q", got)
	}
	if got := Vi(HunterTimedOut, "Binh"); got != "Hết giờ, Binh không bắn ai." {
		t.Errorf("HunterTimedOut = %q", got)
	}
}

// TestVi_NightCallName (rules §3): the calling name of every role that has a
// night turn, and no name for a step that belongs to no role.
func TestVi_NightCallName(t *testing.T) {
	tests := []struct{ role, want string }{
		{"Cupid", "Cupid"},
		{"Visitor", "Nguyệt Nữ"},
		{"Guard", "Bảo vệ"},
		{"CursedWolf", "Sói nguyền"},
		{"WolfSeer", "Sói tiên tri"},
		{"Werewolf", "Bầy sói"}, // never "Sói": that is the card name, not the night call
		{"Witch", "Phù thủy"},
		{"Seer", "Tiên tri"},
		{"Detective", "Thám tử"},
	}
	for _, tc := range tests {
		if got, ok := NightCallName(tc.role); !ok || got != tc.want {
			t.Errorf("NightCallName(%s) = %q, %v; want %q, true", tc.role, got, ok, tc.want)
		}
	}
	for _, step := range []string{"PhoneCheck", "LoversMeet", "Hunter", "RustyKnight", "Villager"} {
		if _, ok := NightCallName(step); ok {
			t.Errorf("NightCallName(%s) should have no calling name", step)
		}
	}
	if got := Vi(NightStepCalled, "Bầy sói"); got != "Bầy sói hãy mở mắt, thực hiện xong thì nhắm mắt lại." {
		t.Errorf("NightStepCalled sentence = %q", got)
	}
}

func TestVi_PhoneCheck_VoiceLineAndBothNotices(t *testing.T) {
	tests := []struct{ key, want string }{
		{NightPhoneCheck, "Mọi người mở mắt xem điện thoại, rồi nhắm mắt lại."},
		{NightLoversMeet, "Cặp đôi hãy mở mắt nhìn nhau."},
		{NoticeNotCursed, "Bạn không bị nguyền."},
		{NoticeCursed, "Bạn đã bị nguyền."},
	}
	for _, tc := range tests {
		if got := Vi(tc.key); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestVi_LoveNoticesAndGrief(t *testing.T) {
	if got := Vi(NoticeNoLove); got != "Bạn không ở trong cặp đôi." {
		t.Errorf("no love notice = %q", got)
	}
	if got := Vi(NoticeInLove, "An", "Tiên tri"); got != "Bạn yêu An, vai Tiên tri." {
		t.Errorf("in love notice = %q", got)
	}
	if Vi(NightGrief) == NightGrief || Vi(NightPackDoomed, "An") != "An không thể vote đêm nay." {
		t.Error("grief or pack notice missing")
	}
}

func TestCards_TextOfRulesNineB(t *testing.T) {
	c, ok := Card("RustyKnight")
	if !ok || c.Name != "Hiệp sĩ kiếm gỉ" || c.Team != "Dân" ||
		c.Description != "Nếu bị Sói cắn chết, con Sói còn sống gần nhất bên trái bạn trúng độc và chết vào đêm sau." {
		t.Fatalf("Rusty knight card = %+v", c)
	}
	if len(CardOrder()) != len(roleCards) {
		t.Fatalf("the order lists %d roles, the cards %d", len(CardOrder()), len(roleCards))
	}
	name, text := StatusCard(StatusLovers, "Chi")
	if name != "Cặp đôi cùng phe" || text != "Bạn và Chi yêu nhau. Một người chết thì người kia chết theo. Vẫn chơi cho phe của mình." {
		t.Fatalf("lovers card = %q / %q", name, text)
	}
}
