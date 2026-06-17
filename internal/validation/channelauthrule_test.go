package validation

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/conduit-ops/mkurator/api/v1alpha1"
)

func TestValidateChannelAuthRuleSpecMissingManagedChannel(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car1",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeNotFound && err.Field == "spec.channelName" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.channelName not found, got %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecValid(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "dev-app-addressmap",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
			UserSource:    "CHANNEL",
		})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecNewRuleTypesTable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		spec messagingv1alpha1.ChannelAuthRuleSpec
	}{
		{
			name: "usermap map",
			spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:   "ORDERS.APP",
				RuleType:      messagingv1alpha1.ChannelAuthRuleTypeUserMap,
				ClientUser:    "johndoe",
				UserSource:    messagingv1alpha1.ChannelAuthUserSourceMap,
				McaUser:       "orders-app",
			},
		},
		{
			name: "usermap channel userSource",
			spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:   "ORDERS.APP",
				RuleType:      messagingv1alpha1.ChannelAuthRuleTypeUserMap,
				ClientUser:    "johndoe",
				UserSource:    messagingv1alpha1.ChannelAuthUserSourceChannel,
			},
		},
		{
			name: "sslpeermap map",
			spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:   "ORDERS.APP",
				RuleType:      messagingv1alpha1.ChannelAuthRuleTypeSSLPeerMap,
				SslPeerName:   "CN=AppClient,O=MyOrg,C=US",
				UserSource:    messagingv1alpha1.ChannelAuthUserSourceMap,
				McaUser:       "orders-app",
			},
		},
		{
			name: "sslpeermap channel userSource",
			spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:   "ORDERS.APP",
				RuleType:      messagingv1alpha1.ChannelAuthRuleTypeSSLPeerMap,
				SslPeerName:   "CN=AppClient,O=MyOrg,C=US",
				UserSource:    messagingv1alpha1.ChannelAuthUserSourceChannel,
			},
		},
		{
			name: "qmgrmap map",
			spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef:      messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:        "ORDERS.APP",
				RuleType:           messagingv1alpha1.ChannelAuthRuleTypeQMGRMap,
				RemoteQueueManager: "QM_PARTNER",
				UserSource:         messagingv1alpha1.ChannelAuthUserSourceMap,
				McaUser:            "orders-app",
			},
		},
		{
			name: "qmgrmap channel userSource",
			spec: messagingv1alpha1.ChannelAuthRuleSpec{
				ConnectionRef:      messagingv1alpha1.LocalObjectReference{Name: "qm1"},
				ChannelName:        "ORDERS.APP",
				RuleType:           messagingv1alpha1.ChannelAuthRuleTypeQMGRMap,
				RemoteQueueManager: "QM_PARTNER",
				UserSource:         messagingv1alpha1.ChannelAuthUserSourceChannel,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scheme := runtime.NewScheme()
			_ = messagingv1alpha1.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			conn := &messagingv1alpha1.QueueManagerConnection{
				ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
				Spec: messagingv1alpha1.QueueManagerConnectionSpec{
					QueueManager:         "QM1",
					Endpoint:             "https://mq.example:9443",
					CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
				},
			}
			secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
			ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
			cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

			errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car-new-type",
				&tt.spec)
			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
		})
	}
}

func TestValidateChannelAuthRuleSpecBlockAddrWildcardSkipsChannelRef(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "blockaddr",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "*",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockAddr,
			Address:       "192.0.2.1",
		})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}
