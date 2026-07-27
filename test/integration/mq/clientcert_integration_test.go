//go:build integration

package mq

import (
	"os"
	"testing"

	"github.com/platformrelay/mkurator/internal/adapter/mqrest"
)

// newClientCertIntegrationClient builds a ClientCert (mTLS) client against a live mqweb
// configured for client-certificate authentication (AUTH-16, ADR-0027). It reads the
// keypair PEMs from KURATOR_INTEGRATION_MQ_CLIENT_CERT / _CLIENT_KEY.
//
// HARNESS LIMITATION (recorded honestly): the pinned mqweb image
// icr.io/ibm-messaging/mq:9.4.5.1-r1 requires SERVER-SIDE configuration to enable
// client-certificate authentication AND a certificate DN -> mqweb user registry mapping.
// That mqweb-side configuration is a DEPLOYMENT PREREQUISITE MKurator does NOT configure
// (ADR-0002 stance; ADR-0027 §mTLS). No headless recipe to configure client-cert mode on
// the pinned image was available in this lane, so this test is written to COMPILE and RUN
// only when such an environment is supplied. The runnable, self-contained proof of the mTLS
// handshake (client cert presented + accepted, no Authorization header, rejected-cert ->
// terminal Unauthorized, bad-keypair -> config error) is the httptest-based unit suite in
// internal/adapter/mqrest/clientcert_test.go, which stands up a real TLS server with
// tls.RequireAndVerifyClientCert and an in-test CA/keypair.
func newClientCertIntegrationClient(t *testing.T) (*mqrest.Client, error) {
	t.Helper()
	certPath := os.Getenv("KURATOR_INTEGRATION_MQ_CLIENT_CERT")
	keyPath := os.Getenv("KURATOR_INTEGRATION_MQ_CLIENT_KEY")
	if certPath == "" || keyPath == "" {
		t.Skip("client-cert integration needs KURATOR_INTEGRATION_MQ_CLIENT_CERT/_CLIENT_KEY " +
			"pointing at a keypair trusted by a client-cert-enabled mqweb (see harness-limitation note)")
	}
	certPEM, err := os.ReadFile(certPath) //nolint:gosec // operator-supplied integration path
	if err != nil {
		return nil, err
	}
	keyPEM, err := os.ReadFile(keyPath) //nolint:gosec // operator-supplied integration path
	if err != nil {
		return nil, err
	}

	cfg, err := integrationConfig()
	if err != nil {
		return nil, err
	}
	// ClientCert mode: no Basic credentials on the identity; auth is at the transport.
	cfg.Username = ""
	cfg.Password = ""
	// integrationConfig injects an HTTPClient; drop it so NewClientCertClient wires the
	// keypair onto a fresh transport (the injected client would bypass the keypair).
	cfg.HTTPClient = nil
	return mqrest.NewClientCertClient(cfg, certPEM, keyPEM)
}

// TestIntegration_ClientCert_MQSCRoundTrip proves the real mTLS flow against a live,
// client-cert-configured mqweb: the client presents the keypair at the transport, Ping and
// an MQSC round-trip succeed, and NO Authorization header is sent. It skips unless a
// client-cert-enabled mqweb + keypair are supplied (harness limitation above); the runnable
// substitute is the httptest mTLS suite.
func TestIntegration_ClientCert_MQSCRoundTrip(t *testing.T) {
	requireIntegration(t)
	ctx := testContext(t)

	c, err := newClientCertIntegrationClient(t)
	if err != nil {
		t.Fatalf("build ClientCert client: %v", err)
	}

	if err := c.Ping(ctx); err != nil {
		t.Fatalf("ClientCert Ping (mTLS handshake): %v", err)
	}
	if err := c.RunMQSC(ctx, "DISPLAY QMGR"); err != nil {
		t.Fatalf("ClientCert RunMQSC (transport-authenticated): %v", err)
	}
}
