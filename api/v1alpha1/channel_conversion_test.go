package v1alpha1

import (
	"testing"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestChannelConvertToFromRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alpha      *Channel
		assertBeta func(t *testing.T, beta *messagingv1beta1.Channel)
		assertBack func(t *testing.T, orig, back *Channel)
	}{
		{
			name: "folds trptype attribute to typed transportType",
			alpha: &Channel{
				ObjectMeta: testObjectMeta("c-fold"),
				Spec: ChannelSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					ChannelName:   "ORDERS.APP",
					Attributes: map[string]string{
						"trptype": "tcp",
						"note":    "y",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Channel) {
				t.Helper()
				if beta.Spec.TransportType != messagingv1beta1.ChannelTransportTypeTCP {
					t.Fatalf("transportType = %q", beta.Spec.TransportType)
				}
				if mapHasKey(beta.Spec.Attributes, "trptype") {
					t.Fatalf("promoted key trptype should be removed: %v", beta.Spec.Attributes)
				}
				if beta.Spec.Attributes["note"] != "y" {
					t.Fatalf("attributes = %v", beta.Spec.Attributes)
				}
			},
			assertBack: func(t *testing.T, orig, back *Channel) {
				t.Helper()
				if back.Spec.TransportType != ChannelTransportTypeTCP {
					t.Fatalf("transportType = %q", back.Spec.TransportType)
				}
				if back.Spec.Attributes["note"] != "y" {
					t.Fatalf("attributes = %v", back.Spec.Attributes)
				}
			},
		},
		{
			name: "folds numeric and ssl channel attributes",
			alpha: &Channel{
				ObjectMeta: testObjectMeta("c-full-fold"),
				Spec: ChannelSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					ChannelName:   "ORDERS.APP",
					Type:          ChannelTypeSvrconn,
					Attributes: map[string]string{
						"maxmsgl":  "1048576",
						"sharecnv": "10",
						"maxinst":  "100",
						"maxinstc": "50",
						"mcauser":  "app",
						"sslciph":  "TLS_RSA_WITH_AES_256_CBC_SHA256",
						"sslcauth": "required",
						"conname":  "mq.example(1414)",
						"xmitq":    "SYSTEM.XMIT",
						"descr":    "Orders channel",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Channel) {
				t.Helper()
				if beta.Spec.MaxMsgLength == nil || *beta.Spec.MaxMsgLength != 1048576 {
					t.Fatalf("maxMsgLength = %v", beta.Spec.MaxMsgLength)
				}
				if beta.Spec.ShareConv == nil || *beta.Spec.ShareConv != 10 {
					t.Fatalf("shareConv = %v", beta.Spec.ShareConv)
				}
				if beta.Spec.MaxInstances == nil || *beta.Spec.MaxInstances != 100 {
					t.Fatalf("maxInstances = %v", beta.Spec.MaxInstances)
				}
				if beta.Spec.MaxInstancesClient == nil || *beta.Spec.MaxInstancesClient != 50 {
					t.Fatalf("maxInstancesClient = %v", beta.Spec.MaxInstancesClient)
				}
				if beta.Spec.McaUser != "app" || beta.Spec.SslCipherSpec == "" {
					t.Fatalf("spec = %+v", beta.Spec)
				}
				if beta.Spec.SslClientAuth != messagingv1beta1.ChannelSslClientAuthRequired {
					t.Fatalf("sslClientAuth = %q", beta.Spec.SslClientAuth)
				}
				if beta.Spec.ConnName != "mq.example(1414)" || beta.Spec.XmitQueue != "SYSTEM.XMIT" {
					t.Fatalf("conn/xmit = %q / %q", beta.Spec.ConnName, beta.Spec.XmitQueue)
				}
				if beta.Spec.Description != "Orders channel" {
					t.Fatalf("description = %q", beta.Spec.Description)
				}
				if len(beta.Spec.Attributes) != 0 {
					t.Fatalf("all promoted keys should be folded: %v", beta.Spec.Attributes)
				}
			},
			assertBack: func(t *testing.T, orig, back *Channel) {
				t.Helper()
				if back.Spec.McaUser != "app" || back.Spec.ConnName != "mq.example(1414)" {
					t.Fatalf("spec = %+v", back.Spec)
				}
			},
		},
		{
			name: "typed field wins over conflicting attribute",
			alpha: &Channel{
				ObjectMeta: testObjectMeta("c-prefer-typed"),
				Spec: ChannelSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					ChannelName:   "ORDERS.APP",
					TransportType: ChannelTransportTypeLU62,
					Attributes:    map[string]string{"trptype": "tcp"},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Channel) {
				t.Helper()
				if beta.Spec.TransportType != messagingv1beta1.ChannelTransportTypeLU62 {
					t.Fatalf("transportType = %q", beta.Spec.TransportType)
				}
				if mapHasKey(beta.Spec.Attributes, "trptype") {
					t.Fatalf("conflicting attribute should be dropped")
				}
			},
			assertBack: func(t *testing.T, orig, back *Channel) {
				t.Helper()
				if back.Spec.TransportType != ChannelTransportTypeLU62 {
					t.Fatalf("transportType = %q", back.Spec.TransportType)
				}
			},
		},
		{
			name: "typed-only channel status round-trips",
			alpha: &Channel{
				ObjectMeta: testObjectMeta("c-typed"),
				Spec: ChannelSpec{
					ConnectionRef:             LocalObjectReference{Name: "qm1"},
					ChannelName:               "ORDERS.APP",
					Type:                      ChannelTypeSdr,
					ConnName:                  "mq.example(1414)",
					XmitQueue:                 "SYSTEM.XMIT",
					Suspend:                   true,
					WorkloadLifecyclePolicies: testWorkloadPolicies(),
				},
				Status: ChannelStatus{
					Conditions:           testSyncedCondition(),
					ObservedGeneration:   4,
					DesiredMQSC:          "DEFINE CHANNEL(ORDERS.APP)",
					MQObjectStatusFields: testMQObjectStatus(),
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Channel) {
				t.Helper()
				if !beta.Spec.Suspend || beta.Status.ObservedGeneration != 4 {
					t.Fatalf("spec/status = %+v / %+v", beta.Spec, beta.Status)
				}
			},
			assertBack: func(t *testing.T, orig, back *Channel) {
				t.Helper()
				if back.Status.DesiredMQSC != orig.Status.DesiredMQSC {
					t.Fatalf("desiredMQSC = %q", back.Status.DesiredMQSC)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			beta, back := roundTripChannel(t, tt.alpha)
			if tt.assertBeta != nil {
				tt.assertBeta(t, beta)
			}
			if tt.assertBack != nil {
				tt.assertBack(t, tt.alpha, back)
			}
		})
	}
}

func TestChannelConvertFromHubRoundTrip(t *testing.T) {
	t.Parallel()

	beta := &messagingv1beta1.Channel{
		ObjectMeta: testObjectMeta("c-hub"),
		Spec: messagingv1beta1.ChannelSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.HUB",
			TransportType: messagingv1beta1.ChannelTransportTypeTCP,
			MaxInstances:  int32Ptr(25),
			Attributes:    map[string]string{"custom": "z"},
		},
	}

	_, back := roundTripBetaChannel(t, beta)
	if back.Spec.MaxInstances == nil || *back.Spec.MaxInstances != 25 {
		t.Fatalf("maxInstances = %v", back.Spec.MaxInstances)
	}
	if back.Spec.Attributes["custom"] != "z" {
		t.Fatalf("attributes = %v", back.Spec.Attributes)
	}
}
