package engine

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"masoi/internal/i18n"
)

// chatState is a table of living "alive" and "defendant" and dead "dead".
func chatState(phase Phase) State {
	s := NewState()
	s.Phase = phase
	s.Day.Defendant = "defendant"
	s.Players["alive"] = &Player{ID: "alive", Alive: true}
	s.Players["defendant"] = &Player{ID: "defendant", Alive: true}
	s.Players["dead"] = &Player{ID: "dead"}
	return s
}

// Rules §4 "Nhắn tin", one row per phase and per writer. A phase added later
// has no row and fails here until someone decides its chat.
func TestChat_WhoMayWriteInEachPhase(t *testing.T) {
	closed := map[string]error{"alive": ErrChatClosed, "defendant": ErrChatClosed, "dead": ErrChatClosed}
	open := map[string]error{"alive": nil, "defendant": nil, "dead": ErrNotAlive}
	rows := map[Phase]map[string]error{
		PhaseLobby:           closed,
		PhaseRoleView:        closed,
		PhaseChiefElection:   open,
		PhaseNight:           closed,
		PhaseDawn:            closed,
		PhaseDay:             open,
		PhaseNominationLock:  open,
		PhaseTrial:           {"alive": ErrChatClosed, "defendant": nil, "dead": ErrNotAlive},
		PhaseVerdict:         closed,
		PhaseDusk:            closed,
		PhaseHunterShot:      closed,
		PhaseChiefSuccession: closed,
		PhaseEnded:           closed,
	}
	for phase := PhaseLobby; phase.String() != "Unknown"; phase++ {
		row, ok := rows[phase]
		if !ok {
			t.Errorf("phase %s has no row: who may write in it?", phase)
			continue
		}
		for writer, want := range row {
			t.Run(phase.String()+"/"+writer, func(t *testing.T) {
				evs, err := Decide(chatState(phase), SendChat{By: PlayerID(writer), Text: "chào"})
				if !errors.Is(err, want) {
					t.Fatalf("err = %v, want %v", err, want)
				}
				if want != nil && len(evs) != 0 {
					t.Fatalf("a refused message produced %v", evs)
				}
				if want == nil && !reflect.DeepEqual(evs, []Event{ChatSent{By: PlayerID(writer), Text: "chào"}}) {
					t.Fatalf("events = %#v", evs)
				}
			})
		}
	}
}

func TestChat_UnknownPlayerIsRefused(t *testing.T) {
	if _, err := Decide(chatState(PhaseDay), SendChat{By: "nobody", Text: "chào"}); !errors.Is(err, ErrUnknownPlayer) {
		t.Fatalf("err = %v, want ErrUnknownPlayer", err)
	}
}

func TestChat_Text(t *testing.T) {
	const family = "👨\u200d👩\u200d👧"
	tests := []struct {
		name, in, want string
	}{
		{"plain Vietnamese", "Tôi nghi An là sói", "Tôi nghi An là sói"},
		{"spaces at both ends", "  chào \u00a0", "chào"},
		{"spaces inside stay", "a  b", "a  b"},
		{"decomposed accents compose", "Bi\u0300nh", "Bình"},
		{"an accent split by an invisible character still composes", "Bi\u200b\u0300nh", "Bình"},
		{"a line break becomes a space", "Tôi nghi\nAn", "Tôi nghi An"},
		{"a Windows line break is one space", "a\r\nb", "a b"},
		{"several line breaks are one space", "a\n\n\nb", "a b"},
		{"a tab becomes a space", "a\tb", "a b"},
		{"a line break at the end goes", "chào\n", "chào"},
		{"other control characters go", "a\x00\x07\x1b[31mb", "a[31mb"},
		{"zero-width characters go", "a\u200b\u200c\u2060\ufeffb", "ab"},
		{"direction overrides go", "a\u202eb\u202c", "ab"},
		{"tag characters go", "a\U000e0041b", "ab"},
		{"a blank-looking letter goes", "a\u3164b", "ab"},
		{"invalid UTF-8 goes", "a\xffb", "ab"},
		{"emoji stay", "sói 🐺", "sói 🐺"},
		{"joined emoji stay joined", family, family},
		{"an emoji's style stays", "❤\ufe0f", "❤\ufe0f"},
		{"nothing", "", ""},
		{"only spaces", " \u00a0\u3000 ", ""},
		{"only line breaks", "\n\r\n\t", ""},
		{"only control characters", "\x00\x07", ""},
		{"only invisible characters", "\u200b\u200d\ufe0f\u3164", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evs, err := Decide(chatState(PhaseDay), SendChat{By: "alive", Text: tc.in})
			if tc.want == "" {
				if !errors.Is(err, ErrChatEmpty) {
					t.Fatalf("err = %v, want ErrChatEmpty", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := evs[0].(ChatSent).Text; got != tc.want {
				t.Fatalf("text = %q, want %q", got, tc.want)
			}
		})
	}
}

// The limit counts characters, after normalizing.
func TestChat_Length(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{"exactly the limit", strings.Repeat("a", MaxChatLen), nil},
		{"one over", strings.Repeat("a", MaxChatLen+1), ErrChatTooLong},
		{"Vietnamese letters count as one each", strings.Repeat("ữ", MaxChatLen), nil},
		{"decomposed letters are counted composed", strings.Repeat("u\u031b\u0303", MaxChatLen), nil},
		{"emoji count as one each", strings.Repeat("🐺", MaxChatLen), nil},
		{"spaces at the ends do not count", "  " + strings.Repeat("a", MaxChatLen) + "\n", nil},
		{"invisible characters do not count", strings.Repeat("a\u200b", MaxChatLen), nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			evs, err := Decide(chatState(PhaseDay), SendChat{By: "alive", Text: tc.in})
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.want == nil && utf8.RuneCountInString(evs[0].(ChatSent).Text) != MaxChatLen {
				t.Fatalf("stored %d characters", utf8.RuneCountInString(evs[0].(ChatSent).Text))
			}
		})
	}
}

func TestChat_TooLongTextNamesTheLimit(t *testing.T) {
	text, _ := i18n.ErrorText("chat_too_long", "")
	if !strings.Contains(text, strconv.Itoa(MaxChatLen)) {
		t.Fatalf("%q does not say %d", text, MaxChatLen)
	}
}

func TestChat_IsNeverReadAloudNorInTheSummary(t *testing.T) {
	var e Event = ChatSent{}
	if _, ok := e.(interface{ AnnouncementKey() string }); ok {
		t.Error("the speaker would read a player's message aloud")
	}
	if _, ok := e.(Narrated); ok {
		t.Error("a player's message would go into the summary")
	}
}
