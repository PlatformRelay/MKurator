package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// AUTH-12 AC4: a v1beta1 hub QMC that uses the authentication union, when down-converted to
// the v1alpha1 spoke, drops `authentication` without panic or corruption. v1beta1 is the
// storage hub (post-8d-7), so nothing is lost at rest — the spoke is a lossy read view.
func TestQueueManagerConnection_DownConvertDropsAuthenticationUnion(t *testing.T) {
	cases := []struct {
		name string
		hub  *messagingv1beta1.QueueManagerConnection
		// wantCredName is the CredentialsSecretRef.Name expected on the spoke after
		// down-conversion (empty when the hub carried only the union).
		wantCredName string
	}{
		{
			name: "union basic only, legacy ref nil",
			hub: &messagingv1beta1.QueueManagerConnection{
				ObjectMeta: metav1.ObjectMeta{Name: "qm-union", Namespace: "ns"},
				Spec: messagingv1beta1.QueueManagerConnectionSpec{
					QueueManager: "QM1",
					Endpoint:     "https://mq.svc:9443",
					Authentication: &messagingv1beta1.MQWebAuthentication{
						Mode: messagingv1beta1.MQWebAuthenticationModeBasic,
						Basic: &messagingv1beta1.BasicAuth{
							SecretRef: messagingv1beta1.SecretReference{Name: "union-creds"},
						},
					},
				},
			},
			wantCredName: "",
		},
		{
			name: "legacy ref present, no union",
			hub: &messagingv1beta1.QueueManagerConnection{
				ObjectMeta: metav1.ObjectMeta{Name: "qm-legacy", Namespace: "ns"},
				Spec: messagingv1beta1.QueueManagerConnectionSpec{
					QueueManager:         "QM1",
					Endpoint:             "https://mq.svc:9443",
					CredentialsSecretRef: &messagingv1beta1.SecretReference{Name: "legacy-creds"},
				},
			},
			wantCredName: "legacy-creds",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spoke := &QueueManagerConnection{}
			// Must not panic and must not error.
			if err := spoke.ConvertFrom(tc.hub); err != nil {
				t.Fatalf("ConvertFrom: %v", err)
			}
			if spoke.Spec.CredentialsSecretRef.Name != tc.wantCredName {
				t.Fatalf("credentialsSecretRef.Name = %q, want %q",
					spoke.Spec.CredentialsSecretRef.Name, tc.wantCredName)
			}
			// The spoke type has no authentication field at all; presence of the union on
			// the hub must not have corrupted the other copied fields.
			if spoke.Spec.QueueManager != tc.hub.Spec.QueueManager ||
				spoke.Spec.Endpoint != tc.hub.Spec.Endpoint {
				t.Fatalf("core fields corrupted during down-conversion: %+v", spoke.Spec)
			}
		})
	}
}

// FuzzQueueManagerConnectionUnionDownConvert fuzzes down-conversion of a hub carrying the
// authentication union: it must never panic regardless of which members/mode are populated
// (the union is dropped; only credentialsSecretRef survives to the spoke).
func FuzzQueueManagerConnectionUnionDownConvert(f *testing.F) {
	f.Add("qm", "ns", "QM1", "https://mq.svc:9443", "Basic", true, false, false, "bsec", "lsec", "csec", true)
	f.Add("qm", "ns", "QM1", "https://mq.svc:9443", "LTPA", false, true, false, "bsec", "lsec", "csec", false)
	f.Add("qm", "ns", "QM1", "https://mq.svc:9443", "ClientCert", false, false, true, "bsec", "lsec", "csec", true)
	f.Add("qm", "ns", "QM1", "https://mq.svc:9443", "", false, false, false, "", "", "", false)

	f.Fuzz(func(t *testing.T,
		name, namespace, qm, endpoint, mode string,
		setBasic, setLTPA, setClientCert bool,
		basicRef, ltpaRef, certRef string,
		setLegacyRef bool,
	) {
		hub := &messagingv1beta1.QueueManagerConnection{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
			Spec: messagingv1beta1.QueueManagerConnectionSpec{
				QueueManager: qm,
				Endpoint:     endpoint,
				Authentication: &messagingv1beta1.MQWebAuthentication{
					Mode: messagingv1beta1.MQWebAuthenticationMode(mode),
				},
			},
		}
		if setBasic {
			hub.Spec.Authentication.Basic = &messagingv1beta1.BasicAuth{
				SecretRef: messagingv1beta1.SecretReference{Name: basicRef},
			}
		}
		if setLTPA {
			hub.Spec.Authentication.LTPA = &messagingv1beta1.LTPAAuth{
				SecretRef: messagingv1beta1.SecretReference{Name: ltpaRef},
			}
		}
		if setClientCert {
			hub.Spec.Authentication.ClientCert = &messagingv1beta1.ClientCertAuth{
				SecretRef: messagingv1beta1.SecretReference{Name: certRef},
			}
		}
		if setLegacyRef {
			hub.Spec.CredentialsSecretRef = &messagingv1beta1.SecretReference{Name: "legacy"}
		}

		spoke := &QueueManagerConnection{}
		// The only assertion is: no panic, no error. The union is intentionally dropped.
		if err := spoke.ConvertFrom(hub); err != nil {
			t.Fatalf("ConvertFrom panicked/errored on union hub: %v", err)
		}
	})
}
