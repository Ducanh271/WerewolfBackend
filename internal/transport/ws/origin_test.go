package ws

import (
	"net/http"
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

func TestOrigin_OnlyConfiguredOriginsMayConnect(t *testing.T) {
	rows := []struct {
		name    string
		mod     func(*Config)
		origin  string
		allowed bool
	}{
		{"listed origin", func(c *Config) { c.AllowAnyOrigin = false; c.AllowedOrigins = []string{"https://masoi.example"} }, "https://masoi.example", true},
		{"listed origin, other capitals", func(c *Config) { c.AllowAnyOrigin = false; c.AllowedOrigins = []string{"https://masoi.example"} }, "https://MASOI.example", true},
		{"unlisted origin", func(c *Config) { c.AllowAnyOrigin = false; c.AllowedOrigins = []string{"https://masoi.example"} }, "https://evil.example", false},
		{"same host, other scheme", func(c *Config) { c.AllowAnyOrigin = false; c.AllowedOrigins = []string{"https://masoi.example"} }, "http://masoi.example", false},
		{"same host, other port", func(c *Config) { c.AllowAnyOrigin = false; c.AllowedOrigins = []string{"https://masoi.example"} }, "https://masoi.example:8443", false},
		{"no origin, default config", func(c *Config) { c.AllowAnyOrigin = false; c.AllowedOrigins = []string{"https://masoi.example"} }, "", false},
		{"no origin, nothing configured", func(c *Config) { c.AllowAnyOrigin = false }, "", false},
		{"no origin, ALLOW_ANY_ORIGIN", func(c *Config) { c.AllowAnyOrigin = true }, "", true},
		{"any origin, ALLOW_ANY_ORIGIN", func(c *Config) { c.AllowAnyOrigin = true }, "https://evil.example", true},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			e := newEnv(t, row.mod)
			header := http.Header{}
			if row.origin != "" {
				header.Set("Origin", row.origin)
			}
			c, err := wsclient.Dial(ctx(t), e.wsURL(), header)
			if err == nil {
				c.Close()
			}
			if (err == nil) != row.allowed {
				t.Fatalf("connected = %v, want %v (err %v)", err == nil, row.allowed, err)
			}
		})
	}
}

func TestOrigin_ComesFromTheEnvironment(t *testing.T) {
	env := map[string]string{"ALLOWED_ORIGINS": " https://a.example, https://b.example ,", "ALLOW_ANY_ORIGIN": "1"}
	cfg := ConfigFromEnv(func(k string) string { return env[k] })
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[1] != "https://b.example" || !cfg.AllowAnyOrigin {
		t.Fatalf("config = %+v", cfg)
	}
	cfg = ConfigFromEnv(func(string) string { return "" })
	if len(cfg.AllowedOrigins) != 0 || cfg.AllowAnyOrigin {
		t.Fatalf("default config must refuse everything: %+v", cfg)
	}
}
