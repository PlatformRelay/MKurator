package validation

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestValidateQueueManagerConnectionSpec(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()

	t.Run("missing credentials secret", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "missing"},
		}
		if _, errs := ValidateQueueManagerConnectionSpec(context.Background(), cl, "ns", nil, spec); len(errs) == 0 {
			t.Fatal("expected secret not found error")
		}
	})
}

func TestValidateQueueManagerConnectionDeleteV1Beta1Dependents(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
		},
	}
	topic := &messagingv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "events", Namespace: "ns"},
		Spec: messagingv1beta1.TopicSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, queue, topic).Build()

	t.Run("deny with dependents", func(t *testing.T) {
		t.Parallel()
		if errs := ValidateQueueManagerConnectionDeleteV1Beta1(context.Background(), cl, conn); len(errs) == 0 {
			t.Fatal("expected delete blocked when dependents exist")
		}
	})
	t.Run("allow without dependents", func(t *testing.T) {
		t.Parallel()
		empty := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
		if errs := ValidateQueueManagerConnectionDeleteV1Beta1(context.Background(), empty, conn); len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
	})
}

func TestValidateQueueManagerConnectionInsecureTLS(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()

	baseSpec := func() *messagingv1beta1.QueueManagerConnectionSpec {
		return &messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "creds"},
			TLS:                  &messagingv1beta1.TLSConfig{InsecureSkipVerify: true},
		}
	}

	tests := []struct {
		name        string
		annotations map[string]string
		wantErr     bool
	}{
		{
			name:        "deny without opt-in annotation",
			annotations: nil,
			wantErr:     true,
		},
		{
			name:        "deny with false annotation",
			annotations: map[string]string{messagingv1beta1.AllowInsecureTLSAnnotation: "false"},
			wantErr:     true,
		},
		{
			name:        "deny with empty annotation",
			annotations: map[string]string{messagingv1beta1.AllowInsecureTLSAnnotation: ""},
			wantErr:     true,
		},
		{
			name:        "allow with true annotation",
			annotations: map[string]string{messagingv1beta1.AllowInsecureTLSAnnotation: "true"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, errs := ValidateQueueManagerConnectionSpec(context.Background(), cl, "ns", tt.annotations, baseSpec())
			if tt.wantErr && len(errs) == 0 {
				t.Fatal("expected insecure TLS error")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestValidateQueueManagerConnectionDeleteV1Beta1MultipleDependents(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
		},
	}
	topic := &messagingv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "events", Namespace: "ns"},
		Spec: messagingv1beta1.TopicSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, queue, topic).Build()
	errs := ValidateQueueManagerConnectionDeleteV1Beta1(context.Background(), cl, conn)
	if len(errs) == 0 {
		t.Fatal("expected delete blocked")
	}
	detail := errs[0].Detail
	if !strings.Contains(detail, "Queue") || !strings.Contains(detail, "Topic") {
		t.Fatalf("detail = %q", detail)
	}
}

func TestValidateQueueManagerConnectionCASecret(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = messagingv1beta1.AddToScheme(scheme)
	creds := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"}}
	ca := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "ca", Namespace: "ns"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(creds, ca).Build()

	spec := &messagingv1beta1.QueueManagerConnectionSpec{
		QueueManager:         "QM1",
		Endpoint:             "https://mq.example:9443",
		CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "creds"},
		TLS: &messagingv1beta1.TLSConfig{
			CASecretRef: &messagingv1beta1.SecretReference{Name: "ca"},
		},
	}
	if _, errs := ValidateQueueManagerConnectionSpec(context.Background(), cl, "ns", nil, spec); len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateQueueManagerConnectionCredentialsUsernameWarning(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	t.Run("warn when username key missing", func(t *testing.T) {
		t.Parallel()
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"},
			Data:       map[string][]byte{"password": []byte("x")},
		}
		cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
		spec := sampleConnection("ns", "qm1").Spec
		warnings, errs := ValidateQueueManagerConnectionSpec(context.Background(), cl, "ns", nil, &spec)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(warnings) != 1 || !strings.Contains(warnings[0], `default to "admin"`) {
			t.Fatalf("warnings = %v", warnings)
		}
	})
	t.Run("no warning when username present", func(t *testing.T) {
		t.Parallel()
		for _, key := range credentialsSecretUsernameKeys {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"},
				Data:       map[string][]byte{key: []byte("mquser"), "password": []byte("x")},
			}
			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
			spec := sampleConnection("ns", "qm1").Spec
			warnings, errs := ValidateQueueManagerConnectionSpec(context.Background(), cl, "ns", nil, &spec)
			if len(errs) > 0 {
				t.Fatalf("key %q: unexpected errors: %v", key, errs)
			}
			if len(warnings) != 0 {
				t.Fatalf("key %q: warnings = %v", key, warnings)
			}
		}
	})
}

func TestListConnectionDependentsAcrossKindsPreservesKindOrder(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)

	ref := messagingv1beta1.LocalObjectReference{Name: "qm1"}
	// listConnectionDependents lists each kind in a fixed order (Queue, then Topic, ...) and
	// appends in that order, so a Queue dependent always precedes a Topic dependent.
	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "ns"},
		Spec:       messagingv1beta1.QueueSpec{ConnectionRef: ref, QueueName: "APP.ORDERS"},
	}
	topic := &messagingv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "events", Namespace: "ns"},
		Spec:       messagingv1beta1.TopicSpec{ConnectionRef: ref, TopicName: "RETAIL.ORDERS"},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(queue, topic).Build()

	dependents, errs := listConnectionDependents(context.Background(), cl, "ns", "qm1")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := []connectionDependent{
		{kind: "Queue", name: "orders"},
		{kind: "Topic", name: "events"},
	}
	if len(dependents) != len(want) {
		t.Fatalf("dependents = %+v, want %+v", dependents, want)
	}
	for i := range want {
		if dependents[i] != want[i] {
			t.Fatalf("dependents[%d] = %+v, want %+v (full: %+v)", i, dependents[i], want[i], dependents)
		}
	}
}

func TestValidateQueueManagerConnectionDeleteWithV1Beta1Dependents(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)

	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
		},
	}
	topic := &messagingv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "events", Namespace: "ns"},
		Spec: messagingv1beta1.TopicSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	channel := &messagingv1beta1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: "orders-app", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		},
	}
	car := &messagingv1beta1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{Name: "car1", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1beta1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		},
	}
	auth := &messagingv1beta1.AuthorityRecord{
		ObjectMeta: metav1.ObjectMeta{Name: "auth1", Namespace: "ns"},
		Spec: messagingv1beta1.AuthorityRecordSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1beta1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{"GET"},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(conn, queue, topic, channel, car, auth).
		Build()

	errs := ValidateQueueManagerConnectionDeleteV1Beta1(context.Background(), cl, conn)
	if len(errs) == 0 {
		t.Fatal("expected delete blocked")
	}
}
