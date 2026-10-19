package room

import (
	"errors"
	"testing"
	"time"

	"masoi/internal/engine"
)

func TestRejoin_SamePhoneWithSecretCode_KeepsRoleAndGetsFullState(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	role := g.snap().Players["p3"].Role
	g.cl["p3"].Disconnect()

	sink := &recSink{}
	cl, err := g.r.RejoinToken(g.cred["p3"].Token, sink)
	if err != nil || cl.Player() != "p3" {
		t.Fatalf("RejoinToken = %v, %v", cl, err)
	}
	m := sink.last()
	if m.View.View.Me == nil || m.View.View.Me.Role != role {
		t.Fatalf("rejoined player sees %+v, want role %s", m.View.View.Me, role)
	}
	if m.Events == nil || !m.Events.Reset {
		t.Fatal("rejoin must send the full list of events the player may know")
	}
}

func TestRejoin_OtherPhoneWithNicknameAndPIN_GetsNewSecretCode(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	old := g.cred["p3"]

	cl, fresh, err := g.r.RejoinPIN(g.nick["p3"], old.PIN, &recSink{})
	if err != nil || cl.Player() != "p3" {
		t.Fatalf("RejoinPIN = %v, %v", cl, err)
	}
	if fresh.Token == "" || fresh.Token == old.Token {
		t.Fatalf("new token = %q, want a different one from %q", fresh.Token, old.Token)
	}
	if _, err := g.r.RejoinToken(old.Token, &recSink{}); !errors.Is(err, ErrBadCredentials) {
		t.Fatalf("the old code still works: %v", err)
	}
	if err := g.cl["p3"].Submit(as(engine.SkipDay{Voter: "p3"})); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("the old phone can still act: %v", err)
	}
	if _, err := g.r.RejoinToken(fresh.Token, &recSink{}); err != nil {
		t.Fatalf("the new code does not work: %v", err)
	}
}

// Rules §2: the nickname is matched in its normal form, so typing the accents
// or the spaces differently on the new phone still finds the player.
func TestRejoin_OtherPhoneWithTheSameNameTypedDifferently_IsFound(t *testing.T) {
	tests := []struct{ joined, typed string }{
		{"B\u00ecnh", "Bi\u0300nh"},
		{"An Bình", "  an   bình "},
	}
	for _, tt := range tests {
		t.Run(tt.typed, func(t *testing.T) {
			g := newGame(t, 4)
			id := g.joinAs(tt.joined, "")
			cl, _, err := g.r.RejoinPIN(tt.typed, g.cred[id].PIN, &recSink{})
			if err != nil || cl.Player() != id {
				t.Fatalf("RejoinPIN(%q) = %v, %v, want %s", tt.typed, cl, err, id)
			}
		})
	}
}

func TestRejoin_FiveWrongPINsLockPINRejoinForOneMinute_ButNotTheSecretCode(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	creds, nick := g.cred["p2"], g.nick["p2"]

	for i := 0; i < 5; i++ {
		if _, _, err := g.r.RejoinPIN(nick, "0000", &recSink{}); !errors.Is(err, ErrBadCredentials) {
			t.Fatalf("wrong PIN #%d: %v", i+1, err)
		}
	}
	_, _, err := g.r.RejoinPIN(nick, creds.PIN, &recSink{}) // even the right PIN
	var locked PINLockedError
	if !errors.Is(err, ErrPINLocked) || !errors.As(err, &locked) || locked.RetryAfter != time.Minute {
		t.Fatalf("right PIN while locked: %v, want a 1 minute lock", err)
	}
	if _, err := g.r.RejoinToken(creds.Token, &recSink{}); err != nil {
		t.Fatalf("the secret code must not be locked: %v", err)
	}
	if _, _, err := g.r.RejoinPIN(g.nick["p4"], g.cred["p4"].PIN, &recSink{}); err != nil {
		t.Fatalf("another nickname must not be locked: %v", err)
	}

	g.clk.Advance(time.Minute - time.Second)
	if _, _, err := g.r.RejoinPIN(nick, creds.PIN, &recSink{}); !errors.Is(err, ErrPINLocked) {
		t.Fatalf("one second early: %v, want still locked", err)
	}
	g.clk.Advance(time.Second)
	if _, _, err := g.r.RejoinPIN(nick, creds.PIN, &recSink{}); err != nil {
		t.Fatalf("after a minute: %v", err)
	}
}

func TestRejoin_NicknameAloneIsRejected(t *testing.T) {
	g := newGame(t, 5)
	g.deal()
	nick := g.nick["p2"]
	if _, _, err := g.r.RejoinPIN(nick, "", &recSink{}); !errors.Is(err, ErrBadCredentials) {
		t.Fatalf("nickname without PIN: %v", err)
	}
	if _, err := g.r.RejoinToken(nick, &recSink{}); !errors.Is(err, ErrBadCredentials) {
		t.Fatalf("nickname used as a secret code: %v", err)
	}
	if _, err := g.r.RejoinToken("", &recSink{}); !errors.Is(err, ErrBadCredentials) {
		t.Fatalf("empty code: %v", err)
	}
}
