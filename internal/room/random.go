package room

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Randomness is where the room gets every secret and every seed. Nothing a
// client sends ever ends up here or replaces it.
type Randomness interface {
	// Seed settles a deal or an election that votes did not decide.
	Seed() int64
	// Token is the secret code a phone keeps to rejoin.
	Token() string
	// PIN is the 4-digit code shown on the player's screen.
	PIN() string
}

// CryptoRandom draws from crypto/rand.
type CryptoRandom struct{}

// Seed implements Randomness.
func (CryptoRandom) Seed() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return int64(binary.BigEndian.Uint64(b[:]))
}

// Token implements Randomness: 128 bits, hex encoded.
func (CryptoRandom) Token() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// PIN implements Randomness.
func (CryptoRandom) PIN() string {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "0000"
	}
	return fmt.Sprintf("%04d", n.Int64())
}
