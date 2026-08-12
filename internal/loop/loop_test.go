package loop

import (
	"math/rand"
	"testing"
	"time"
)

func TestNextProbeTimeUsesCurrentOrNextMinute(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	next := nextProbeTime(now, rng)
	if next.Second() < 5 || next.Second() > 50 || !next.After(now) {
		t.Fatalf("next = %v", next)
	}

	late := time.Date(2026, 8, 13, 12, 0, 55, 0, time.UTC)
	next = nextProbeTime(late, rng)
	if next.Minute() != 1 || next.Second() < 5 || next.Second() > 50 {
		t.Fatalf("next after late time = %v", next)
	}
}

func TestBackoffCapsAtThirtySeconds(t *testing.T) {
	if got := retryDelay(10); got != 30*time.Second {
		t.Fatalf("retryDelay(10) = %v", got)
	}
	if got := retryDelay(0); got != time.Second {
		t.Fatalf("retryDelay(0) = %v", got)
	}
}
