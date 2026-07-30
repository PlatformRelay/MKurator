package mqrest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

// unionRotationFixtures builds a v1beta1 hub whose authentication union references a Secret
// DISTINCT from any legacy credentialsSecretRef, plus the union Secret itself. The reconciler
// hands the factory this hub directly (8e-1), so there is no down-converted spoke.
func unionRotationFixtures(
	ns, unionSecretName, credRV string,
) (*messagingv1beta1.QueueManagerConnection, *corev1.Secret) {
	unionSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: unionSecretName, Namespace: ns, ResourceVersion: credRV},
		Data: map[string][]byte{
			"username":        []byte("svc"),
			"mqAdminPassword": []byte("s3cret"),
		},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns, Generation: 1},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode:  messagingv1beta1.MQWebAuthenticationModeBasic,
				Basic: &messagingv1beta1.BasicAuth{SecretRef: messagingv1beta1.SecretReference{Name: unionSecretName}},
			},
		},
	}
	return hub, unionSecret
}

// AUTH-14 AC1: rotating a union auth Secret (distinct from credentialsSecretRef) makes
// ForConnection build a replacement client, swap it in, and close the OLD transport's idle
// connections (ADR-0023 replace-on-mismatch). The cache stays bounded at one entry across N
// rotations. This closes the AUTH-12 rotation gap: the fingerprint keys off the union
// Secret's resourceVersion (via resolveCredentialsSecretName -> credRV).
func TestClientFactory_UnionSecretRotationReplacesAndCloses(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	hub, unionSecret := unionRotationFixtures(ns, "union-creds", "1")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, hub).Build()

	tr := &idleTrackingTransport{}
	factory := NewClientFactory(cl).(*ClientFactory)
	factory.newClient = func(cfg Config) (mqadmin.Admin, error) {
		cfg.HTTPClient = &http.Client{Transport: tr}
		return NewClient(cfg)
	}

	c1, err := factory.ForConnection(ctx, hub)
	if err != nil {
		t.Fatalf("ForConnection first: %v", err)
	}

	rotateCredSecret(ctx, t, cl, unionSecret, "2")

	c2, err := factory.ForConnection(ctx, hub)
	if err != nil {
		t.Fatalf("ForConnection after union rotation: %v", err)
	}
	if c1 == c2 {
		t.Fatal("expected a new client after union secret rotation (AUTH-12 rotation gap must be closed)")
	}
	if got := tr.closeCalls.Load(); got != 1 {
		t.Fatalf("CloseIdleConnections calls = %d, want 1 (ADR-0023 replace-on-mismatch)", got)
	}

	factory.mu.Lock()
	n := len(factory.cache)
	factory.mu.Unlock()
	if n != 1 {
		t.Fatalf("expected single cache entry after union rotation, got %d", n)
	}
}

// AUTH-14 AC1 (bounded): the cache stays at one entry per live QMC across N union-secret
// rotations, and each replacement closes exactly one old transport.
func TestClientFactory_UnionCacheBoundedAcrossRotations(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	hub, unionSecret := unionRotationFixtures(ns, "union-creds", "0")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, hub).Build()

	tr := &idleTrackingTransport{}
	factory := NewClientFactory(cl).(*ClientFactory)
	factory.newClient = func(cfg Config) (mqadmin.Admin, error) {
		cfg.HTTPClient = &http.Client{Transport: tr}
		return NewClient(cfg)
	}

	const rotations = 25
	for i := 0; i < rotations; i++ {
		rotateCredSecret(ctx, t, cl, unionSecret, fmt.Sprintf("%d", i+1))
		if _, err := factory.ForConnection(ctx, hub); err != nil {
			t.Fatalf("ForConnection union rotation %d: %v", i, err)
		}
	}

	factory.mu.Lock()
	n := len(factory.cache)
	factory.mu.Unlock()
	if n != 1 {
		t.Fatalf("expected cache size 1 after %d union rotations, got %d", rotations, n)
	}
	if got := tr.closeCalls.Load(); got != rotations-1 {
		t.Fatalf("CloseIdleConnections calls = %d, want %d", got, rotations-1)
	}
}

// AUTH-14 AC3: releasing a union QMC whose auth Secret was deleted first (namespace teardown,
// ADR-0023 rule 3) succeeds — release reads no Secrets, so a fingerprint lookup failure must
// never block the finalizer.
func TestClientFactory_ReleaseUnionConnectionMissingSecret(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	hub, unionSecret := unionRotationFixtures(ns, "union-creds", "1")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, hub).Build()
	factory := NewClientFactory(cl)

	if _, err := factory.ForConnection(ctx, hub); err != nil {
		t.Fatalf("ForConnection: %v", err)
	}

	// Delete the union auth Secret AND the hub (namespace teardown removes both). Release must
	// still succeed: it evicts by namespace/name and treats missing referents as success.
	if err := cl.Delete(ctx, unionSecret); err != nil {
		t.Fatalf("delete union secret: %v", err)
	}
	if err := cl.Delete(ctx, hub); err != nil {
		t.Fatalf("delete hub: %v", err)
	}

	if err := factory.ReleaseConnection(ctx, hub); err != nil {
		t.Fatalf("ReleaseConnection with missing union secret: %v", err)
	}

	f := factory.(*ClientFactory)
	f.mu.Lock()
	n := len(f.cache)
	f.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected empty cache after release, got %d entries", n)
	}
}

// AUTH-14 AC4 (perf-neutral): a legacy QMC with NO authentication union performs exactly the
// Secret Gets it did before AUTH-14 — the credentials Secret once per ForConnection
// (fingerprint + buildConfig), plus one CA Get when TLS.caSecretRef is set. No extra union
// Secret Gets appear on the Basic path. Asserted with a counting fake client.
func TestClientFactory_BasicPathSecretGetCountUnchanged(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	credSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "legacy-creds", Namespace: ns, ResourceVersion: "1"},
		Data: map[string][]byte{
			"username":        []byte("admin"),
			"mqAdminPassword": []byte("passw0rd"),
		},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-legacy", Namespace: ns, Generation: 1},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://ibm-mq.ibm-mq.svc:9443",
			CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "legacy-creds"},
		},
	}
	counter := &secretGetCounter{}
	cl := fake.NewClientBuilder().WithScheme(s).
		WithObjects(credSecret, hub).
		WithInterceptorFuncs(counter.interceptor()).
		Build()
	factory := NewClientFactory(cl).(*ClientFactory)

	if _, err := factory.ForConnection(ctx, hub); err != nil {
		t.Fatalf("ForConnection: %v", err)
	}

	// The Basic path reads only the credentials Secret (fingerprint + buildConfig = 2 Gets),
	// no CA (no TLS.caSecretRef), and NO union Secret. Fingerprint reads exactly one distinct
	// Secret name: the credentials Secret.
	if got := counter.distinctNames(); len(got) != 1 || got["legacy-creds"] == 0 {
		t.Fatalf("Basic path read unexpected Secrets: %v (want only legacy-creds)", got)
	}
	if got := counter.total(); got != 2 {
		t.Fatalf("Basic-path Secret Gets = %d, want 2 (fingerprint + buildConfig, no union, no CA)", got)
	}
}

// secretGetCounter records how many times each Secret name is Get through a fake client.
type secretGetCounter struct {
	names map[string]int
}

func (c *secretGetCounter) interceptor() interceptor.Funcs {
	if c.names == nil {
		c.names = map[string]int{}
	}
	return interceptor.Funcs{
		Get: func(
			ctx context.Context, cl client.WithWatch, key client.ObjectKey,
			obj client.Object, opts ...client.GetOption,
		) error {
			if _, ok := obj.(*corev1.Secret); ok {
				c.names[key.Name]++
			}
			return cl.Get(ctx, key, obj, opts...)
		},
	}
}

func (c *secretGetCounter) total() int {
	n := 0
	for _, v := range c.names {
		n += v
	}
	return n
}

func (c *secretGetCounter) distinctNames() map[string]int {
	return c.names
}
