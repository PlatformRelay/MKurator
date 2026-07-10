package mqrest

import (
	"errors"
	"testing"
	"time"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

func TestCircuitBreakerConfigFromResilience(t *testing.T) {
	t.Parallel()
	cfg := circuitBreakerConfigFromResilience(ResilienceConfig{
		FailureThreshold: 7,
		OpenTimeout:      12 * time.Second,
	})
	if cfg.failureThreshold != 7 || cfg.openTimeout != 12*time.Second {
		t.Fatalf("cfg=%+v", cfg)
	}
}

func TestNewCircuitBreakerDefaults(t *testing.T) {
	t.Parallel()
	b := newCircuitBreaker(circuitBreakerConfig{failureThreshold: 0, openTimeout: 0, now: nil})
	if b.cfg.failureThreshold != defaultFailureThreshold || b.cfg.openTimeout != defaultOpenTimeout {
		t.Fatalf("cfg=%+v", b.cfg)
	}
	if b.cfg.now == nil {
		t.Fatal("expected default now func")
	}
}

func TestCircuitBreakerOpensAfterThreshold(t *testing.T) {
	t.Parallel()
	b := newCircuitBreaker(circuitBreakerConfig{failureThreshold: 3, openTimeout: time.Minute, now: time.Now})
	for i := 0; i < 3; i++ {
		if err := b.beforeRequest(); err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
		b.recordFailure()
	}
	if err := b.beforeRequest(); err == nil || !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected open circuit transient error, got %v", err)
	}
}

func TestCircuitBreakerRecordFailureBelowThreshold(t *testing.T) {
	t.Parallel()
	b := newCircuitBreaker(circuitBreakerConfig{failureThreshold: 3, openTimeout: time.Minute, now: time.Now})
	if err := b.beforeRequest(); err != nil {
		t.Fatal(err)
	}
	b.recordFailure()
	if err := b.beforeRequest(); err != nil {
		t.Fatal("circuit should stay closed below threshold")
	}
}

func TestCircuitBreakerHalfOpenBlocksConcurrentProbe(t *testing.T) {
	t.Parallel()
	start := time.Now()
	now := start
	b := newCircuitBreaker(circuitBreakerConfig{
		failureThreshold: 1,
		openTimeout:      time.Millisecond,
		now:              func() time.Time { return now },
	})
	b.recordFailure()
	now = start.Add(2 * time.Millisecond)
	if err := b.beforeRequest(); err != nil {
		t.Fatal(err)
	}
	if err := b.beforeRequest(); err == nil || !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected probe block, got %v", err)
	}
}

func TestCircuitBreakerTransitionNoOp(t *testing.T) {
	t.Parallel()
	var calls int
	b := newCircuitBreaker(circuitBreakerConfig{
		failureThreshold: 5,
		openTimeout:      time.Minute,
		now:              time.Now,
		onTransition:     func(from, to string) { calls++ },
	})
	b.transitionLocked(breakerStateClosed)
	if calls != 0 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestCircuitBreakerHalfOpenProbe(t *testing.T) {
	t.Parallel()
	start := time.Now()
	now := start
	b := newCircuitBreaker(
		circuitBreakerConfig{
			failureThreshold: 1,
			openTimeout:      10 * time.Millisecond,
			now:              func() time.Time { return now },
		},
	)
	_ = b.beforeRequest()
	b.recordFailure()
	if err := b.beforeRequest(); err == nil {
		t.Fatal("expected open")
	}
	now = start.Add(20 * time.Millisecond)
	if err := b.beforeRequest(); err != nil {
		t.Fatal(err)
	}
	b.recordSuccess()
	if err := b.beforeRequest(); err != nil {
		t.Fatal(err)
	}
}

// A probe that fails in half_open must re-open the breaker and reset openedAt,
// so the open timeout starts again from the failed probe.
func TestCircuitBreakerHalfOpenProbeFailureReopens(t *testing.T) {
	t.Parallel()
	start := time.Now()
	now := start
	var transitions []string
	b := newCircuitBreaker(circuitBreakerConfig{
		failureThreshold: 1,
		openTimeout:      10 * time.Millisecond,
		now:              func() time.Time { return now },
		onTransition:     func(from, to string) { transitions = append(transitions, from+"->"+to) },
	})
	// Trip closed -> open.
	if err := b.beforeRequest(); err != nil {
		t.Fatal(err)
	}
	b.recordFailure()
	// After the open timeout, the next request is admitted as the half_open probe.
	now = start.Add(20 * time.Millisecond)
	if err := b.beforeRequest(); err != nil {
		t.Fatalf("expected probe admitted, got %v", err)
	}
	// The probe fails: half_open -> open, openedAt reset to now.
	b.recordFailure()
	if b.state != breakerStateOpen {
		t.Fatalf("state=%q, want open", b.state)
	}
	if !b.openedAt.Equal(now) {
		t.Fatalf("openedAt=%v, want %v", b.openedAt, now)
	}
	// Circuit is open again: an immediate retry (before the timeout) is rejected.
	if err := b.beforeRequest(); err == nil || !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected open circuit, got %v", err)
	}
	if len(transitions) == 0 || transitions[len(transitions)-1] != "half_open->open" {
		t.Fatalf("transitions=%v, want last half_open->open", transitions)
	}
}

// An unknown breaker state must fail safe (treated as open) rather than admit traffic.
func TestCircuitBreakerBeforeRequestUnknownStateFailsSafe(t *testing.T) {
	t.Parallel()
	b := newCircuitBreaker(circuitBreakerConfig{failureThreshold: 1, openTimeout: time.Minute, now: time.Now})
	b.state = "bogus"
	if err := b.beforeRequest(); err == nil || !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expected fail-safe transient error, got %v", err)
	}
}
