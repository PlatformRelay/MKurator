package mqrest

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// unionTestScheme registers v1beta1 (the hub the factory is handed directly, 8e-1) and
// core/v1 (Secrets).
func unionTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := messagingv1beta1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

// AUTH-12 AC3: a v1beta1 QMC with authentication mode Basic + basic.secretRef resolves
// its credentials end-to-end via the factory, reading THAT secret (not the legacy
// credentialsSecretRef, which is nil for a union-only spec). The reconciler hands the factory
// the hub directly (8e-1), so the union is read straight off the object — no hub re-read.
func TestClientFactory_ResolvesBasicUnionSecretName(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	unionSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "7"},
		Data:       map[string][]byte{"username": []byte("svc"), "password": []byte("s3cret")},
	}
	// Hub carries the union; legacy CredentialsSecretRef is nil.
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns, Generation: 3},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode:  messagingv1beta1.MQWebAuthenticationModeBasic,
				Basic: &messagingv1beta1.BasicAuth{SecretRef: messagingv1beta1.SecretReference{Name: "union-creds"}},
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, hub).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	name, err := f.resolveCredentialsSecretName(hub)
	if err != nil {
		t.Fatalf("resolveCredentialsSecretName: %v", err)
	}
	if name != "union-creds" {
		t.Fatalf("expected union-creds, got %q", name)
	}

	// End-to-end: buildConfig must read the union secret and populate credentials.
	cfg, err := f.buildConfig(ctx, hub, resolvedAuth{secretName: name, mode: authModeBasic})
	if err != nil {
		t.Fatalf("buildConfig: %v", err)
	}
	if cfg.Username != "svc" || cfg.Password != "s3cret" {
		t.Fatalf("credentials not resolved from union secret: user=%q pass=%q", cfg.Username, cfg.Password)
	}

	// Sharpest constraint (ADR-0023): the fingerprint must key off the union secret's RV.
	fp, err := f.cacheFingerprint(ctx, hub, name)
	if err != nil {
		t.Fatalf("cacheFingerprint: %v", err)
	}
	if fp.credRV != "7" {
		t.Fatalf("fingerprint must track the union secret RV, got %q", fp.credRV)
	}

	// End-to-end via the public seam (AC3): ForConnection resolves+builds+caches without
	// error, and a second call returns the same cached client (identity-keyed, ADR-0023).
	// NewClient only constructs the client (no network), so no live mqweb is needed.
	admin1, err := f.ForConnection(ctx, hub)
	if err != nil {
		t.Fatalf("ForConnection: %v", err)
	}
	admin2, err := f.ForConnection(ctx, hub)
	if err != nil {
		t.Fatalf("ForConnection (cached): %v", err)
	}
	if admin1 != admin2 {
		t.Fatal("expected the cached client on the second ForConnection call")
	}
}

// AUTH-12 AC1: a legacy QMC (no authentication union) still resolves credentials from
// credentialsSecretRef — the pre-ADR-0027 Basic-compat path, unchanged.
func TestClientFactory_LegacyCredentialsSecretRefStillResolves(t *testing.T) {
	ns := "mkurator-system"
	s := unionTestScheme(t)

	legacySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "legacy-creds", Namespace: ns, ResourceVersion: "1"},
		Data:       map[string][]byte{"username": []byte("admin"), "password": []byte("passw0rd")},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-legacy", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://ibm-mq.ibm-mq.svc:9443",
			CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "legacy-creds"},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(legacySecret, hub).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	name, err := f.resolveCredentialsSecretName(hub)
	if err != nil {
		t.Fatalf("resolveCredentialsSecretName: %v", err)
	}
	if name != "legacy-creds" {
		t.Fatalf("expected legacy-creds, got %q", name)
	}
}

// A spec carrying neither the authentication union nor credentialsSecretRef (CEL-rejected at
// admission, but defensively guarded) yields an explicit error rather than a nil-deref on the
// now-optional, pointer CredentialsSecretRef.
func TestClientFactory_NoAuthNoCredentialsErrors(t *testing.T) {
	s := unionTestScheme(t)
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-empty", Namespace: "mkurator-system"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(hub).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	if _, err := f.resolveCredentialsSecretName(hub); err == nil {
		t.Fatal("expected an error when neither authentication union nor credentialsSecretRef is set")
	}
}

// AUTH-13 AC1 (factory wiring): a v1beta1 QMC with authentication mode LTPA +
// ltpa.secretRef resolves to LTPA mode against THAT secret, and buildConfig
// produces an LTPA Config (cfg.LTPA set, no Basic authenticator). Rotation of the
// LTPA login Secret is covered by credRV (same secret name), not duplicated here.
func TestClientFactory_ResolvesLTPAUnion(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)

	ltpaSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "ltpa-creds", Namespace: ns, ResourceVersion: "11"},
		Data:       map[string][]byte{"username": []byte("ltpa-admin"), "password": []byte("ltpa-pass")},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-ltpa", Namespace: ns, Generation: 4},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode: messagingv1beta1.MQWebAuthenticationModeLTPA,
				LTPA: &messagingv1beta1.LTPAAuth{SecretRef: messagingv1beta1.SecretReference{Name: "ltpa-creds"}},
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(ltpaSecret, hub).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	auth, err := f.resolveAuthentication(hub)
	if err != nil {
		t.Fatalf("resolveAuthentication: %v", err)
	}
	if auth.mode != authModeLTPA {
		t.Fatalf("mode = %v, want authModeLTPA", auth.mode)
	}
	if auth.secretName != "ltpa-creds" {
		t.Fatalf("secretName = %q, want ltpa-creds", auth.secretName)
	}

	cfg, err := f.buildConfig(ctx, hub, auth)
	if err != nil {
		t.Fatalf("buildConfig: %v", err)
	}
	if cfg.LTPA == nil {
		t.Fatal("cfg.LTPA is nil; LTPA mode must build an LTPA config")
	}
	if cfg.LTPA.Username != "ltpa-admin" || cfg.LTPA.Password != "ltpa-pass" {
		t.Fatalf("LTPA creds not resolved: user=%q", cfg.LTPA.Username)
	}
	if cfg.authenticator != nil {
		t.Fatal("LTPA mode must not set a Basic authenticator")
	}
}
