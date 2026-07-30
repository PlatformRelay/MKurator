package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// unionWatchScheme registers v1beta1 (the QMC the watch lists natively) and core/v1 (Secrets).
func unionWatchScheme(t *testing.T) *runtime.Scheme {
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

// A union QMC references its auth Secret through authentication.basic.secretRef, NOT the legacy
// credentialsSecretRef. The Secret watch lists the v1beta1 QMC natively and reads the union ref
// directly, enqueuing the owning QMC when that Secret changes (union rotation, AUTH-12).
func TestRequestsForSecret_EnqueuesUnionAuthRef(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	unionSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "1"},
		Data:       map[string][]byte{"password": []byte("old")},
	}
	qmc := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode:  messagingv1beta1.MQWebAuthenticationModeBasic,
				Basic: &messagingv1beta1.BasicAuth{SecretRef: messagingv1beta1.SecretReference{Name: "union-creds"}},
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, qmc).Build()
	reqs := requestsForSecret(ctx, cl, unionSecret)
	if len(reqs) != 1 {
		t.Fatalf("requests = %d, want 1 (union auth-secret ref must enqueue owning QMC)", len(reqs))
	}
	if reqs[0].Name != "qm-union" || reqs[0].Namespace != ns {
		t.Fatalf("request = %+v", reqs[0])
	}
}

// Stripped-data informer path: when the informer cache strips Secret data, a rotation surfaces
// only as a resourceVersion bump. The predicate must still fire AND the mapper must still enqueue
// the owning union QMC end-to-end.
func TestSecretWatch_UnionRefStrippedDataRotationEnqueues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	// Stripped Secret in the informer cache: no Data, only a resourceVersion.
	strippedOld := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "1"},
	}
	strippedNew := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "2"},
	}

	qmc := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode:  messagingv1beta1.MQWebAuthenticationModeBasic,
				Basic: &messagingv1beta1.BasicAuth{SecretRef: messagingv1beta1.SecretReference{Name: "union-creds"}},
			},
		},
	}

	// Predicate: stripped -> stripped RV-only change must pass (unchanged behaviour, asserted here for the union ref).
	preds := secretWatchPredicates()
	if !preds.Update(event.UpdateEvent{ObjectOld: strippedOld, ObjectNew: strippedNew}) {
		t.Fatal("expected predicate to fire on stripped-data resourceVersion change for union secret")
	}

	// Mapper: the enqueue path must resolve the union ref and enqueue the owning QMC.
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(strippedNew, qmc).Build()
	mapper := secretEnqueueMapper(cl)
	reqs := mapper(ctx, strippedNew)
	if len(reqs) != 1 || reqs[0].Name != "qm-union" {
		t.Fatalf("reqs = %+v, want single enqueue of qm-union", reqs)
	}
}

// Generic over union members: connectionReferencesSecret matches EVERY set member ref
// (basic/ltpa/clientCert), independent of mode, so AUTH-13/16 inherit the watch. CEL exclusivity
// is not enforced in unit tests, so we set all three members to lock the generic extraction.
func TestConnectionReferencesSecret_MatchesEveryUnionMember(t *testing.T) {
	t.Parallel()

	qmc := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-all", Namespace: "mkurator-system"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode:  messagingv1beta1.MQWebAuthenticationModeBasic,
				Basic: &messagingv1beta1.BasicAuth{SecretRef: messagingv1beta1.SecretReference{Name: "basic-ref"}},
				LTPA:  &messagingv1beta1.LTPAAuth{SecretRef: messagingv1beta1.SecretReference{Name: "ltpa-ref"}},
				ClientCert: &messagingv1beta1.ClientCertAuth{
					SecretRef: messagingv1beta1.SecretReference{Name: "cc-ref"},
				},
			},
		},
	}

	for _, name := range []string{"basic-ref", "ltpa-ref", "cc-ref"} {
		if !connectionReferencesSecret(qmc, name) {
			t.Fatalf("union member %q should match via connectionReferencesSecret", name)
		}
	}
	if connectionReferencesSecret(qmc, "unrelated") {
		t.Fatal("unrelated secret name must not match")
	}
}

// Isolation: a union QMC whose auth Secret is a DIFFERENT name is not enqueued by an unrelated
// Secret change.
func TestRequestsForSecret_UnionRefNoMatchForOtherSecret(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	otherSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: ns, ResourceVersion: "1"},
	}
	qmc := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
			Authentication: &messagingv1beta1.MQWebAuthentication{
				Mode:  messagingv1beta1.MQWebAuthenticationModeBasic,
				Basic: &messagingv1beta1.BasicAuth{SecretRef: messagingv1beta1.SecretReference{Name: "union-creds"}},
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(otherSecret, qmc).Build()
	reqs := requestsForSecret(ctx, cl, otherSecret)
	if len(reqs) != 0 {
		t.Fatalf("requests = %d, want 0 (unrelated secret must not enqueue union QMC)", len(reqs))
	}
}

// Stale-cache graceful degradation (8e-2): the informer may briefly return a QMC whose
// authentication union is not yet populated. The mapper must NOT panic on the now-optional
// pointer fields — it matches on whatever refs ARE readable.
//   - authentication nil + credentialsSecretRef = rotated secret -> still enqueues (matches the ref).
//   - authentication nil + credentialsSecretRef nil (fully partial read) -> no panic, no enqueue.
func TestRequestsForSecret_StaleCacheNilAuthenticationDegradesGracefully(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "rotated", Namespace: ns, ResourceVersion: "2"},
	}
	// authentication nil, but the legacy credential ref IS present: match on it.
	legacy := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-legacy", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://ibm-mq.ibm-mq.svc:9443",
			CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "rotated"},
		},
	}
	// authentication nil AND credentialsSecretRef nil: a partially-read object. Must not panic.
	partial := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-partial", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager: "QM2",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret, legacy, partial).Build()
	reqs := requestsForSecret(ctx, cl, secret)
	if len(reqs) != 1 || reqs[0].Name != "qm-legacy" {
		t.Fatalf(
			"reqs = %+v, want single enqueue of qm-legacy (matched on credentialsSecretRef despite nil auth)",
			reqs,
		)
	}
}
