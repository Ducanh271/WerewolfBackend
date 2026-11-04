package engine

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// MaxChatLen is the longest chat message accepted, in runes, after
// normalizing (rules §4 "Nhắn tin").
const MaxChatLen = 200

// zeroWidthJoiner glues emoji into one (a family, a profession), so it is the
// one invisible character a message keeps.
const zeroWidthJoiner = '‍'

// SendChat is a player's message on the room's one public channel (rules §4
// "Nhắn tin"). How often a player may send is the room's limit, not the
// engine's: it takes a clock.
type SendChat struct {
	CommandMarker
	By   PlayerID
	Text string
}

// ChatSent is one message on the public channel. Text is normalized but still
// untrusted: it is what a player typed. The speaker never reads it aloud.
type ChatSent struct {
	EventMarker
	By   PlayerID
	Text string
}

func decideSendChat(s State, cmd SendChat) ([]Event, error) {
	if err := chatWriter(s, cmd.By); err != nil {
		return nil, err
	}
	text := normalizeChat(cmd.Text)
	if text == "" {
		return nil, ErrChatEmpty
	}
	if n := utf8.RuneCountInString(text); n > MaxChatLen {
		return nil, fmt.Errorf("%w: %d characters", ErrChatTooLong, n)
	}
	return []Event{ChatSent{By: cmd.By, Text: text}}, nil
}

// chatWriter says whether a player may write now (rules §4 "Nhắn tin"). The
// phase comes first, so at night everybody gets the same answer whoever they
// are and whatever has happened to them (rules §3).
func chatWriter(s State, id PlayerID) error {
	switch s.Phase {
	case PhaseChiefElection, PhaseDay, PhaseNominationLock, PhaseTrial:
	default:
		return fmt.Errorf("%w: during %s", ErrChatClosed, s.Phase)
	}
	p, ok := s.Players[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	if !p.Alive {
		return fmt.Errorf("%w: %s", ErrNotAlive, id)
	}
	if s.Phase == PhaseTrial && id != s.Day.Defendant {
		return fmt.Errorf("%w: only the player on trial writes now", ErrChatClosed)
	}
	return nil
}

// normalizeChat is the form a message is checked and stored in (rules §4
// "Nhắn tin"): a line break or tab becomes one space, control and invisible characters
// are gone, NFC, no spaces at either end. It is "" when nothing visible is
// left. Removing comes before NFC: a removed character may have stood between
// a letter and its accent.
func normalizeChat(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inBreak := false
	for _, r := range strings.ToValidUTF8(s, "") {
		if isBreak(r) {
			if !inBreak {
				b.WriteByte(' ')
			}
			inBreak = true
			continue
		}
		inBreak = false
		if !isInvisible(r) {
			b.WriteRune(r)
		}
	}
	text := strings.TrimFunc(norm.NFC.String(b.String()), func(r rune) bool {
		return unicode.IsSpace(r) || r == zeroWidthJoiner
	})
	if !strings.ContainsFunc(text, isVisible) {
		return ""
	}
	return text
}

// isBreak reports a character that separates words without being a space, so
// dropping it would glue them together.
func isBreak(r rune) bool {
	switch r {
	case '\t', '\n', '\r', '\v', '\f', '\u0085', ' ', ' ':
		return true
	}
	return false
}

// isInvisible reports a character a message drops: control characters, format
// characters (zero-width spaces, text direction overrides, tag characters)
// and the blank-looking letters Unicode marks as ignorable (Hangul fillers).
func isInvisible(r rune) bool {
	if r == zeroWidthJoiner {
		return false
	}
	return unicode.In(r, unicode.Cc, unicode.Cf, unicode.Other_Default_Ignorable_Code_Point)
}

// isVisible reports a character that shows something by itself. A variation
// selector stays in a message (it picks an emoji's style) but shows nothing
// alone, like the joiner and spaces.
func isVisible(r rune) bool {
	return !unicode.IsSpace(r) && r != zeroWidthJoiner && !unicode.Is(unicode.Variation_Selector, r)
}
