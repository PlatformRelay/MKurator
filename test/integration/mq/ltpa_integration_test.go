//go:build integration

package mq

import (
	"testing"

	"github.com/platformrelay/mkurator/internal/adapter/mqrest"
)

// newLTPAIntegrationClient builds an LTPA-mode client against the live mqweb using
// the same endpoint/TLS as the Basic integration client, but selecting LTPA login
// (POST /login -> cookie) instead of HTTP Basic (AUTH-13, ADR-0027).
func newLTPAIntegrationClient() (*mqrest.Client, error) {
	cfg, err := integrationConfig()
	if err != nil {
		return nil, err
	}
	// Switch to LTPA mode: reuse the credentials as login creds, clear Basic.
	cfg.LTPA = &mqrest.LTPAConfig{Username: cfg.Username, Password: cfg.Password}
	cfg.Username = ""
	cfg.Password = ""
	return mqrest.NewClient(cfg)
}

// TestIntegration_LTPA_LoginAndMQSCRoundTrip proves the real LTPA flow against a
// live mqweb: the client logs in once (POST /login, JSON creds, 204 + LtpaToken2
// cookie per AUTH-10), then a Ping and an MQSC round-trip both succeed carrying
// the cookie (no Authorization header). Re-login is 401-driven (MQWB0112E) and
// handled in-client; timer expiry was NOT observed in AUTH-10, so this test does
// NOT attempt to force expiry.
//
// This test is compile-gated behind the `integration` build tag and skips unless
// KURATOR_INTEGRATION_MQ=1 with a Docker mqweb reachable (task mq:integration:up).
func TestIntegration_LTPA_LoginAndMQSCRoundTrip(t *testing.T) {
	requireIntegration(t)
	ctx := testContext(t)

	c, err := newLTPAIntegrationClient()
	if err != nil {
		t.Fatalf("build LTPA client: %v", err)
	}

	// First call triggers the lazy login; success proves login + cookie caching.
	if err := c.Ping(ctx); err != nil {
		t.Fatalf("LTPA Ping (login + cookie): %v", err)
	}

	// A state-changing MQSC round-trip proves the cached cookie + CSRF header path.
	if err := c.RunMQSC(ctx, "DISPLAY QMGR"); err != nil {
		t.Fatalf("LTPA RunMQSC (cookie-authenticated): %v", err)
	}

	// A second MQSC reuses the cached cookie without a fresh login (login-once).
	if err := c.RunMQSC(ctx, "DISPLAY QMGR"); err != nil {
		t.Fatalf("LTPA second RunMQSC (cookie reuse): %v", err)
	}
}
