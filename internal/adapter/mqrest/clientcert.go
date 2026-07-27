package mqrest

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// clientCertReason is the QMC status Reason for a build-time client-certificate
// configuration failure (bad/mismatched keypair). It is a CONFIGURATION-class error:
// non-transient, so the reconciler sets Ready=False without a hot-loop requeue (AC2).
const clientCertReason = "InvalidClientCertificate"

// dnMappingPrereq is appended to the terminal-Unauthorized message when mqweb rejects the
// presented client certificate (AC3). Mapping the certificate DN to an mqweb user in the
// queue-manager registry is a DEPLOYMENT prerequisite MKurator does NOT configure
// (ADR-0002 stance — documented, not implemented). MKurator never tries to configure
// mqweb; this message points the operator at that external prerequisite.
const dnMappingPrereq = "mqweb rejected the client certificate: verify the certificate is " +
	"trusted by mqweb and that its DN is mapped to an mqweb user in the queue manager's user " +
	"registry — this DN-to-user registry mapping is a deployment prerequisite MKurator does not configure"

// noopRequestAuthenticator is the ClientCert-mode request authenticator: authentication
// happens entirely at the transport (mTLS), so it attaches NO Authorization header. It is
// deliberately NOT nil — NewClient falls back to Basic when the authenticator is nil, which
// would send shared credentials (AC1: the admin identity carries no shared secret).
type noopRequestAuthenticator struct{}

func (noopRequestAuthenticator) authenticate(_ context.Context, _ *http.Request) error {
	return nil
}

// loadClientCertificate parses a kubernetes.io/tls-shaped keypair (tls.crt/tls.key PEM) into
// a tls.Certificate. A malformed or mismatched keypair is a CONFIGURATION-class error
// (AC2): the returned *mqadmin.TerminalError is non-transient so the reconciler surfaces it
// as Ready=False without a hot loop. The X509KeyPair error names the parse failure but
// carries no key bytes; we never echo the raw Secret data (NFR SEC-5).
func loadClientCertificate(certPEM, keyPEM []byte) (tls.Certificate, error) {
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, &mqadmin.TerminalError{
			Reason:  clientCertReason,
			Message: "invalid client certificate keypair (tls.crt/tls.key)",
			Cause:   err,
		}
	}
	return pair, nil
}

// NewClientCertClient builds a ClientCert-mode (mTLS) client. The keypair is loaded into
// cfg.TLSConfig.Certificates on the HTTP transport, so mqweb authenticates MKurator by the
// presented client certificate. Server-auth trust (cfg.TLSConfig.RootCAs from caSecretRef)
// is left untouched. No Authorization header is sent (AC1). A bad keypair surfaces as a
// configuration-class error before any request (AC2).
func NewClientCertClient(cfg Config, certPEM, keyPEM []byte) (*Client, error) {
	pair, err := loadClientCertificate(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	if cfg.TLSConfig == nil {
		cfg.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	} else {
		cfg.TLSConfig = cfg.TLSConfig.Clone()
	}
	cfg.TLSConfig.Certificates = append(cfg.TLSConfig.Certificates, pair)

	cfg.authenticator = noopRequestAuthenticator{}
	cfg.clientCertMode = true
	return NewClient(cfg)
}

// isClientCertRejection reports whether a transport-level request error is a server-side
// rejection of the presented client certificate during the TLS handshake. Such rejections
// surface as a server-originated TLS alert whose string carries "remote error: tls:" (Go's
// tls.AlertError leaf is unexported, so errors.As cannot match it — the string is the
// stable, observed discriminator; see the AUTH-16 probe). A purely LOCAL failure
// (connection refused, timeout, our own server-trust verification) does NOT carry the
// "remote error:" prefix and stays transient (AC3 negative).
func isClientCertRejection(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "remote error: tls")
}

// clientCertRejectionError maps a rejected-handshake error to the terminal Unauthorized
// condition carrying the DN-mapping prerequisite pointer (AC3).
func clientCertRejectionError() error {
	return &mqadmin.TerminalError{
		Reason:  reasonUnauthorized,
		Message: dnMappingPrereq,
	}
}

// unauthorizedMessage decorates a 401/403 message with the DN-mapping prerequisite pointer
// when the client is in ClientCert mode (AC3): an HTTP 401/403 in mTLS mode most often means
// the certificate was accepted at the transport but its DN is not mapped to an mqweb user.
// Basic/LTPA messages are returned unchanged.
func (c *Client) unauthorizedMessage(base string) string {
	if c.clientCertMode {
		return base + " — " + dnMappingPrereq
	}
	return base
}
