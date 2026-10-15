package room

import "time"

// dayClock is the 7-minute day clock. It runs through the discussion and the
// nomination countdown and pauses during defense and verdict (rules §2, §4).
type dayClock struct {
	remaining time.Duration
	running   bool
	since     time.Time
}

// reset starts a fresh day of d.
func (c *dayClock) reset(d time.Duration, now time.Time) {
	c.remaining, c.running, c.since = d, false, now
}

// tick takes the time that passed while running off the clock.
func (c *dayClock) tick(now time.Time) {
	if c.running {
		c.remaining -= now.Sub(c.since)
		if c.remaining < 0 {
			c.remaining = 0
		}
	}
	c.since = now
}

// setRunning must follow a tick at the same instant.
func (c *dayClock) setRunning(running bool) { c.running = running }
