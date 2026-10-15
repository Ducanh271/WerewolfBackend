package main

import (
	"fmt"
	"reflect"
	"sync"

	"masoi/internal/transport/ws/wsclient"
)

// printer shows what the speaker phone (the host) gets: the voice lines it
// would read aloud, and the public cues around them.
type printer struct {
	mu      sync.Mutex
	names   map[string]string
	printed int
	events  []wsclient.Frame
}

func newPrinter() *printer { return &printer{names: map[string]string{}} }

// host is called for every message the host phone receives.
func (p *printer) host(f wsclient.Frame) {
	if f.Type() != "events" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	list, _ := f["events"].([]any)
	var evs []wsclient.Frame
	for _, e := range list {
		if m, ok := e.(map[string]any); ok {
			evs = append(evs, wsclient.Frame(m))
		}
	}
	from := 0
	if reset, _ := f["reset"].(bool); !reset {
		from = int(f["from"].(float64))
	}
	reset, _ := f["reset"].(bool)
	old := append([]wsclient.Frame(nil), p.events[:min(p.printed, len(p.events))]...) // a copy: the append below reuses the array
	p.events = append(p.events[:min(from, len(p.events))], evs...)
	if reset {
		// A phone that sees more (the host died, or the game ended) gets the
		// whole history again, with the hidden events put back in place. What
		// was printed is a subsequence of the new list: carry on after it.
		p.printed = afterSubsequence(old, p.events)
	}
	for ; p.printed < len(p.events); p.printed++ {
		p.show(p.events[p.printed])
	}
}

func (p *printer) nameOf(id string) string {
	if n, ok := p.names[id]; ok {
		return n
	}
	return id
}

func (p *printer) show(ev wsclient.Frame) {
	name := func(key string) string { return p.nameOf(ev.Str("data", key)) }
	if say := ev.Str("say"); say != "" {
		fmt.Printf("  🔊 %s\n", say)
		return
	}
	switch ev.Str("type") {
	case "NightStarted":
		fmt.Printf("\n=== Đêm %v ===\n", ev.Get("data", "night"))
	case "NightStepCalled":
		fmt.Printf("  🌙 gọi %s (%vs)\n", ev.Str("data", "role"), ev.Get("data", "duration").(float64)/1000)
	case "PhoneChecked":
		if ev.Get("data", "cursed") == true {
			fmt.Printf("  📱 Bạn đã bị nguyền\n")
		} else {
			fmt.Printf("  📱 Bạn không bị nguyền\n")
		}
		if ev.Get("data", "love", "inLove") == true {
			fmt.Printf("  📱 Bạn yêu %s, vai %s (liên kết %s)\n", p.nameOf(ev.Str("data", "love", "partner")), ev.Str("data", "love", "partnerRole"), ev.Str("data", "love", "colour"))
		} else {
			fmt.Printf("  📱 Bạn không ở trong cặp đôi\n")
		}
	case "DayStarted":
		fmt.Printf("\n=== Ngày ===\n")
	case "ChiefElectionStarted":
		fmt.Printf("=== Bầu Trưởng làng ===\n")
	case "TrialStarted":
		fmt.Printf("  ⚖  %s ra bào chữa\n", name("defendant"))
	case "VerdictStarted":
		fmt.Printf("  ⚖  bỏ phiếu giết hay tha\n")
	}
}

// afterSubsequence returns the index in list just past the last event of
// printed, matched in order.
func afterSubsequence(printed, list []wsclient.Frame) int {
	if len(printed) == 0 {
		return 0
	}
	j := 0
	for i, e := range list {
		if j < len(printed) && reflect.DeepEqual(map[string]any(e), map[string]any(printed[j])) {
			j++
			if j == len(printed) {
				return i + 1
			}
		}
	}
	return len(list)
}
