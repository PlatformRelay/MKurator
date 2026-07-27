package mqrest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

const (
	defaultMaxAttempts      = 4
	defaultInitialBackoff   = 200 * time.Millisecond
	defaultMaxBackoff       = 5 * time.Second
	defaultFailureThreshold = 5
	defaultOpenTimeout      = 30 * time.Second
)

// ResilienceConfig tunes mqweb retry and per-connection circuit breaking.
type ResilienceConfig struct {
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	FailureThreshold int
	OpenTimeout      time.Duration
}

type retryPolicy struct {
	maxAttempts    int
	initialBackoff time.Duration
	maxBackoff     time.Duration
	sleep          func(time.Duration)
}

func defaultRetryPolicy() retryPolicy {
	return retryPolicyFromResilience(ResilienceConfig{})
}

func retryPolicyFromResilience(cfg ResilienceConfig) retryPolicy {
	p := retryPolicy{
		maxAttempts:    defaultMaxAttempts,
		initialBackoff: defaultInitialBackoff,
		maxBackoff:     defaultMaxBackoff,
		sleep:          time.Sleep,
	}
	if cfg.MaxAttempts > 0 {
		p.maxAttempts = cfg.MaxAttempts
	}
	if cfg.InitialBackoff > 0 {
		p.initialBackoff = cfg.InitialBackoff
	}
	if cfg.MaxBackoff > 0 {
		p.maxBackoff = cfg.MaxBackoff
	}
	return p
}

func (p retryPolicy) backoff(attempt int) time.Duration {
	if attempt <= 0 {
		return p.initialBackoff
	}
	d := p.initialBackoff
	for i := 1; i < attempt; i++ {
		d *= 2
		if d >= p.maxBackoff {
			return p.maxBackoff
		}
	}
	return d
}

func isRetryableHTTPStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}

type requestBuilder func(context.Context) (*http.Request, error)

// roundTripWithReauth wraps roundTrip with at most one in-client re-authentication
// and retry per request (AUTH-13, ADR-0027). For Basic (c.reauth == nil) it is a
// straight pass-through, so the Basic path is byte-for-byte unchanged.
//
// For LTPA: it captures the session token BEFORE the request, and on a 401 whose
// body carries the re-login classifier (MQWB0112E) it re-logs-in ONCE
// (single-flighted across concurrent requests by the authenticator's generation
// double-check) and re-runs the builder — which re-invokes authenticate and
// attaches the fresh cookie. A second consecutive 401 (or any non-MQWB0112E 401,
// e.g. MQWB0104E) is returned as-is for the caller's normal 401 ->
// TerminalError{Unauthorized} mapping. There is no retry loop and no cache
// eviction (ADR-0023).
func (c *Client) roundTripWithReauth(ctx context.Context, build requestBuilder) (*http.Response, error) {
	if c.reauth == nil {
		return c.roundTrip(ctx, build)
	}

	// Establish the session (lazy initial login) and snapshot the generation of
	// the cookie this request will actually use, so a concurrent re-login is
	// detected as a no-op (single-flight). A login failure here surfaces as the
	// terminal error directly (AC3).
	seenToken, err := c.reauth.ensureLoggedIn(ctx)
	if err != nil {
		return nil, err
	}

	res, err := c.roundTrip(ctx, build)
	if err != nil || res.StatusCode != http.StatusUnauthorized {
		return res, err
	}

	// Peek the 401 body to classify, then restore it so callers read it unchanged.
	body, readErr := io.ReadAll(res.Body)
	closeBody(res.Body)
	if readErr != nil {
		res.Body = io.NopCloser(bytes.NewReader(nil))
		return res, nil
	}
	res.Body = io.NopCloser(bytes.NewReader(body))

	if !c.reauth.shouldReauthenticate(res.StatusCode, body) {
		return res, nil
	}

	// Token failed verification: re-login once (single-flighted), then retry once.
	if reErr := c.reauth.reauthenticate(ctx, seenToken); reErr != nil {
		drainAndClose(res.Body)
		return nil, reErr
	}
	drainAndClose(res.Body)
	return c.roundTrip(ctx, build)
}

func (c *Client) roundTrip(ctx context.Context, build requestBuilder) (*http.Response, error) {
	if err := c.breaker.beforeRequest(); err != nil {
		return nil, err
	}

	var lastNetErr error
	for attempt := 0; attempt < c.retry.maxAttempts; attempt++ {
		if attempt > 0 {
			if err := sleepWithContext(ctx, c.retry.sleep, c.retry.backoff(attempt)); err != nil {
				return nil, err
			}
		}

		req, err := build(ctx)
		if err != nil {
			return nil, fmt.Errorf("build mqweb request: %w", err)
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// ClientCert (mTLS): a server-originated TLS handshake alert means mqweb
			// rejected the presented client certificate. That is a terminal Unauthorized
			// (untrusted/expired/no DN mapping) — never retry it, and never let it hot-loop
			// as transient (AUTH-16 AC3). A purely local net error (reset/refused/timeout)
			// does not carry the remote-alert marker and stays transient below.
			if c.clientCertMode && isClientCertRejection(err) {
				c.breaker.recordFailure()
				return nil, clientCertRejectionError()
			}
			lastNetErr = err
			continue
		}

		if isRetryableHTTPStatus(res.StatusCode) {
			drainAndClose(res.Body)
			if attempt+1 < c.retry.maxAttempts {
				continue
			}
			c.breaker.recordFailure()
			return nil, &mqadmin.TransientError{
				Message: fmt.Sprintf("mqweb returned HTTP %d", res.StatusCode),
			}
		}

		c.breaker.recordSuccess()
		return res, nil
	}

	c.breaker.recordFailure()
	if lastNetErr != nil {
		return nil, &mqadmin.TransientError{Message: "mqweb request failed", Cause: lastNetErr}
	}
	return nil, &mqadmin.TransientError{Message: "mqweb request failed after retries"}
}

func sleepWithContext(ctx context.Context, sleepFn func(time.Duration), d time.Duration) error {
	if sleepFn == nil {
		sleepFn = time.Sleep
	}
	done := make(chan struct{})
	go func() {
		sleepFn(d)
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, body)
	closeBody(body)
}
