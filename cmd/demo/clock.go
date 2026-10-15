package main

import (
	"time"

	"masoi/internal/room"
)

// scaledClock runs the room's time speed times faster than the wall clock, so a
// demo game does not take ten minutes. Only the demo uses it; the server always
// runs on room.RealClock.
type scaledClock struct {
	start time.Time
	speed float64
}

func newScaledClock(speed float64) scaledClock {
	return scaledClock{start: time.Now(), speed: speed}
}

// Now implements room.Clock: time since the start, multiplied.
func (c scaledClock) Now() time.Time {
	return c.start.Add(time.Duration(float64(time.Since(c.start)) * c.speed))
}

// AfterFunc implements room.Clock: a wait of d room-time is d/speed real time.
func (c scaledClock) AfterFunc(d time.Duration, f func()) room.Timer {
	return time.AfterFunc(time.Duration(float64(d)/c.speed), f)
}
