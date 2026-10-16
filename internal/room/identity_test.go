package room

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestSecretIdentity_FailsTableStaysBounded is S7: a flood of made-up
// nicknames must not grow the PIN-failures table without limit.
func TestSecretIdentity_FailsTableStaysBounded(t *testing.T) {
	clk := NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	id := NewSecretIdentity(clk, &fixedRandom{seed: 1})
	for i := 0; i < maxPINFailEntries+500; i++ {
		_, _, _ = id.ByPIN(fmt.Sprintf("ghost-%d", i), "0000")
		clk.Advance(time.Millisecond)
	}
	if len(id.fails) > maxPINFailEntries {
		t.Fatalf("fails table has %d entries, want at most %d", len(id.fails), maxPINFailEntries)
	}
}

// TestSecretIdentity_LockingStillWorksAfterTheTableFills checks the bound
// does not break the real 5-wrong-PINs-per-minute lock for a nickname that
// keeps being tried, even once eviction is in play.
func TestSecretIdentity_LockingStillWorksAfterTheTableFills(t *testing.T) {
	clk := NewFakeClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	id := NewSecretIdentity(clk, &fixedRandom{seed: 1})
	for i := 0; i < maxPINFailEntries+500; i++ {
		_, _, _ = id.ByPIN(fmt.Sprintf("ghost-%d", i), "0000")
	}
	for i := 0; i < maxPINFailures; i++ {
		clk.Advance(time.Millisecond)
		if _, _, err := id.ByPIN("An", "0000"); !errors.Is(err, ErrBadCredentials) {
			t.Fatalf("attempt %d: err = %v, want ErrBadCredentials", i, err)
		}
	}
	if _, _, err := id.ByPIN("An", "0000"); !errors.Is(err, ErrPINLocked) {
		t.Fatalf("after %d wrong PINs: err = %v, want ErrPINLocked", maxPINFailures, err)
	}
}
