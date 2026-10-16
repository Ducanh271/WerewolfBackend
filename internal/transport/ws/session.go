package ws

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"

	"masoi/internal/engine"
	"masoi/internal/room"
)

const (
	outQueue     = 64
	writeTimeout = 5 * time.Second
)

// session is one WebSocket: a reader (the handler goroutine), a writer, and
// the room.Sink the room delivers into. The connection decides who acts: the
// room.Client it holds is the only way commands go in.
type session struct {
	srv    *Server
	ws     *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
	lim    *limiter

	out        chan []byte
	writerDone chan struct{}

	mu        sync.Mutex
	ready     bool     // the first reply was queued; before that the room's messages wait
	pending   [][]byte // views delivered while joining, sent after `welcome`
	closing   *closeReason
	pingTimer room.Timer // the next scheduled ping, if any

	// Touched only by the reader goroutine.
	client       *room.Client
	rm           *room.Room
	player       engine.PlayerID
	joinFailures int // failed join/rejoin messages so far, before client is set
}

type closeReason struct {
	status websocket.StatusCode
	reason string
}

func newSession(srv *Server, c *websocket.Conn, ctx context.Context, cancel context.CancelFunc) *session {
	c.SetReadLimit(int64(srv.cfg.MaxMessageBytes * hardCapFactor))
	return &session{
		srv: srv, ws: c, ctx: ctx, cancel: cancel,
		lim: newLimiter(srv.cfg.RatePerSecond, srv.cfg.Burst, srv.cfg.FloodAfter, srv.cfg.Now()),
		out: make(chan []byte, outQueue), writerDone: make(chan struct{}),
	}
}

// Deliver implements room.Sink. It runs on the room's goroutine and never
// blocks: a phone that cannot keep up is disconnected, and rejoins to get the
// full state again.
func (s *session) Deliver(m room.Message) {
	fr, err := frames(m)
	if err != nil {
		s.srv.log.Error("encode failed", "err", err)
		s.shut(websocket.StatusInternalError, "encode failed")
		return
	}
	s.mu.Lock()
	if !s.ready {
		s.pending = append(s.pending, fr...)
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	for _, b := range fr {
		s.enqueue(b)
	}
}

// reply queues a message to this phone from the reader goroutine.
func (s *session) reply(b []byte) { s.enqueue(b) }

// Evict implements room.Sink. A rejoin elsewhere took this player over
// (R2/S7): tell this phone why, then close it, without waiting for it to
// send anything itself.
//
// This writes directly instead of going through the s.out queue, and only
// cancels s.ctx (which the reader goroutine's blocked Read also uses) once
// that write is done. Queuing first and cancelling right after, as reply+
// shut do elsewhere, races the reader: cancelling s.ctx while its Read is
// still blocked can tear the whole connection down as a side effect, before
// writeLoop gets a turn to drain the queue. A direct Write is safe to run
// concurrently with a blocked Read (unlike two Reads together), so writing
// before cancelling avoids the race.
func (s *session) Evict(reason error) {
	go func() {
		s.write(errorFrame(reason))
		s.shut(websocket.StatusPolicyViolation, "replaced by another phone")
	}()
}

func (s *session) enqueue(b []byte) {
	select {
	case s.out <- b:
	default:
		s.shut(websocket.StatusPolicyViolation, "too slow")
	}
}

// open queues the first reply (welcome), then whatever the room delivered
// while the join was running, so `welcome` always comes first.
func (s *session) open(first []byte) {
	s.enqueue(first)
	for {
		s.mu.Lock()
		pending := s.pending
		s.pending = nil
		if len(pending) == 0 {
			s.ready = true
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		for _, b := range pending {
			s.enqueue(b)
		}
	}
}

// shut asks for the connection to be closed with this status. The first
// reason wins.
func (s *session) shut(status websocket.StatusCode, reason string) {
	s.mu.Lock()
	if s.closing == nil {
		s.closing = &closeReason{status, reason}
	}
	s.mu.Unlock()
	s.cancel()
}

func (s *session) closeWith() closeReason {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing != nil {
		return *s.closing
	}
	return closeReason{websocket.StatusNormalClosure, ""}
}

// writeLoop sends queued messages until the session ends, then gives what is
// still queued (an error reply, say) one last short chance to go out.
func (s *session) writeLoop() {
	defer close(s.writerDone)
	for {
		select {
		case b := <-s.out:
			if !s.write(b) {
				s.cancel()
				return
			}
		case <-s.ctx.Done():
			for {
				select {
				case b := <-s.out:
					if !s.write(b) {
						return
					}
				default:
					return
				}
			}
		}
	}
}

// write is not tied to the session's context: a message queued just before
// the session ends (an error reply) must still go out. A dead peer is cut off
// by the timeout.
func (s *session) write(b []byte) bool {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	return s.ws.Write(ctx, websocket.MessageText, b) == nil
}

// startHeartbeat arms the first ping (R1). A silent connection — one that
// stops answering, not one that was cleanly closed — is dropped once a ping
// goes unanswered, so a dead speaker phone is replaced quickly.
func (s *session) startHeartbeat() { s.scheduleNextPing() }

func (s *session) setPingTimer(t room.Timer) {
	s.mu.Lock()
	s.pingTimer = t
	s.mu.Unlock()
}

func (s *session) stopPingTimer() {
	s.mu.Lock()
	t := s.pingTimer
	s.pingTimer = nil
	s.mu.Unlock()
	if t != nil {
		t.Stop()
	}
}

func (s *session) scheduleNextPing() {
	s.setPingTimer(s.srv.cfg.PingClock.AfterFunc(s.srv.cfg.PingInterval, s.sendPing))
}

// sendPing writes one ping and waits for its pong on a goroutine of its own,
// so the scheduling timer (PingClock.AfterFunc, fake in tests) never blocks.
// The wait is bounded by cancelling the ping's context when PongTimeout
// elapses, rather than by the ping's own ctx deadline: that keeps the whole
// thing driven by PingClock instead of the wall clock, so a fake clock can
// drive it in tests without a real wait.
func (s *session) sendPing() {
	if s.ctx.Err() != nil {
		return // already closing
	}
	pingCtx, cancelPing := context.WithCancel(s.ctx)
	timeout := s.srv.cfg.PingClock.AfterFunc(s.srv.cfg.PongTimeout, cancelPing)
	go func() {
		err := s.ws.Ping(pingCtx)
		timeout.Stop()
		cancelPing()
		if err != nil {
			if s.ctx.Err() == nil { // not already closing for another reason
				s.shut(websocket.StatusPolicyViolation, "ping timeout")
			}
			return
		}
		s.scheduleNextPing()
	}()
}

// watchRoom closes the connection when the room is deleted.
func (s *session) watchRoom(done <-chan struct{}) {
	select {
	case <-done:
		s.shut(websocket.StatusGoingAway, "room closed")
	case <-s.ctx.Done():
	}
}

// run reads messages until the connection ends. A dropped connection only
// disconnects the phone: the player keeps their place (rejoin works).
//
// The join deadline (S6) is set once here, from connect, not recreated on
// every message: a connection that keeps sending something every few
// seconds must not stay unjoined forever.
func (s *session) run() {
	defer s.finish()
	joinCtx, cancelJoin := context.WithTimeout(s.ctx, s.srv.cfg.JoinTimeout)
	defer cancelJoin()
	for {
		ctx := s.ctx
		if s.client == nil {
			ctx = joinCtx
		}
		_, data, err := s.ws.Read(ctx)
		if err != nil {
			if s.client == nil && errors.Is(err, context.DeadlineExceeded) {
				s.shut(websocket.StatusPolicyViolation, "no join message")
			}
			return
		}
		if !s.handle(data) {
			return
		}
	}
}

func (s *session) finish() {
	s.cancel()
	s.stopPingTimer()
	<-s.writerDone
	if s.client != nil {
		s.client.Disconnect()
		s.srv.log.Info("disconnected", "room", s.rm.Code(), "player", s.player)
	}
	c := s.closeWith()
	_ = s.ws.Close(c.status, c.reason)
}
