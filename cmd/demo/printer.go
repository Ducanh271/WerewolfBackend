package main

import (
	"fmt"
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
	p.events = append(p.events[:min(from, len(p.events))], evs...)
	for ; p.printed < len(p.events); p.printed++ {
		p.show(p.events[p.printed])
	}
}

func (p *printer) show(ev wsclient.Frame) {
	name := func(key string) string {
		id := ev.Str("data", key)
		if n, ok := p.names[id]; ok {
			return n
		}
		return id
	}
	if say := ev.Str("say"); say != "" {
		fmt.Printf("  🔊 %s\n", say)
		return
	}
	switch ev.Str("type") {
	case "NightStarted":
		fmt.Printf("\n=== Đêm %v ===\n", ev.Get("data", "night"))
	case "NightStepCalled":
		fmt.Printf("  🌙 gọi %s (%vs)\n", ev.Str("data", "role"), ev.Get("data", "duration").(float64)/1000)
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
