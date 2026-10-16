package ws

import (
	"log/slog"
	"strings"
	"time"

	"masoi/internal/room"
)

// Defaults of Config.
const (
	DefaultMaxMessageBytes = 4 << 10
	DefaultRatePerSecond   = 10
	DefaultBurst           = 20
	DefaultFloodAfter      = 3 * time.Second
	DefaultJoinTimeout     = 10 * time.Second

	// DefaultPingInterval and DefaultPongTimeout: a silent connection (R1,
	// a phone that lost signal) is dropped within roughly their sum, so the
	// speaker role moves on quickly.
	DefaultPingInterval = 12 * time.Second
	DefaultPongTimeout  = 10 * time.Second

	// DefaultMaxJoinAttempts closes a connection that keeps failing to
	// join/rejoin (S6: typos and timeouts; S7: guessing a PIN across
	// nicknames on one connection).
	DefaultMaxJoinAttempts = 8

	// DefaultMaxConnections is the global cap on open WebSockets.
	DefaultMaxConnections = 2000
	// DefaultMaxConnectionsPerIP is deliberately high: a table of players
	// usually shares one public IP (café Wi-Fi, carrier NAT), so this is a
	// backstop, not the main defense (see DefaultNewConnPerMinute).
	DefaultMaxConnectionsPerIP = 100
	// DefaultNewConnPerMinute and DefaultNewConnBurst are the real control
	// on connection abuse from one IP: how fast NEW connections may open,
	// not how many may be open at once.
	DefaultNewConnPerMinute = 30
	DefaultNewConnBurst     = 30

	// hardCapFactor times MaxMessageBytes is the biggest frame the socket
	// reads at all. Between the two limits the server answers with an error;
	// above it the connection is closed.
	hardCapFactor = 16
)

// Config wires the transport. Only Rooms is required.
type Config struct {
	Rooms *room.Manager

	// AllowedOrigins are the origins (scheme://host[:port]) browsers may open
	// a socket from. A request without an Origin header is refused unless
	// AllowAnyOrigin is set (demo and tests).
	AllowedOrigins []string
	AllowAnyOrigin bool

	MaxMessageBytes int           // per message, default 4 KB
	RatePerSecond   float64       // messages per second per connection, default 10
	Burst           int           // how many may come at once, default 20
	FloodAfter      time.Duration // a connection over the limit this long is closed, default 3 s
	JoinTimeout     time.Duration // time to send the first message, counted once from connect, default 10 s
	MaxJoinAttempts int           // failed join/rejoin messages before closing a connection, default 8

	// PingInterval and PongTimeout drive the server's WebSocket heartbeat
	// (R1): every connection is pinged on PingInterval; no pong within
	// PongTimeout closes it.
	PingInterval time.Duration
	PongTimeout  time.Duration
	// PingClock schedules the heartbeat; default room.RealClock{}. Tests
	// inject a room.FakeClock to drive it without real waits.
	PingClock room.Clock

	// MaxConnections is the global cap on open WebSockets, default 2000.
	MaxConnections int
	// MaxConnectionsPerIP is a backstop, not the main defense: a table of
	// players usually shares one IP. Default 100.
	MaxConnectionsPerIP int
	// NewConnPerMinute and NewConnBurst limit how fast new connections may
	// open from one IP: the real control on connection-flood abuse.
	// Defaults 30 and 30.
	NewConnPerMinute float64
	NewConnBurst     int

	Now    func() time.Time // rate limit clock, default time.Now
	Logger *slog.Logger     // default: discard
}

// ConfigFromEnv reads ALLOWED_ORIGINS (comma separated) and ALLOW_ANY_ORIGIN=1.
func ConfigFromEnv(getenv func(string) string) Config {
	var cfg Config
	for _, o := range strings.Split(getenv("ALLOWED_ORIGINS"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
		}
	}
	cfg.AllowAnyOrigin = getenv("ALLOW_ANY_ORIGIN") == "1"
	return cfg
}

func (c *Config) setDefaults() {
	if c.MaxMessageBytes <= 0 {
		c.MaxMessageBytes = DefaultMaxMessageBytes
	}
	if c.RatePerSecond <= 0 {
		c.RatePerSecond = DefaultRatePerSecond
	}
	if c.Burst <= 0 {
		c.Burst = DefaultBurst
	}
	if c.FloodAfter <= 0 {
		c.FloodAfter = DefaultFloodAfter
	}
	if c.JoinTimeout <= 0 {
		c.JoinTimeout = DefaultJoinTimeout
	}
	if c.MaxJoinAttempts <= 0 {
		c.MaxJoinAttempts = DefaultMaxJoinAttempts
	}
	if c.PingInterval <= 0 {
		c.PingInterval = DefaultPingInterval
	}
	if c.PongTimeout <= 0 {
		c.PongTimeout = DefaultPongTimeout
	}
	if c.PingClock == nil {
		c.PingClock = room.RealClock{}
	}
	if c.MaxConnections <= 0 {
		c.MaxConnections = DefaultMaxConnections
	}
	if c.MaxConnectionsPerIP <= 0 {
		c.MaxConnectionsPerIP = DefaultMaxConnectionsPerIP
	}
	if c.NewConnPerMinute <= 0 {
		c.NewConnPerMinute = DefaultNewConnPerMinute
	}
	if c.NewConnBurst <= 0 {
		c.NewConnBurst = DefaultNewConnBurst
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.DiscardHandler)
	}
}
