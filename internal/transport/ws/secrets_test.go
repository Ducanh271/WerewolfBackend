package ws

import (
	"strings"
	"testing"

	"masoi/internal/transport/ws/wsclient"
)

func TestSecrets_TheSecretCodeTravelsInMessagesNeverInTheURL(t *testing.T) {
	e := newEnv(t, nil)

	// Anything in the query string is refused, so no client can be tempted to
	// put a token or a PIN there (URLs end up in proxy and access logs).
	_, err := wsclient.Dial(ctx(t), e.wsURL()+"?token=abc&pin=1234", nil)
	if err == nil {
		t.Fatal("a socket with a query string was accepted")
	}

	code := e.createRoom()
	c := e.join(code, "An")
	w := c.Snapshot().Welcome
	if w.Str("token") == "" || w.Str("pin") == "" {
		t.Fatalf("welcome has no token or PIN: %v", w)
	}

	// The token in the first message of a new connection is what rejoins.
	back := e.dial(nil)
	e.send(back, map[string]any{"type": "rejoin", "room": code, "token": w.Str("token")})
	if !back.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Welcome != nil }) {
		t.Fatalf("rejoin by token failed: %v", back.Snapshot().Errors)
	}
}

func TestSecrets_ServerLogsNeverPrintSecretCodesOrPINs(t *testing.T) {
	e := newEnv(t, nil)
	code := e.createRoom()
	c := e.join(code, "An")
	token, pin := c.Snapshot().Welcome.Str("token"), c.Snapshot().Welcome.Str("pin")

	// Every kind of join message, good and bad, plus junk that echoes them.
	other := e.dial(nil)
	e.send(other, map[string]any{"type": "rejoin", "room": code, "nickname": "An", "pin": "0000"})
	e.waitError(other, 1)
	e.send(other, map[string]any{"type": "rejoin", "room": code, "nickname": "An", "pin": pin})
	other.Wait(ctx(t), func(s wsclient.Snapshot) bool { return s.Welcome != nil })
	newToken := other.Snapshot().Welcome.Str("token")
	e.send(other, map[string]any{"type": "rejoin", "room": code, "token": token, "seed": 5})
	e.waitError(other, 2)
	bad := e.dial(nil)
	e.send(bad, map[string]any{"type": "rejoin", "room": code, "token": token + "x", "pin": pin})
	e.waitError(bad, 1)
	e.send(bad, map[string]any{"type": "rejoin", "room": code, "token": "wrong-" + token})
	e.waitError(bad, 2)
	c.Close()
	other.Close()

	logs := e.logs.String()
	if !strings.Contains(logs, "connected") {
		t.Fatalf("expected the server to log connections, got:\n%s", logs)
	}
	for name, secret := range map[string]string{"token": token, "new token": newToken, "PIN": pin} {
		if secret == "" || strings.Contains(logs, secret) {
			t.Errorf("the %s %q is in the logs (or empty):\n%s", name, secret, logs)
		}
	}
}
