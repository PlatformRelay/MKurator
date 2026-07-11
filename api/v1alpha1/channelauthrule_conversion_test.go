package v1alpha1

import (
	"testing"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestChannelAuthRuleConvertToFromRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alpha      *ChannelAuthRule
		assertBeta func(t *testing.T, beta *messagingv1beta1.ChannelAuthRule)
		assertBack func(t *testing.T, orig, back *ChannelAuthRule)
	}{
		{
			name: "ADDRESSMAP rule copies typed auth fields",
			alpha: &ChannelAuthRule{
				ObjectMeta: testObjectMeta("car-addressmap"),
				Spec: ChannelAuthRuleSpec{
					ConnectionRef:             LocalObjectReference{Name: "qm1"},
					ChannelName:               "ORDERS.APP",
					RuleType:                  ChannelAuthRuleTypeAddressMap,
					Address:                   "*",
					UserSource:                ChannelAuthUserSourceMap,
					CheckClient:               ChannelAuthCheckClientRequired,
					McaUser:                   "app",
					Description:               "Allow app users",
					Suspend:                   true,
					WorkloadLifecyclePolicies: testWorkloadPolicies(),
				},
				Status: ChannelAuthRuleStatus{
					Conditions:           testSyncedCondition(),
					ObservedGeneration:   3,
					DesiredMQSC:          "SET CHLAUTH(ORDERS.APP) TYPE(ADDRESSMAP)",
					MQObjectStatusFields: testMQObjectStatus(),
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.ChannelAuthRule) {
				t.Helper()
				if beta.Spec.McaUser != "app" || beta.Spec.RuleType != messagingv1beta1.ChannelAuthRuleTypeAddressMap {
					t.Fatalf("spec = %+v", beta.Spec)
				}
				if beta.Status.DesiredMQSC == "" || beta.Status.ObservedGeneration != 3 {
					t.Fatalf("status = %+v", beta.Status)
				}
			},
			assertBack: func(t *testing.T, orig, back *ChannelAuthRule) {
				t.Helper()
				if back.Spec.McaUser != orig.Spec.McaUser || back.Spec.Address != orig.Spec.Address {
					t.Fatalf("spec = %+v", back.Spec)
				}
				if back.Status.DesiredMQSC != orig.Status.DesiredMQSC {
					t.Fatalf("desiredMQSC = %q", back.Status.DesiredMQSC)
				}
			},
		},
		{
			name: "USERMAP rule copies clientUser and userSource",
			alpha: &ChannelAuthRule{
				ObjectMeta: testObjectMeta("car-usermap"),
				Spec: ChannelAuthRuleSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					ChannelName:   "ORDERS.APP",
					RuleType:      ChannelAuthRuleTypeUserMap,
					ClientUser:    "alice",
					UserSource:    ChannelAuthUserSourceMap,
					McaUser:       "mqalice",
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.ChannelAuthRule) {
				t.Helper()
				if beta.Spec.ClientUser != "alice" || beta.Spec.McaUser != "mqalice" {
					t.Fatalf("spec = %+v", beta.Spec)
				}
			},
			assertBack: func(t *testing.T, orig, back *ChannelAuthRule) {
				t.Helper()
				if back.Spec.ClientUser != "alice" {
					t.Fatalf("clientUser = %q", back.Spec.ClientUser)
				}
			},
		},
		{
			name: "SSLPEERMAP and QMGRMAP copy match fields",
			alpha: &ChannelAuthRule{
				ObjectMeta: testObjectMeta("car-ssl-qmgr"),
				Spec: ChannelAuthRuleSpec{
					ConnectionRef:      LocalObjectReference{Name: "qm1"},
					ChannelName:        "ORDERS.APP",
					RuleType:           ChannelAuthRuleTypeSSLPeerMap,
					SslPeerName:        "CN=app",
					UserSource:         ChannelAuthUserSourceMap,
					McaUser:            "app",
					RemoteQueueManager: "QM2",
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.ChannelAuthRule) {
				t.Helper()
				if beta.Spec.SslPeerName != "CN=app" {
					t.Fatalf("sslPeerName = %q", beta.Spec.SslPeerName)
				}
			},
			assertBack: func(t *testing.T, orig, back *ChannelAuthRule) {
				t.Helper()
				if back.Spec.SslPeerName != "CN=app" {
					t.Fatalf("sslPeerName = %q", back.Spec.SslPeerName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			beta, back := roundTripChannelAuthRule(t, tt.alpha)
			if tt.assertBeta != nil {
				tt.assertBeta(t, beta)
			}
			if tt.assertBack != nil {
				tt.assertBack(t, tt.alpha, back)
			}
		})
	}
}
