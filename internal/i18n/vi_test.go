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
