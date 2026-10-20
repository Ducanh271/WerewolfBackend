package i18n

import (
	"strings"
	"testing"
)

func TestErrorText(t *testing.T) {
	tests := []struct {
		name            string
		code, situation string
		args            []string
		want            string
		own             bool
	}{
		{"own text", "nickname_taken", "", nil, "Trong phòng đã có người tên này, bạn chọn tên khác nhé.", true},
		{"second text for its situation", "wrong_phase", SituationJoin, nil, "Ván đã bắt đầu nên không vào mới được; nếu bạn đã ở trong ván, hãy vào lại bằng tên và PIN.", true},
		{"usual text without a situation", "wrong_phase", "", nil, "Bước này đã qua rồi, bạn không làm được nữa.", true},
		{"a situation the code does not have", "bad_credentials", SituationJoin, nil, "Tên hoặc PIN không đúng, bạn kiểm tra lại nhé.", true},
		{"token rejoin", "bad_credentials", SituationToken, nil, "Máy này không còn vào được phòng, bạn vào lại bằng tên và PIN nhé.", true},
		{"placeholder", "pin_locked", "", []string{"42"}, "Bạn nhập sai PIN quá nhiều lần, thử lại sau 42 giây nhé.", true},
		{"alias reads the notice", "blocked", "", nil, Vi(NightBlocked), true},
		{"buggy client", "bad_message", "", nil, Vi(ErrorGeneric), false},
		{"unknown code", "something_new", "", nil, Vi(ErrorGeneric), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, own := ErrorText(tc.code, tc.situation, tc.args...)
			if got != tc.want || own != tc.own {
				t.Fatalf("got %q, %v; want %q, %v", got, own, tc.want, tc.own)
			}
		})
	}
}

// not_your_turn and wrong_phase reach any player at night, whatever their
// role and state (rules §3): they never name a role or say who is alive.
func TestErrorText_SentToAnybodyTellsNothing(t *testing.T) {
	banned := []string{"chết", "sống", "chặn", "ghé", "cắn", "bầy"}
	for _, c := range roleCards {
		banned = append(banned, strings.ToLower(c.Name))
	}
	for _, name := range nightCallNames {
		banned = append(banned, strings.ToLower(name))
	}
	for _, code := range []string{"not_your_turn", "wrong_phase"} {
		text, _ := ErrorText(code, "")
		low := strings.ToLower(text)
		for _, b := range banned {
			if strings.Contains(low, b) {
				t.Errorf("%s says %q: %s", code, b, text)
			}
		}
	}
}

func TestErrorTemplate_KeepsPlaceholders(t *testing.T) {
	if got, _ := ErrorTemplate("pin_locked", ""); !strings.Contains(got, "%s") {
		t.Fatalf("got %q", got)
	}
}

func TestErrorKeys(t *testing.T) {
	keys := ErrorKeys()
	has := map[string]bool{}
	for _, k := range keys {
		has[k] = true
	}
	for _, want := range []string{"blocked", "wrong_phase", "wrong_phase.join", "bad_credentials.token"} {
		if !has[want] {
			t.Errorf("ErrorKeys misses %s: %v", want, keys)
		}
	}
	if has["generic"] {
		t.Error("the generic text is no code's own")
	}
}
