package ws

import (
	"sort"
	"sync"
	"time"
)

// connGuard decides whether a new WebSocket connection may open. It enforces
// three limits (S6, S5's "per-IP" item): a global concurrent cap, a per-IP
// concurrent cap (kept high: a table of players usually shares one public
// IP), and the real control — how fast NEW connections may open from one
// IP. The per-IP rate limiter map is bounded the same way identity.go's PIN
// fails table is: least-recently-used entries are dropped once it is full.
type connGuard struct {
	mu    sync.Mutex
	total int
	perIP map[string]int
	rate  map[string]*ipRate

	maxTotal int
	maxPerIP int
	newRate  float64 // tokens per second
	newBurst int
	now      func() time.Time
}

type ipRate struct {
	lim      *limiter
	lastUsed time.Time
}

// maxTrackedIPs bounds the rate map so a flood from many spoofed-looking
// source ports (same IP, which is all that matters here) or many distinct
// IPs cannot grow it without limit.
const maxTrackedIPs = 20000

func newConnGuard(maxTotal, maxPerIP int, newConnPerMinute float64, newConnBurst int, now func() time.Time) *connGuard {
	return &connGuard{
		perIP: map[string]int{}, rate: map[string]*ipRate{},
		maxTotal: maxTotal, maxPerIP: maxPerIP,
		newRate: newConnPerMinute / 60, newBurst: newConnBurst, now: now,
	}
}

// admit decides whether a new connection from ip may open, and counts it in
// if so (release must be called exactly once when such a connection ends).
func (g *connGuard) admit(ip string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	r := g.rate[ip]
	if r == nil {
		g.evictOldRates(now)
		r = &ipRate{lim: newLimiter(g.newRate, g.newBurst, time.Hour, now)}
		g.rate[ip] = r
	}
	r.lastUsed = now
	if ok, _ := r.lim.allow(now); !ok {
		return false
	}
	if g.total >= g.maxTotal || g.perIP[ip] >= g.maxPerIP {
		return false
	}
	g.total++
	g.perIP[ip]++
	return true
}

// release frees one connection admitted for ip.
func (g *connGuard) release(ip string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.total--
	if g.perIP[ip] <= 1 {
		delete(g.perIP, ip)
	} else {
		g.perIP[ip]--
	}
}

// evictOldRates drops the least recently used IPs once the table is full, so
// a flood of distinct source IPs cannot grow it without limit. now is unused
// but kept for symmetry with identity.go's evictOldFails.
func (g *connGuard) evictOldRates(time.Time) {
	if len(g.rate) < maxTrackedIPs {
		return
	}
	type entry struct {
		ip   string
		used time.Time
	}
	all := make([]entry, 0, len(g.rate))
	for ip, r := range g.rate {
		all = append(all, entry{ip, r.lastUsed})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].used.Before(all[j].used) })
	drop := len(all) - maxTrackedIPs + 1
	for _, e := range all[:drop] {
		delete(g.rate, e.ip)
	}
}
