package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestQueueManagerConnectionConvertToFromRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alpha      *QueueManagerConnection
		assertBeta func(t *testing.T, beta *messagingv1beta1.QueueManagerConnection)
		assertBack func(t *testing.T, orig, back *QueueManagerConnection)
	}{
		{
			name: "copies endpoint credentials and TLS config",
			alpha: &QueueManagerConnection{
				ObjectMeta: func() metav1.ObjectMeta {
					meta := testObjectMeta("qmc-tls")
					meta.Annotations = map[string]string{
						AllowInsecureTLSAnnotation: "true",
						"note":                     "fixture",
					}
					return meta
				}(),
				Spec: QueueManagerConnectionSpec{
					QueueManager: "QM1",
					Endpoint:     "https://mq.example:9443",
					RESTPrefix:   "/ibmmq/rest/v3",
					CredentialsSecretRef: SecretReference{
						Name: "creds",
					},
					TLS: &TLSConfig{
						InsecureSkipVerify: true,
						CASecretRef:        &SecretReference{Name: "ca"},
					},
				},
				Status: QueueManagerConnectionStatus{
					Conditions:         testSyncedCondition(),
					ObservedGeneration: 1,
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.QueueManagerConnection) {
				t.Helper()
				if beta.Spec.TLS == nil || !beta.Spec.TLS.InsecureSkipVerify {
					t.Fatalf("tls = %+v", beta.Spec.TLS)
				}
				if beta.Spec.TLS.CASecretRef == nil || beta.Spec.TLS.CASecretRef.Name != "ca" {
					t.Fatalf("ca secret ref = %+v", beta.Spec.TLS)
				}
				if beta.Spec.CredentialsSecretRef.Name != "creds" {
					t.Fatalf("credentials = %+v", beta.Spec.CredentialsSecretRef)
				}
				if beta.Status.ObservedGeneration != 1 {
					t.Fatalf("observedGeneration = %d", beta.Status.ObservedGeneration)
				}
			},
			assertBack: func(t *testing.T, orig, back *QueueManagerConnection) {
				t.Helper()
				if back.Spec.TLS == nil || !back.Spec.TLS.InsecureSkipVerify {
					t.Fatalf("tls = %+v", back.Spec.TLS)
				}
				if back.Spec.Endpoint != orig.Spec.Endpoint {
					t.Fatalf("endpoint = %q", back.Spec.Endpoint)
				}
				if back.Annotations[AllowInsecureTLSAnnotation] != "true" {
					t.Fatalf("annotations = %v", back.Annotations)
				}
			},
		},
		{
			name: "nil TLS round-trips as nil",
			alpha: &QueueManagerConnection{
				ObjectMeta: testObjectMeta("qmc-plain"),
				Spec: QueueManagerConnectionSpec{
					QueueManager: "QM1",
					Endpoint:     "https://mq.example:9443",
					CredentialsSecretRef: SecretReference{
						Name: "creds",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.QueueManagerConnection) {
				t.Helper()
				if beta.Spec.TLS != nil {
					t.Fatalf("tls should be nil, got %+v", beta.Spec.TLS)
				}
			},
			assertBack: func(t *testing.T, orig, back *QueueManagerConnection) {
				t.Helper()
				if back.Spec.TLS != nil {
					t.Fatalf("tls should be nil, got %+v", back.Spec.TLS)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			beta, back := roundTripQueueManagerConnection(t, tt.alpha)
			if tt.assertBeta != nil {
				tt.assertBeta(t, beta)
			}
			if tt.assertBack != nil {
				tt.assertBack(t, tt.alpha, back)
			}
		})
	}
}

// TestQueueManagerConnectionLosslessRoundTrip asserts a fully-populated v1alpha1
// QueueManagerConnection survives a hub round-trip byte-for-byte (reflect.DeepEqual).
// Storage-migration guardrail: see TestQueueLosslessRoundTrip. Populates the nested
// TLS + CASecretRef pointers so a dropped nested field is caught.
func TestQueueManagerConnectionLosslessRoundTrip(t *testing.T) {
	t.Parallel()

	orig := &QueueManagerConnection{
		ObjectMeta: testObjectMeta("qmc-lossless"),
		Spec: QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			RESTPrefix:           "/ibmmq/rest/v3",
			CredentialsSecretRef: SecretReference{Name: "creds"},
			TLS: &TLSConfig{
				InsecureSkipVerify: true,
				CASecretRef:        &SecretReference{Name: "ca"},
			},
		},
		Status: QueueManagerConnectionStatus{
			Conditions:         testSyncedCondition(),
			ObservedGeneration: 1,
		},
	}

	_, back := roundTripQueueManagerConnection(t, orig.DeepCopy())
	assertLossless(t, orig, back)
}
