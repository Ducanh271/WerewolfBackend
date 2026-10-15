// Package wsclient is a small WebSocket client for docs/protocol.md. It keeps
// what a phone would keep (the latest view, the event list, the welcome) and
// lets a caller wait for a condition. The end-to-end test and `make demo` use
// it as their fake phones.
package wsclient

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// Frame is one server message decoded as generic JSON.
type Frame map[string]any

// Type is the message type.
func (f Frame) Type() string { s, _ := f["type"].(string); return s }

// Get follows a path of object keys and returns what it finds, or nil.
func (f Frame) Get(path ...string) any {
	var cur any = map[string]any(f)
	for _, k := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[k]
	}
	return cur
}

// Str is Get as a string ("" when absent).
func (f Frame) Str(path ...string) string { s, _ := f.Get(path...).(string); return s }

// Snapshot is the phone's state at one moment. Do not modify it.
type Snapshot struct {
	Welcome Frame   // the welcome message, nil before joining
	View    Frame   // the latest `view` message (its "view" key is the screen)
	Events  []Frame // the event list the phone knows, rebuilt from `events` messages
	Errors  []Frame // every `error` message so far
	Frames  []Frame // every message, in order
	Closed  bool
}

// Client is one connected phone.
type Client struct {
	conn *websocket.Conn

	mu     sync.Mutex
	snap   Snapshot
	notify chan struct{}
	done   chan struct{}

	// OnFrame, if set before the first message arrives (use DialWith), sees
	// every message as it comes, on the reader goroutine.
	onFrame func(Frame)
}

// Dial connects to a ws:// or http:// URL of /ws.
func Dial(ctx context.Context, url string, header http.Header) (*Client, error) {
	return DialWith(ctx, url, header, nil)
}

// DialWith is Dial with a callback for every message.
func DialWith(ctx context.Context, url string, header http.Header, onFrame func(Frame)) (*Client, error) {
	conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		return nil, err
	}
	conn.SetReadLimit(1 << 20)
	c := &Client{conn: conn, notify: make(chan struct{}, 1), done: make(chan struct{}), onFrame: onFrame}
	go c.read()
	return c, nil
}

func (c *Client) read() {
	defer func() {
		c.mu.Lock()
		c.snap.Closed = true
		c.mu.Unlock()
		close(c.done)
		c.poke()
	}()
	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			return
		}
		var f Frame
		if json.Unmarshal(data, &f) != nil {
			continue
		}
		c.mu.Lock()
		c.apply(f)
		c.mu.Unlock()
		if c.onFrame != nil {
			c.onFrame(f)
		}
		c.poke()
	}
}

func (c *Client) apply(f Frame) {
	c.snap.Frames = append(append([]Frame(nil), c.snap.Frames...), f)
	switch f.Type() {
	case "welcome":
		c.snap.Welcome = f
	case "view":
		c.snap.View = f
	case "error":
		c.snap.Errors = append(append([]Frame(nil), c.snap.Errors...), f)
	case "events":
		list, _ := f["events"].([]any)
		var evs []Frame
		for _, e := range list {
			if m, ok := e.(map[string]any); ok {
				evs = append(evs, Frame(m))
			}
		}
		from := 0
		if reset, _ := f["reset"].(bool); !reset {
			from = int(num(f["from"]))
		}
		if from > len(c.snap.Events) {
			from = len(c.snap.Events)
		}
		c.snap.Events = append(append([]Frame(nil), c.snap.Events[:from]...), evs...)
	}
}

func num(v any) float64 { n, _ := v.(float64); return n }

func (c *Client) poke() {
	select {
	case c.notify <- struct{}{}:
	default:
	}
}

// Send writes one message (any JSON-encodable value).
func (c *Client) Send(ctx context.Context, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.SendRaw(ctx, b)
}

// SendRaw writes the bytes as one text message.
func (c *Client) SendRaw(ctx context.Context, b []byte) error {
	return c.conn.Write(ctx, websocket.MessageText, b)
}

// Snapshot returns the current state.
func (c *Client) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.snap
}

// Wait blocks until cond is true for the phone's state, the connection ends
// or ctx is done. It reports whether cond held.
func (c *Client) Wait(ctx context.Context, cond func(Snapshot) bool) bool {
	for {
		if s := c.Snapshot(); cond(s) {
			return true
		} else if s.Closed {
			return false
		}
		select {
		case <-c.notify:
		case <-c.done:
		case <-ctx.Done():
			return false
		}
	}
}

// Done is closed when the connection has ended.
func (c *Client) Done() <-chan struct{} { return c.done }

// Close drops the connection, as a phone losing signal would.
func (c *Client) Close() { _ = c.conn.CloseNow() }
