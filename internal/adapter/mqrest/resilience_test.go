package mqrest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/konih/mkurator/internal/mqadmin"
)

func TestRoundTripRetriesTransientHTTPStatus(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c := &Client{
		httpClient: srv.Client(),
		retry: retryPolicy{
			maxAttempts:    4,
			initialBackoff: time.Millisecond,
			maxBackoff:     5 * time.Millisecond,
			sleep:          time.Sleep,
		},
		breaker: newCircuitBreaker(defaultCircuitBreakerConfig()),
	}
	res, err := c.roundTrip(context.Background(), func(ctx context.Context) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer closeBody(res.Body)
	if attempts.Load() != 3 {
		t.Fatalf("attempts=%d", attempts.Load())
	}
}

func TestRoundTripOpenCircuitSkipsHTTP(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := &Client{
		httpClient: srv.Client(),
		retry: retryPolicy{
			maxAttempts:    1,
			initialBackoff: time.Millisecond,
			maxBackoff:     time.Millisecond,
			sleep:          time.Sleep,
		},
		breaker: newCircuitBreaker(
			circuitBreakerConfig{failureThreshold: 1, openTimeout: time.Minute, now: time.Now},
		),
	}
	_, _ = c.roundTrip(context.Background(), func(ctx context.Context) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	})
	_, err := c.roundTrip(context.Background(), func(ctx context.Context) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	})
	if err == nil || !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("got %v", err)
	}
	if attempts.Load() != 1 {
		t.Fatalf("attempts=%d", attempts.Load())
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	t.Parallel()
	if !isRetryableHTTPStatus(http.StatusTooManyRequests) || isRetryableHTTPStatus(http.StatusBadRequest) {
		t.Fatal("unexpected retry classification")
	}
}

func TestRoundTripNetworkErrorRetries(t *testing.T) {
	t.Parallel()
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			hj, _ := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()
	c := &Client{
		httpClient: srv.Client(),
		retry: retryPolicy{
			maxAttempts:    3,
			initialBackoff: time.Millisecond,
			maxBackoff:     5 * time.Millisecond,
			sleep:          time.Sleep,
		},
		breaker: newCircuitBreaker(defaultCircuitBreakerConfig()),
	}
	res, err := c.roundTrip(context.Background(), func(ctx context.Context) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer closeBody(res.Body)
	if attempts.Load() != 2 {
		t.Fatalf("attempts=%d", attempts.Load())
	}
}
