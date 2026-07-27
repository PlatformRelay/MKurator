package mqrest

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
	"net/http"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// genKeypair returns a valid self-signed tls.crt/tls.key PEM pair for factory tests.
func genKeypair(t *testing.T, cn string) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
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

// clientCertFixtures builds the tls Secret, v1beta1 hub (ClientCert union) and the
// down-converted v1alpha1 spoke the reconciler hands the factory.
func clientCertFixtures(
	t *testing.T, ns, secretName, rv string,
) (*messagingv1beta1.QueueManagerConnection, *messagingv1alpha1.QueueManagerConnection, *corev1.Secret) {
	t.Helper()
	certPEM, keyPEM := genKeypair(t, "mkurator-admin")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns, ResourceVersion: rv},
		Type:       corev1.SecretTypeTLS,
		Data:       map[string][]byte{"tls.crt": certPEM, "tls.key": keyPEM},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-mtls", Namespace: ns, Generation: 1},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode: messagingv1beta1.MQWebAuthenticationModeClientCert,
				ClientCert: &messagingv1beta1.ClientCertAuth{
					SecretRef: messagingv1beta1.SecretReference{Name: secretName},
				},
			},
		},
	}
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-mtls", Namespace: ns, Generation: 1},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}
	return hub, spoke, secret
}

// AC1: the factory resolves ClientCert mode from the union, reading the tls Secret (not the
// empty legacy credentialsSecretRef), and builds a client with the keypair on the transport
// and a no-op (no Authorization header) authenticator.
func TestClientFactory_ResolvesClientCertUnion(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	hub, spoke, secret := clientCertFixtures(t, ns, "mtls-creds", "1")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret, hub, spoke).Build()
	factory := NewClientFactory(cl).(*ClientFactory)

	auth, err := factory.resolveAuthentication(ctx, spoke)
	if err != nil {
		t.Fatalf("resolveAuthentication: %v", err)
	}
	if auth.secretName != "mtls-creds" {
		t.Fatalf("secretName = %q, want mtls-creds", auth.secretName)
	}
	if auth.mode != authModeClientCert {
		t.Fatalf("mode = %v, want authModeClientCert", auth.mode)
	}

	admin, err := factory.ForConnection(ctx, spoke)
	if err != nil {
		t.Fatalf("ForConnection: %v", err)
	}
	c, ok := admin.(*Client)
	if !ok {
		t.Fatalf("expected *Client, got %T", admin)
	}
	if !c.clientCertMode {
		t.Fatal("expected clientCertMode set on the built client")
	}
	if _, isNoop := c.authenticator.(noopRequestAuthenticator); !isNoop {
		t.Fatalf("expected noopRequestAuthenticator, got %T", c.authenticator)
	}
	// The keypair must be on the transport's tls.Config.
	tr, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected *http.Transport, got %T", c.httpClient.Transport)
	}
	if tr.TLSClientConfig == nil || len(tr.TLSClientConfig.Certificates) != 1 {
		t.Fatal("expected exactly one client certificate on the transport")
	}
}

// AC2: a tls Secret with a malformed/mismatched keypair yields a configuration-class error
// from the factory build (non-transient), so the reconciler sets Ready=False without a hot
// loop. No panic.
func TestClientFactory_ClientCertBadKeypairIsConfigurationError(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	hub, spoke, secret := clientCertFixtures(t, ns, "mtls-creds", "1")
	secret.Data["tls.crt"] = []byte("-----BEGIN CERTIFICATE-----\nnonsense\n-----END CERTIFICATE-----")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret, hub, spoke).Build()
	factory := NewClientFactory(cl)

	_, err := factory.ForConnection(ctx, spoke)
	if err == nil {
		t.Fatal("expected a build error for a bad keypair")
	}
	if errors.Is(err, mqadmin.ErrTransient) {
		t.Fatalf("bad keypair must be configuration-class, not transient: %v", err)
	}
	var te *mqadmin.TerminalError
	if !errors.As(err, &te) || te.Reason != clientCertReason {
		t.Fatalf("expected TerminalError{Reason:%s}, got %T %v", clientCertReason, err, err)
	}
}

// AC4: rotating the tls Secret (resourceVersion change) makes ForConnection build a
// replacement client and close the old transport — reusing the AUTH-14 rotation contract
// with a cert Secret (the fingerprint already keys off the union secret's resourceVersion).
func TestClientFactory_ClientCertSecretRotationReplacesAndCloses(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	hub, spoke, secret := clientCertFixtures(t, ns, "mtls-creds", "1")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret, hub, spoke).Build()

	tr := &idleTrackingTransport{}
	factory := NewClientFactory(cl).(*ClientFactory)
	factory.newClient = func(cfg Config) (mqadmin.Admin, error) {
		// The tracking transport replaces the keypair-bearing transport; the swap/close
		// contract is what this test proves (the handshake proof is the httptest test).
		cfg.HTTPClient = &http.Client{Transport: tr}
		return NewClient(cfg)
	}

	c1, err := factory.ForConnection(ctx, spoke)
	if err != nil {
		t.Fatalf("ForConnection first: %v", err)
	}

	rotateCredSecret(ctx, t, cl, secret, "2")

	c2, err := factory.ForConnection(ctx, spoke)
	if err != nil {
		t.Fatalf("ForConnection after cert rotation: %v", err)
	}
	if c1 == c2 {
		t.Fatal("expected a new client after tls Secret rotation")
	}
	if got := tr.closeCalls.Load(); got != 1 {
		t.Fatalf("CloseIdleConnections calls = %d, want 1 (ADR-0023 replace-on-mismatch)", got)
	}
}

// Unit coverage for the small classifier/label helpers (AUTH-16).
func TestClientCert_Helpers(t *testing.T) {
	if isClientCertRejection(nil) {
		t.Fatal("nil error must not be a client-cert rejection")
	}
	if isClientCertRejection(errors.New("dial tcp: connection refused")) {
		t.Fatal("a local net error must not be a client-cert rejection")
	}
	if !isClientCertRejection(errors.New(`Get "https://x": remote error: tls: unknown certificate authority`)) {
		t.Fatal("a remote TLS alert must be a client-cert rejection")
	}
	if got := credSecretRole(authModeClientCert); got != "client certificate" {
		t.Fatalf("credSecretRole(ClientCert) = %q", got)
	}
	if got := credSecretRole(authModeBasic); got != "credentials" {
		t.Fatalf("credSecretRole(Basic) = %q", got)
	}
	// Basic mode leaves the Unauthorized message undecorated.
	basic := &Client{clientCertMode: false}
	if basic.unauthorizedMessage("base") != "base" {
		t.Fatal("Basic mode must not decorate the Unauthorized message")
	}
	cc := &Client{clientCertMode: true}
	if !strings.Contains(cc.unauthorizedMessage("base"), "registry") {
		t.Fatal("ClientCert mode must decorate with the DN->registry prerequisite")
	}
}

// keep tls import used even if the file's assertions change.
var _ = tls.VersionTLS12
