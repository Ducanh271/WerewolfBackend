package engine

import (
	"strings"
	"unicode/utf8"
)

// MaxNicknameLen is the longest nickname accepted, in runes, after trimming.
// Not in rules.md: picked so every phone and the speaker's announcements
// stay readable (audit S4).
const MaxNicknameLen = 20

// vietnameseLetters holds every Vietnamese letter outside plain ASCII, lower
// and upper case. Restricting nicknames to this plus ASCII letters, digits
// and spaces closes both tricks audit S4 found: invisible characters hidden
// inside a nickname, and letters from another script that only look like a
// Latin one (e.g. Cyrillic "А" for "A").
const vietnameseLetters = "" +
	"àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ" +
	"ÀÁẢÃẠĂẰẮẲẴẶÂẦẤẨẪẬÈÉẺẼẸÊỀẾỂỄỆÌÍỈĨỊÒÓỎÕỌÔỒỐỔỖỘƠỜỚỞỠỢÙÚỦŨỤƯỪỨỬỮỰỲÝỶỸỴĐ"

// validNickname reports whether a trimmed nickname may be accepted: 1–20
// characters, each one an ASCII letter or digit, a space, or a Vietnamese
// letter (rules §2 does not give exact limits; this is the room's own rule).
func validNickname(s string) bool {
	if s == "" || utf8.RuneCountInString(s) > MaxNicknameLen {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == ' ':
		case strings.ContainsRune(vietnameseLetters, r):
		default:
			return false
		}
	}
	return true
}
