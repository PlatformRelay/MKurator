package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	messagingv1alpha1 "github.com/platformrelay/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// unionWatchScheme registers v1alpha1 (spokes the watch lists), v1beta1 (the hub the watch
// re-reads for the authentication union) and core/v1 (Secrets).
func unionWatchScheme(t *testing.T) *runtime.Scheme {
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

// AUTH-14 AC2: a union QMC references its auth Secret through authentication.basic.secretRef,
// NOT the legacy credentialsSecretRef (which is empty on the down-converted spoke). The Secret
// watch must re-read the v1beta1 hub to see the union ref and enqueue the owning QMC when that
// Secret changes — closing the AUTH-12 rotation gap on the event-driven side.
func TestRequestsForSecret_EnqueuesUnionAuthRef(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	unionSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "1"},
		Data:       map[string][]byte{"password": []byte("old")},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
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
	// Spoke as the watch lists it: union dropped, legacy ref empty.
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(unionSecret, hub, spoke).Build()
	reqs := requestsForSecret(ctx, cl, unionSecret)
	if len(reqs) != 1 {
		t.Fatalf("requests = %d, want 1 (union auth-secret ref must enqueue owning QMC)", len(reqs))
	}
	if reqs[0].Name != "qm-union" || reqs[0].Namespace != ns {
		t.Fatalf("request = %+v", reqs[0])
	}
}

// AUTH-14 AC2 (stripped-data informer path): when the informer cache strips Secret data, a
// rotation surfaces only as a resourceVersion bump. The predicate must still fire AND the
// mapper must still enqueue the owning union QMC end-to-end.
func TestSecretWatch_UnionRefStrippedDataRotationEnqueues(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	// Stripped Secret in the informer cache: no Data, only a resourceVersion.
	strippedOld := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "1"}}
	strippedNew := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "union-creds", Namespace: ns, ResourceVersion: "2"}}

	hub := &messagingv1beta1.QueueManagerConnection{
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
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}

	// Predicate: stripped -> stripped RV-only change must pass (unchanged behaviour, asserted here for the union ref).
	preds := secretWatchPredicates()
	if !preds.Update(event.UpdateEvent{ObjectOld: strippedOld, ObjectNew: strippedNew}) {
		t.Fatal("expected predicate to fire on stripped-data resourceVersion change for union secret")
	}

	// Mapper: the enqueue path must resolve the union ref and enqueue the owning QMC.
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(strippedNew, hub, spoke).Build()
	mapper := secretEnqueueMapper(cl)
	reqs := mapper(ctx, strippedNew)
	if len(reqs) != 1 || reqs[0].Name != "qm-union" {
		t.Fatalf("reqs = %+v, want single enqueue of qm-union", reqs)
	}
}

// AUTH-14 AC2 (isolation): a union QMC whose auth Secret is a DIFFERENT name is not enqueued
// by an unrelated Secret change.
func TestRequestsForSecret_UnionRefNoMatchForOtherSecret(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"
	s := unionWatchScheme(t)

	otherSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: ns, ResourceVersion: "1"},
	}
	hub := &messagingv1beta1.QueueManagerConnection{
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
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager: "QM1",
			Endpoint:     "https://ibm-mq.ibm-mq.svc:9443",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(otherSecret, hub, spoke).Build()
	reqs := requestsForSecret(ctx, cl, otherSecret)
	if len(reqs) != 0 {
		t.Fatalf("requests = %d, want 0 (unrelated secret must not enqueue union QMC)", len(reqs))
	}
}

// AUTH-14 AC2 (legacy compat / hub unreadable): when the v1beta1 hub is not registered in the
// scheme (older/fake clients) the watch degrades to spoke-only matching — preserving the
// pre-AUTH-14 behaviour verbatim rather than failing the whole enqueue.
func TestRequestsForSecret_HubUnreadableFallsBackToSpokeRef(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := "mkurator-system"

	// Scheme WITHOUT v1beta1: hub Get returns not-registered -> spoke-only fallback.
	s := runtime.NewScheme()
	if err := messagingv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "legacy-creds", Namespace: ns, ResourceVersion: "1"},
	}
	spoke := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm-legacy", Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "legacy-creds"},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret, spoke).Build()
	reqs := requestsForSecret(ctx, cl, secret)
	if len(reqs) != 1 || reqs[0].Name != "qm-legacy" {
		t.Fatalf("reqs = %+v, want spoke-only enqueue of qm-legacy on hub-unreadable fallback", reqs)
	}
}
