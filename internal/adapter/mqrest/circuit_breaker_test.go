package mqrest

import (
	"errors"
	"testing"
	"time"

	"github.com/konih/mkurator/internal/mqadmin"
)

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
