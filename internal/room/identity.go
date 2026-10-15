package room

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"
	"time"

	"masoi/internal/engine"
)

// Credentials are what a player is given on joining. The phone keeps the
// token; the PIN is shown on that player's screen only.
type Credentials struct {
	Token string
	PIN   string
}

// Errors of rejoining.
var (
	ErrBadCredentials = errors.New("bad credentials")
	ErrPINLocked      = errors.New("pin rejoin is locked")
)

// PINLockedError says how long PIN rejoin stays locked for a nickname.
type PINLockedError struct{ RetryAfter time.Duration }

func (e PINLockedError) Error() string {
	return fmt.Sprintf("%v: retry in %s", ErrPINLocked, e.RetryAfter)
}

// Is makes errors.Is(err, ErrPINLocked) work.
func (e PINLockedError) Is(target error) bool { return target == ErrPINLocked }

// Identity decides who a phone is. It sits behind an interface so a Google
// login can replace it later (rules §10). It is only called from the room's
// goroutine.
type Identity interface {
	// Issue gives a new player a token and a PIN.
	Issue(p engine.PlayerID, nickname string) Credentials
	// ByToken finds the player a secret code belongs to.
	ByToken(token string) (engine.PlayerID, error)
	// ByPIN finds the player for nickname + PIN and hands out a new token:
	// the old one stops working.
	ByPIN(nickname, pin string) (engine.PlayerID, string, error)
	// Revoke forgets a player who left.
	Revoke(p engine.PlayerID)
}

const (
	maxPINFailures = 5
	pinLockTime    = time.Minute
)

// SecretIdentity is the v1 identity: secret token and 4-digit PIN.
type SecretIdentity struct {
	clock  Clock
	rnd    Randomness
	tokens map[string]engine.PlayerID
	pins   map[engine.PlayerID]string
	nicks  map[string]engine.PlayerID
	fails  map[string]*pinFailures
}

type pinFailures struct {
	count       int
	lockedUntil time.Time
}

// NewSecretIdentity returns an identity that locks PIN rejoin on the clock.
func NewSecretIdentity(clock Clock, rnd Randomness) *SecretIdentity {
	return &SecretIdentity{
		clock:  clock,
		rnd:    rnd,
		tokens: map[string]engine.PlayerID{},
		pins:   map[engine.PlayerID]string{},
		nicks:  map[string]engine.PlayerID{},
		fails:  map[string]*pinFailures{},
	}
}

func nickKey(nickname string) string { return strings.ToLower(strings.TrimSpace(nickname)) }

// Issue implements Identity.
func (id *SecretIdentity) Issue(p engine.PlayerID, nickname string) Credentials {
	c := Credentials{Token: id.rnd.Token(), PIN: id.rnd.PIN()}
	id.tokens[c.Token] = p
	id.pins[p] = c.PIN
	id.nicks[nickKey(nickname)] = p
	return c
}

// ByToken implements Identity. A token is never locked out.
func (id *SecretIdentity) ByToken(token string) (engine.PlayerID, error) {
	p, ok := id.tokens[token]
	if !ok {
		return "", ErrBadCredentials
	}
	return p, nil
}

// ByPIN implements Identity. Five wrong PINs for a nickname lock PIN rejoin
// for that nickname for a minute, whether or not the nickname exists.
func (id *SecretIdentity) ByPIN(nickname, pin string) (engine.PlayerID, string, error) {
	key := nickKey(nickname)
	now := id.clock.Now()
	f := id.fails[key]
	if f == nil {
		f = &pinFailures{}
		id.fails[key] = f
	}
	if now.Before(f.lockedUntil) {
		return "", "", PINLockedError{RetryAfter: f.lockedUntil.Sub(now)}
	}
	p, known := id.nicks[key]
	if !known || subtle.ConstantTimeCompare([]byte(id.pins[p]), []byte(pin)) != 1 {
		f.count++
		if f.count >= maxPINFailures {
			f.count = 0
			f.lockedUntil = now.Add(pinLockTime)
		}
		return "", "", ErrBadCredentials
	}
	f.count = 0
	for token, owner := range id.tokens {
		if owner == p {
			delete(id.tokens, token)
		}
	}
	token := id.rnd.Token()
	id.tokens[token] = p
	return p, token, nil
}

// Revoke implements Identity.
func (id *SecretIdentity) Revoke(p engine.PlayerID) {
	for token, owner := range id.tokens {
		if owner == p {
			delete(id.tokens, token)
		}
	}
	delete(id.pins, p)
	for key, owner := range id.nicks {
		if owner == p {
			delete(id.nicks, key)
		}
	}
}
