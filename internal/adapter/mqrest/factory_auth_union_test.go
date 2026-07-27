package mqrest

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// unionTestScheme registers v1alpha1 (spoke, what the factory is handed), v1beta1
// (hub, what the factory re-reads for the authentication union) and core/v1 (Secrets).
func unionTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := messagingv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
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
// credentialsSecretRef, which is nil for a union-only spec). Because the reconciler hands
// the factory a down-converted v1alpha1 spoke (union dropped), the factory must re-read the
// v1beta1 hub by identity to see the union.
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
	// Spoke as the reconciler sees it: union dropped, legacy ref empty.
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns, Generation: 3},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, hub, spoke).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	name, err := f.resolveCredentialsSecretName(ctx, spoke)
	if err != nil {
		t.Fatalf("resolveCredentialsSecretName: %v", err)
	}
	if name != "union-creds" {
		t.Fatalf("expected union-creds, got %q", name)
	}

	// End-to-end: buildConfig must read the union secret and populate credentials.
	cfg, err := f.buildConfig(ctx, spoke, resolvedAuth{secretName: name, mode: authModeBasic})
	if err != nil {
		t.Fatalf("buildConfig: %v", err)
	}
	if cfg.Username != "svc" || cfg.Password != "s3cret" {
		t.Fatalf("credentials not resolved from union secret: user=%q pass=%q", cfg.Username, cfg.Password)
	}

	// Sharpest constraint (ADR-0023): the fingerprint must key off the union secret's RV.
	fp, err := f.cacheFingerprint(ctx, spoke, name)
	if err != nil {
		t.Fatalf("cacheFingerprint: %v", err)
	}
	if fp.credRV != "7" {
		t.Fatalf("fingerprint must track the union secret RV, got %q", fp.credRV)
	}

	// End-to-end via the public seam (AC3): ForConnection resolves+builds+caches without
	// error, and a second call returns the same cached client (identity-keyed, ADR-0023).
	// NewClient only constructs the client (no network), so no live mqweb is needed.
	admin1, err := f.ForConnection(ctx, spoke)
	if err != nil {
		t.Fatalf("ForConnection: %v", err)
	}
	admin2, err := f.ForConnection(ctx, spoke)
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
	ctx := context.Background()
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
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-legacy", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://ibm-mq.ibm-mq.svc:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "legacy-creds"},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(legacySecret, hub, spoke).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	name, err := f.resolveCredentialsSecretName(ctx, spoke)
	if err != nil {
		t.Fatalf("resolveCredentialsSecretName: %v", err)
	}
	if name != "legacy-creds" {
		t.Fatalf("expected legacy-creds, got %q", name)
	}
}

// When the v1beta1 hub is not registered in the scheme (older/fake clients) or the hub
// object is absent, the factory falls back to the spoke's CredentialsSecretRef — preserving
// legacy behaviour verbatim rather than failing.
func TestClientFactory_HubUnreadableFallsBackToSpokeRef(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"

	// Scheme WITHOUT v1beta1: hub Get returns a not-registered error -> legacy fallback.
	s := runtime.NewScheme()
	if err := messagingv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-fallback", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://ibm-mq.ibm-mq.svc:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "spoke-creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(spoke).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	name, err := f.resolveCredentialsSecretName(ctx, spoke)
	if err != nil {
		t.Fatalf("resolveCredentialsSecretName: %v", err)
	}
	if name != "spoke-creds" {
		t.Fatalf("expected spoke-creds fallback, got %q", name)
	}
}

// A non-NotFound, non-scheme hub read error is propagated, not swallowed: swallowing would
// yield empty credentials and a confusing downstream "secret not found" for the wrong name.
func TestClientFactory_HubReadErrorPropagates(t *testing.T) {
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionTestScheme(t)
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-err", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://ibm-mq.ibm-mq.svc:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "spoke-creds"},
		},
	}
	boom := errors.New("apiserver unavailable")
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(spoke).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(
				_ context.Context, _ client.WithWatch, _ client.ObjectKey,
				obj client.Object, _ ...client.GetOption,
			) error {
				if _, ok := obj.(*messagingv1beta1.QueueManagerConnection); ok {
					return boom
				}
				return errors.New("unexpected get")
			},
		}).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	if _, err := f.resolveCredentialsSecretName(ctx, spoke); err == nil {
		t.Fatal("expected hub read error to propagate, got nil")
	} else if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got %v", err)
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
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-ltpa", Namespace: ns, Generation: 4},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(ltpaSecret, hub, spoke).Build()
	f := NewClientFactory(cl).(*ClientFactory)

	auth, err := f.resolveAuthentication(ctx, spoke)
	if err != nil {
		t.Fatalf("resolveAuthentication: %v", err)
	}
	if auth.mode != authModeLTPA {
		t.Fatalf("mode = %v, want authModeLTPA", auth.mode)
	}
	if auth.secretName != "ltpa-creds" {
		t.Fatalf("secretName = %q, want ltpa-creds", auth.secretName)
	}

	cfg, err := f.buildConfig(ctx, spoke, auth)
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
