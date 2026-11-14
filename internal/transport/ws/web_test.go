package ws

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

// webBuild writes a tiny web/dist: index.html, a hashed asset, a folder.
func webBuild(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range map[string]string{
		"index.html":           "<!doctype html><title>Ma Sói</title>",
		"assets/app-1a2b3c.js": "console.log(1)",
		"icons/icon.svg":       "<svg/>",
	} {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func get(t *testing.T, rawURL string) (int, http.Header, string) {
	t.Helper()
	resp, err := http.Get(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, string(b)
}

func TestWeb_ServesTheAppAndTheQRLink(t *testing.T) {
	dir := webBuild(t)
	e := newEnv(t, func(c *Config) { c.WebDir = dir })
	tests := []struct {
		name, path string
		status     int
		body       string // a part of it
		cache      string
	}{
		{"the page", "/", 200, "<title>Ma Sói</title>", "no-cache"},
		{"index.html itself", "/index.html", 200, "<title>Ma Sói</title>", "no-cache"},
		{"the QR link opens the page", "/j/K7QP", 200, "<title>Ma Sói</title>", "no-cache"},
		{"a hashed asset is cached for good", "/assets/app-1a2b3c.js", 200, "console.log", "public, max-age=31536000, immutable"},
		{"another file is checked again", "/icons/icon.svg", 200, "<svg/>", "no-cache"},
		{"a missing file", "/assets/nope.js", 404, "", ""},
		{"a folder is not listed", "/icons/", 404, "", ""},
		{"no way out of the folder", "/../go.mod", 404, "", ""},
		{"the API still answers", "/guide", 200, "quickRules", ""},
		{"the room check still answers", "/join/ZZZZ", 404, "room_not_found", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, h, body := get(t, e.srv.URL+tt.path)
			if status != tt.status || !strings.Contains(body, tt.body) {
				t.Fatalf("GET %s = %d %q, want %d containing %q", tt.path, status, body, tt.status, tt.body)
			}
			if tt.cache != "" && h.Get("Cache-Control") != tt.cache {
				t.Errorf("Cache-Control = %q, want %q", h.Get("Cache-Control"), tt.cache)
			}
			if status == 200 && tt.cache != "" && h.Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("no nosniff header")
			}
		})
	}
}

func TestWeb_WithoutWebDir_NoPage(t *testing.T) {
	e := newEnv(t, nil)
	if status, _, _ := get(t, e.srv.URL+"/"); status != 404 {
		t.Fatalf("GET / without WEB_DIR = %d, want 404", status)
	}
}

// A WEB_DIR without a build must stop the server at start, not serve 404s.
func TestWeb_WebDirWithoutIndex_ServerDoesNotStart(t *testing.T) {
	if _, err := New(Config{Rooms: nil, WebDir: t.TempDir()}); err == nil || !strings.Contains(err.Error(), "index.html") {
		t.Fatalf("New = %v, want an error about index.html", err)
	}
}

// The page the server serves may open the socket, whatever the domain is
// called (an ngrok name changes on every start), without ALLOWED_ORIGINS.
func TestOrigin_TheServersOwnAddressMayConnect(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.AllowAnyOrigin = false })
	u, _ := url.Parse(e.srv.URL)
	for _, row := range []struct {
		origin  string
		allowed bool
	}{
		{"http://" + u.Host, true},
		{"https://" + strings.ToUpper(u.Host), true}, // TLS ends before the server (ngrok)
		{"http://evil.example", false},
		{"http://" + u.Hostname() + ":1", false},
	} {
		header := http.Header{"Origin": {row.origin}}
		c, err := wsclient.Dial(ctx(t), e.wsURL(), header)
		if err == nil {
			c.Close()
		}
		if (err == nil) != row.allowed {
			t.Errorf("origin %s: connected = %v, want %v", row.origin, err == nil, row.allowed)
		}
	}
}

func TestWeb_WebDirComesFromTheEnvironment(t *testing.T) {
	cfg := ConfigFromEnv(func(k string) string { return map[string]string{"WEB_DIR": " web/dist "}[k] })
	if cfg.WebDir != "web/dist" {
		t.Fatalf("WebDir = %q", cfg.WebDir)
	}
}

// rawGet sends the path exactly as written (net/http would clean "..").
func rawGet(t *testing.T, base, rawPath string) int {
	t.Helper()
	u, _ := url.Parse(base)
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprintf(conn, "GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", rawPath, u.Host)
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

// Nothing outside WEB_DIR, and no hidden file in it, is ever served: a
// wrongly set WEB_DIR must not hand out a .env or the .git folder.
func TestWeb_NothingOutsideTheBuildNorHidden(t *testing.T) {
	dir := webBuild(t)
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{".env": "DSN=secret", ".git/config": "[core]"} {
		p := filepath.Join(dir, name)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside, filepath.Join(dir, "assets", "link.txt")); err != nil {
		t.Skip("no symlinks here:", err)
	}
	e := newEnv(t, func(c *Config) { c.WebDir = dir })
	for _, p := range []string{"/.env", "/.git/config", "/assets/link.txt"} {
		if status, _, body := get(t, e.srv.URL+p); status != 404 || strings.Contains(body, "secret") {
			t.Errorf("GET %s = %d %q, want 404", p, status, body)
		}
	}
	for _, p := range []string{"/../go.mod", "/assets/../../go.mod", "/%2e%2e/go.mod", "/assets/%2e%2e%2f%2e%2e%2fgo.mod"} {
		if status := rawGet(t, e.srv.URL, p); status == 200 {
			t.Errorf("raw GET %s = 200, want it refused", p)
		}
	}
}

// A missing asset is not cached for a year, and the page cannot be framed by
// another site nor leak the room code in a Referer.
func TestWeb_Headers(t *testing.T) {
	e := newEnv(t, func(c *Config) { c.WebDir = webBuild(t) })
	status, h, _ := get(t, e.srv.URL+"/assets/missing-9z9z.js")
	if status != 404 || strings.Contains(h.Get("Cache-Control"), "immutable") {
		t.Errorf("missing asset: %d, Cache-Control %q; want 404 not cached for good", status, h.Get("Cache-Control"))
	}
	_, h, _ = get(t, e.srv.URL+"/j/K7QP")
	for k, want := range map[string]string{
		"X-Frame-Options": "DENY", "Content-Security-Policy": "frame-ancestors 'none'",
		"Referrer-Policy": "no-referrer", "X-Content-Type-Options": "nosniff",
	} {
		if h.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, h.Get(k), want)
		}
	}
}
