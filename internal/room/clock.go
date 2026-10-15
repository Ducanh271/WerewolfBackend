package room

import (
	"sort"
	"sync"
	"time"
)

// Timer is a pending call that can be cancelled.
type Timer interface {
	Stop() bool
}

// Clock is the room's only source of time, so tests can run a 7-minute day
// in no time.
type Clock interface {
	Now() time.Time
	AfterFunc(d time.Duration, f func()) Timer
}

// RealClock is the wall clock.
type RealClock struct{}

// Now implements Clock.
func (RealClock) Now() time.Time { return time.Now() }

// AfterFunc implements Clock.
func (RealClock) AfterFunc(d time.Duration, f func()) Timer { return time.AfterFunc(d, f) }

// FakeClock only moves when Advance is called. Timers due within the step run
// in order, on the caller's goroutine.
type FakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
	made   int
}

type fakeTimer struct {
	clock *FakeClock
	when  time.Time
	order int
	f     func()
}

// NewFakeClock starts a fake clock at the given time.
func NewFakeClock(start time.Time) *FakeClock { return &FakeClock{now: start} }

// Now implements Clock.
func (c *FakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// AfterFunc implements Clock. A timer with d <= 0 fires on the next Advance.
func (c *FakeClock) AfterFunc(d time.Duration, f func()) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.made++
	t := &fakeTimer{clock: c, when: c.now.Add(d), order: c.made, f: f}
	c.timers = append(c.timers, t)
	return t
}

// Stop implements Timer.
func (t *fakeTimer) Stop() bool {
	c := t.clock
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, other := range c.timers {
		if other == t {
			c.timers = append(c.timers[:i], c.timers[i+1:]...)
			return true
		}
	}
	return false
}

// Pending is the number of timers that have not fired or been stopped.
func (c *FakeClock) Pending() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.timers)
}

// Advance moves time forward by d and runs every timer that comes due.
func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	target := c.now.Add(d)
	c.mu.Unlock()
	for {
		c.mu.Lock()
		sort.SliceStable(c.timers, func(i, j int) bool {
			if c.timers[i].when.Equal(c.timers[j].when) {
				return c.timers[i].order < c.timers[j].order
			}
			return c.timers[i].when.Before(c.timers[j].when)
		})
		if len(c.timers) == 0 || c.timers[0].when.After(target) {
			c.now = target
			c.mu.Unlock()
			return
		}
		t := c.timers[0]
		c.timers = c.timers[1:]
		if t.when.After(c.now) {
			c.now = t.when
		}
		c.mu.Unlock()
		t.f()
	}
}
