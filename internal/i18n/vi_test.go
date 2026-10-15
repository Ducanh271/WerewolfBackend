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
		{"protected", DawnProtected, nil, "Không có ai chết, được bảo vệ."},
		{"nobody died", DawnNobodyDied, nil, "Không có ai chết."},
		{"chief elected", ChiefElected, []string{"An"}, "An đã được bầu làm Trưởng làng."},
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
