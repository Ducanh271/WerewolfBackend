package room

import (
	"time"

	"masoi/internal/engine"
)

// How much the table may write (rules §4 "Nhắn tin"). One player writes one
// message per ChatInterval on average, up to ChatBurst at once after a silence.
// A game holds ChatMaxPerGame messages: the log, and with it every broadcast and
// the whole-log message a phone gets on coming back, grows with the chat, and
// BenchmarkBroadcastLongChat keeps this number under the 5 ms line of audit R5.
const (
	ChatInterval   = 2 * time.Second
	ChatBurst      = 5
	ChatMaxPerGame = 2000
)

// Refusals of a valid message.
var (
	ErrChatTooFast = engine.NewError("chat_too_fast", "too many chat messages, slow down")
	ErrChatFull    = engine.NewError("chat_full", "this game has reached its chat limit")
)

// chatLimiter is the room's chat limit, only touched on the room goroutine.
// Credit is kept by player, not by connection, so coming back on a new socket
// does not refill it.
type chatLimiter struct {
	credits map[engine.PlayerID]*chatCredit
	sent    int // messages in the current game
}

// chatCredit is a token bucket counted in time, so no rounding can give a
// message away: writing costs ChatInterval of credit, and credit grows with the
// clock up to ChatBurst messages' worth.
type chatCredit struct {
	credit time.Duration
	last   time.Time
}

// take spends one message of p's credit, or says why p cannot write now.
func (l *chatLimiter) take(p engine.PlayerID, now time.Time) error {
	if l.sent >= ChatMaxPerGame {
		return ErrChatFull
	}
	const full = ChatBurst * ChatInterval
	c := l.credits[p]
	if c == nil {
		c = &chatCredit{credit: full}
		l.credits[p] = c
	} else if elapsed := now.Sub(c.last); elapsed > 0 {
		c.credit = min(c.credit+elapsed, full)
	}
	c.last = now
	if c.credit < ChatInterval {
		return ErrChatTooFast
	}
	c.credit -= ChatInterval
	return nil
}

// count keeps the game's total up to date with an event of the log.
func (l *chatLimiter) count(e engine.Event) {
	switch e.(type) {
	case engine.ChatSent:
		l.sent++
	case engine.NewGameStarted:
		l.sent = 0
	}
}
