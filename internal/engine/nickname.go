package engine

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// MaxNicknameLen is the longest nickname accepted, in runes, after trimming
// (rules §2 "Nickname, rời phòng và đá người"; audit S4).
const MaxNicknameLen = 20

// vietnameseLetters holds every Vietnamese letter outside plain ASCII, lower
// and upper case. Restricting nicknames to this plus ASCII letters, digits
// and spaces closes both tricks audit S4 found: invisible characters hidden
// inside a nickname, and letters from another script that only look like a
// Latin one (e.g. Cyrillic "А" for "A").
const vietnameseLetters = "" +
	"àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ" +
	"ÀÁẢÃẠĂẰẮẲẴẶÂẦẤẨẪẬÈÉẺẼẸÊỀẾỂỄỆÌÍỈĨỊÒÓỎÕỌÔỒỐỔỖỘƠỜỚỞỠỢÙÚỦŨỤƯỪỨỬỮỰỲÝỶỸỴĐ"

// NormalizeNickname is the form a nickname is checked and stored in (rules §2
// "Nickname, rời phòng và đá người"): NFC, no spaces at either end, and runs
// of spaces collapsed to one, so two names that look the same are the same
// string. Only the plain space is collapsed: any other blank stays and is
// refused by validNickname.
func NormalizeNickname(s string) string {
	s = strings.TrimSpace(norm.NFC.String(s))
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// validNickname reports whether a normalized nickname may be accepted (rules §2
// "Nickname, rời phòng và đá người"): 1–20 characters, each one an ASCII
// letter or digit, a space, or a Vietnamese letter.
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
