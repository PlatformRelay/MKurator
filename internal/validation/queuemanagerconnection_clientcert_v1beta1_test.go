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

func clientCertScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

func clientCertSpec(secretName string) *messagingv1beta1.QueueManagerConnectionSpec {
	return &messagingv1beta1.QueueManagerConnectionSpec{
		QueueManager: "QM1",
		Endpoint:     "https://ibm-mq.svc:9443",
		Authentication: &messagingv1beta1.MQWebAuthentication{
			Mode:       messagingv1beta1.MQWebAuthenticationModeClientCert,
			ClientCert: &messagingv1beta1.ClientCertAuth{SecretRef: messagingv1beta1.SecretReference{Name: secretName}},
		},
	}
}

// AC2: the stateful webhook checks the ClientCert Secret has BOTH tls.crt and tls.key. A
// well-formed tls Secret is admitted with no spurious username warning (unlike Basic/LTPA).
func TestValidateClientCert_TLSSecretShapeOK(t *testing.T) {
	s := clientCertScheme(t)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "mtls-creds", Namespace: "ns"},
		Type:       corev1.SecretTypeTLS,
		Data:       map[string][]byte{"tls.crt": []byte("cert"), "tls.key": []byte("key")},
	}
	cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret).Build()

	warnings, errs := ValidateQueueManagerConnectionSpecV1Beta1(
		context.Background(), cl, "ns", nil, clientCertSpec("mtls-creds"))
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
	for _, w := range warnings {
		if strings.Contains(w, "username") {
			t.Fatalf("ClientCert must not emit a username warning, got: %q", w)
		}
	}
}

// AC2: a missing tls.key (or tls.crt) is a stateful validation error naming the missing key.
func TestValidateClientCert_MissingKeyErrors(t *testing.T) {
	cases := map[string]map[string][]byte{
		"missing tls.key": {"tls.crt": []byte("cert")},
		"missing tls.crt": {"tls.key": []byte("key")},
		"empty tls.key":   {"tls.crt": []byte("cert"), "tls.key": {}},
		"both missing":    {},
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			s := clientCertScheme(t)
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "mtls-creds", Namespace: "ns"},
				Type:       corev1.SecretTypeTLS,
				Data:       data,
			}
			cl := fake.NewClientBuilder().WithScheme(s).WithObjects(secret).Build()
			_, errs := ValidateQueueManagerConnectionSpecV1Beta1(
				context.Background(), cl, "ns", nil, clientCertSpec("mtls-creds"))
			if len(errs) == 0 {
				t.Fatalf("expected a validation error for %s", name)
			}
			joined := errs.ToAggregate().Error()
			if !strings.Contains(joined, "tls.crt") && !strings.Contains(joined, "tls.key") {
				t.Fatalf("error must name the missing key, got: %s", joined)
			}
		})
	}
}

// AC2: a missing ClientCert Secret is a stateful not-found error (config-class).
func TestValidateClientCert_SecretNotFound(t *testing.T) {
	s := clientCertScheme(t)
	cl := fake.NewClientBuilder().WithScheme(s).Build()
	_, errs := ValidateQueueManagerConnectionSpecV1Beta1(
		context.Background(), cl, "ns", nil, clientCertSpec("absent"))
	if len(errs) == 0 {
		t.Fatal("expected a not-found error for a missing ClientCert Secret")
	}
}
