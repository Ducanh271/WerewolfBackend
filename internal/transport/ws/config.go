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
	JoinTimeout     time.Duration // time to send the first message, default 10 s

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
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.DiscardHandler)
	}
}
