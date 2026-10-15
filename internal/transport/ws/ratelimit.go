package ws

import "time"

// limiter is a token bucket for one connection, plus flood detection: a
// connection that keeps hitting the limit for floodAfter is a flood.
type limiter struct {
	rate       float64 // tokens per second
	burst      float64
	floodAfter time.Duration

	tokens float64
	last   time.Time

	overSince time.Time // when the current run of refusals began
	lastOver  time.Time // the latest refusal
}

func newLimiter(rate float64, burst int, floodAfter time.Duration, now time.Time) *limiter {
	return &limiter{rate: rate, burst: float64(burst), floodAfter: floodAfter, tokens: float64(burst), last: now}
}

// allow takes a token. ok is false when the connection is over its limit;
// flood is true when it has been over the limit without a break (a refusal at
// least once a second) for floodAfter.
func (l *limiter) allow(now time.Time) (ok, flood bool) {
	l.tokens += now.Sub(l.last).Seconds() * l.rate
	if l.tokens > l.burst {
		l.tokens = l.burst
	}
	l.last = now
	if l.tokens >= 1 {
		l.tokens--
		return true, false
	}
	if l.lastOver.IsZero() || now.Sub(l.lastOver) > time.Second {
		l.overSince = now
	}
	l.lastOver = now
	return false, now.Sub(l.overSince) >= l.floodAfter
}
