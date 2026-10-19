package room

import (
	"crypto/subtle"
	"fmt"
	"sort"
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
	ErrBadCredentials = engine.NewError("bad_credentials", "bad credentials")
	ErrPINLocked      = engine.NewError("pin_locked", "pin rejoin is locked")
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
	// MaxPINFailures and PINLockTime are exported so the in-app guide (T25)
	// can quote the real numbers instead of typing them into guide text.
	MaxPINFailures = 5
	PINLockTime    = time.Minute
	// maxPINFailEntries bounds the fails table (S7): a flood of bogus
	// nicknames must not grow it without limit. Once full, the entries
	// least recently tried are dropped to make room.
	maxPINFailEntries = 5000
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
	lastTry     time.Time
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

// nickKey matches a nickname the way the engine compares them (rules §2):
// normalized, any case.
func nickKey(nickname string) string { return strings.ToLower(engine.NormalizeNickname(nickname)) }

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
		id.evictOldFails(now)
		f = &pinFailures{}
		id.fails[key] = f
	}
	f.lastTry = now
	if now.Before(f.lockedUntil) {
		return "", "", PINLockedError{RetryAfter: f.lockedUntil.Sub(now)}
	}
	p, known := id.nicks[key]
	if !known || subtle.ConstantTimeCompare([]byte(id.pins[p]), []byte(pin)) != 1 {
		f.count++
		if f.count >= MaxPINFailures {
			f.count = 0
			f.lockedUntil = now.Add(PINLockTime)
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

// evictOldFails drops the least recently tried entries once the table is
// full, so a flood of made-up nicknames cannot grow it without limit.
func (id *SecretIdentity) evictOldFails(now time.Time) {
	if len(id.fails) < maxPINFailEntries {
		return
	}
	type entry struct {
		key  string
		last time.Time
	}
	all := make([]entry, 0, len(id.fails))
	for k, f := range id.fails {
		all = append(all, entry{k, f.lastTry})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].last.Before(all[j].last) })
	drop := len(all) - maxPINFailEntries + 1
	for _, e := range all[:drop] {
		delete(id.fails, e.key)
	}
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
