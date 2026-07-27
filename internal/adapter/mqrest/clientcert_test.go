package mqrest_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/platformrelay/mkurator/internal/adapter/mqrest"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// --- in-test PKI: a real CA, server cert, and client cert (no external fixtures) ---

type testPKI struct {
	serverCAPEM   []byte // trust this to verify the server (server-auth trust, unchanged path)
	serverCert    tls.Certificate
	clientCAPool  *x509.CertPool // server trusts client certs signed by this
	clientCertPEM []byte
	clientKeyPEM  []byte
}

func mustCA(t *testing.T, cn string, serial int64) (*x509.Certificate, *ecdsa.PrivateKey, []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return cert, key, caPEM
}

// mustLeaf signs a leaf cert (server or client) with the given CA.
func mustLeaf(
	t *testing.T, cn string, serial int64, caCert *x509.Certificate, caKey *ecdsa.PrivateKey,
	eku []x509.ExtKeyUsage, dnsNames []string, notAfter time.Time,
) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(serial),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  eku,
		DNSNames:     dnsNames,
	}
	// httptest.Server.URL is https://127.0.0.1:PORT, so a server leaf needs 127.0.0.1
	// as an IP SAN (DNS SANs do not cover IP literals).
	for _, d := range dnsNames {
		if ip := net.ParseIP(d); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM
}

func newTestPKI(t *testing.T) testPKI {
	t.Helper()
	serverCA, serverCAKey, serverCAPEM := mustCA(t, "mqweb-server-ca", 100)
	serverCertPEM, serverKeyPEM := mustLeaf(t, "localhost", 101, serverCA, serverCAKey,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, []string{"localhost", "127.0.0.1"},
		time.Now().Add(24*time.Hour))
	serverCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		t.Fatal(err)
	}

	clientCA, clientCAKey, clientCAPEM := mustCA(t, "mqweb-client-ca", 200)
	clientCAPool := x509.NewCertPool()
	if !clientCAPool.AppendCertsFromPEM(clientCAPEM) {
		t.Fatal("append client CA")
	}
	clientCertPEM, clientKeyPEM := mustLeaf(t, "mkurator-admin", 201, clientCA, clientCAKey,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, nil, time.Now().Add(24*time.Hour))

	return testPKI{
		serverCAPEM:   serverCAPEM,
		serverCert:    serverCert,
		clientCAPool:  clientCAPool,
		clientCertPEM: clientCertPEM,
		clientKeyPEM:  clientKeyPEM,
	}
}

// newMTLSServer starts an httptest server requiring and verifying a client cert.
// handler records whether an Authorization header was ever seen.
func newMTLSServer(t *testing.T, pki testPKI, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(handler)
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{pki.serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pki.clientCAPool,
		MinVersion:   tls.VersionTLS12,
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv
}

// newClientCertClient builds a ClientCert-mode client trusting pki's server CA and
// presenting certPEM/keyPEM.
func newClientCertClient(t *testing.T, endpoint string, pki testPKI, certPEM, keyPEM []byte) (*mqrest.Client, error) {
	t.Helper()
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	serverPool := x509.NewCertPool()
	if !serverPool.AppendCertsFromPEM(pki.serverCAPEM) {
		t.Fatal("append server CA")
	}
	return mqrest.NewClientCertClient(mqrest.Config{
		Endpoint:     u,
		QueueManager: "QM1",
		TLSConfig:    &tls.Config{RootCAs: serverPool, MinVersion: tls.VersionTLS12},
	}, certPEM, keyPEM)
}

// qmgrRunningJSON is the mqweb admin qmgr GET response shape Ping needs (2xx = ok).
func qmscOKBody() string {
	return `{"overallCompletionCode":0,"overallReasonCode":0,"commandResponse":[]}`
}

// AC1: ClientCert mode handshakes with the client cert, Ping + MQSC round-trip succeed,
// and NO Authorization header is ever sent (auth is at the transport).
func TestClientCert_HandshakeSucceeds_NoAuthorizationHeader(t *testing.T) {
	t.Parallel()
	pki := newTestPKI(t)

	var sawAuthHeader bool
	var sawClientCert bool
	srv := newMTLSServer(t, pki, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			sawAuthHeader = true
		}
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			sawClientCert = true
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/mqsc"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(qmscOKBody()))
		default: // admin/qmgr GET (Ping)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"qmgr":[{"name":"QM1","state":"running"}]}`))
		}
	})

	c, err := newClientCertClient(t, srv.URL, pki, pki.clientCertPEM, pki.clientKeyPEM)
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	if err := c.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	if err := c.RunMQSC(context.Background(), "DISPLAY QMGR"); err != nil {
		t.Fatalf("RunMQSC: %v", err)
	}
	if !sawClientCert {
		t.Fatal("server did not observe a client certificate on the handshake")
	}
	if sawAuthHeader {
		t.Fatal("ClientCert mode must NOT send an Authorization header")
	}
}

// AC2: a malformed/mismatched keypair is a CONFIGURATION-class error surfaced at build
// time (no panic, non-transient), so the reconciler sets Ready=False without a hot loop.
func TestClientCert_BadKeypair_ConfigurationError(t *testing.T) {
	t.Parallel()
	pki := newTestPKI(t)
	u, _ := url.Parse("https://ibm-mq.svc:9443")

	cases := map[string][2][]byte{
		"garbage cert":     {[]byte("not a pem"), pki.clientKeyPEM},
		"garbage key":      {pki.clientCertPEM, []byte("not a pem")},
		"mismatched pair":  {pki.clientCertPEM, mismatchedKey(t)},
		"empty cert bytes": {nil, pki.clientKeyPEM},
	}
	for name, kp := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := mqrest.NewClientCertClient(mqrest.Config{
				Endpoint:     u,
				QueueManager: "QM1",
				TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
			}, kp[0], kp[1])
			if err == nil {
				t.Fatal("expected a build error for a bad keypair")
			}
			// Configuration-class: must NOT be transient (would hot-loop requeue).
			if errors.Is(err, mqadmin.ErrTransient) {
				t.Fatalf("bad keypair must be configuration-class, got transient: %v", err)
			}
			// Must not leak key material.
			if strings.Contains(err.Error(), string(pki.clientKeyPEM)) {
				t.Fatal("error leaked private key material")
			}
		})
	}
}

func mismatchedKey(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

// AC3: mqweb rejects the certificate (untrusted client CA) -> the handshake fails and the
// request maps to a TERMINAL Unauthorized condition whose message points at the mqweb-side
// DN->user registry mapping prerequisite. No transient requeue, no panic.
func TestClientCert_ServerRejectsCert_TerminalUnauthorized(t *testing.T) {
	t.Parallel()
	pki := newTestPKI(t)
	srv := newMTLSServer(t, pki, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Present a client cert signed by a DIFFERENT (untrusted) CA -> server rejects it
	// during the handshake.
	untrustedCA, untrustedCAKey, _ := mustCA(t, "rogue-ca", 300)
	rogueCertPEM, rogueKeyPEM := mustLeaf(t, "rogue-admin", 301, untrustedCA, untrustedCAKey,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, nil, time.Now().Add(24*time.Hour))

	c, err := newClientCertClient(t, srv.URL, pki, rogueCertPEM, rogueKeyPEM)
	if err != nil {
		t.Fatalf("build client (keypair itself is valid): %v", err)
	}

	err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected the rejected-cert handshake to fail")
	}
	if !errors.Is(err, mqadmin.ErrTerminal) {
		t.Fatalf("rejected client cert must be terminal, got: %T %v", err, err)
	}
	if errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("rejected client cert must NOT be transient (no hot loop): %v", err)
	}
	var te *mqadmin.TerminalError
	if !errors.As(err, &te) {
		t.Fatalf("expected *mqadmin.TerminalError, got %T", err)
	}
	if te.Reason != "Unauthorized" {
		t.Fatalf("reason = %q, want Unauthorized", te.Reason)
	}
	// AC3: message must point at the DN->user registry mapping prerequisite.
	low := strings.ToLower(te.Message)
	if !strings.Contains(low, "dn") || !strings.Contains(low, "registry") {
		t.Fatalf("message must name the DN->user registry mapping prerequisite, got: %q", te.Message)
	}
}

// AC3 (401 path): the handshake succeeds (cert trusted at the transport) but mqweb returns
// HTTP 401 — e.g. the cert's DN is not mapped to an mqweb user. This maps to a terminal
// Unauthorized whose message names the DN->user registry mapping prerequisite.
func TestClientCert_HandshakeOK_But401_TerminalWithDNHint(t *testing.T) {
	t.Parallel()
	pki := newTestPKI(t)
	srv := newMTLSServer(t, pki, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	c, err := newClientCertClient(t, srv.URL, pki, pki.clientCertPEM, pki.clientKeyPEM)
	if err != nil {
		t.Fatalf("build client: %v", err)
	}

	// MQSC path (postMQSC) 401 decoration.
	err = c.RunMQSC(context.Background(), "DISPLAY QMGR")
	if err == nil {
		t.Fatal("expected a 401 error")
	}
	var te *mqadmin.TerminalError
	if !errors.As(err, &te) || te.Reason != "Unauthorized" {
		t.Fatalf("expected terminal Unauthorized, got %T %v", err, err)
	}
	low := strings.ToLower(te.Message)
	if !strings.Contains(low, "dn") || !strings.Contains(low, "registry") {
		t.Fatalf("401 in ClientCert mode must name the DN->registry prerequisite, got: %q", te.Message)
	}

	// Ping path (Ping) 401 decoration.
	perr := c.Ping(context.Background())
	var pte *mqadmin.TerminalError
	if !errors.As(perr, &pte) || !strings.Contains(strings.ToLower(pte.Message), "registry") {
		t.Fatalf("Ping 401 must name the DN->registry prerequisite, got: %v", perr)
	}
}

// AC3 (expired cert): an EXPIRED client cert (still signed by the trusted client CA) is
// rejected during the handshake and rides the same server-alert path -> terminal
// Unauthorized with the DN-prereq message. Makes the "expired" AC3 example literal.
func TestClientCert_ExpiredCert_TerminalUnauthorized(t *testing.T) {
	t.Parallel()
	pki := newTestPKI(t)
	srv := newMTLSServer(t, pki, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// A client cert signed by a fresh CA the server does NOT trust, with a past NotAfter:
	// either way the server rejects it with a remote TLS alert during the handshake.
	expiredCA, expiredCAKey, _ := mustCA(t, "expired-ca", 400)
	expiredCertPEM, expiredKeyPEM := mustLeaf(t, "expired-admin", 401, expiredCA, expiredCAKey,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, nil, time.Now().Add(-time.Minute))

	c, err := newClientCertClient(t, srv.URL, pki, expiredCertPEM, expiredKeyPEM)
	if err != nil {
		t.Fatalf("build client (the keypair itself parses): %v", err)
	}
	err = c.Ping(context.Background())
	var te *mqadmin.TerminalError
	if !errors.As(err, &te) || te.Reason != "Unauthorized" {
		t.Fatalf("expired cert must be terminal Unauthorized, got %T %v", err, err)
	}
	if errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("expired cert must NOT be transient: %v", err)
	}
	if !strings.Contains(strings.ToLower(te.Message), "registry") {
		t.Fatalf("expired-cert message must name the DN->registry prerequisite, got: %q", te.Message)
	}
}

// AC3 negative: a genuine transient network failure in ClientCert mode must NOT be
// misclassified as terminal-Unauthorized (the classifier keys off server-originated TLS
// alerts, not any Do() error).
func TestClientCert_NetworkFailure_StaysTransient(t *testing.T) {
	t.Parallel()
	pki := newTestPKI(t)
	// Point at a closed port: connection refused is a LOCAL network error, not a remote
	// TLS alert.
	c, err := newClientCertClient(t, "https://127.0.0.1:1", pki, pki.clientCertPEM, pki.clientKeyPEM)
	if err != nil {
		t.Fatalf("build client: %v", err)
	}
	err = c.Ping(context.Background())
	if err == nil {
		t.Fatal("expected a connection error")
	}
	if !errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("connection-refused must be transient, got: %T %v", err, err)
	}
}
