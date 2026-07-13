package v1alpha1

import (
	"testing"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestQueueConvertToFromRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alpha      *Queue
		assertBeta func(t *testing.T, beta *messagingv1beta1.Queue)
		assertBack func(t *testing.T, orig, back *Queue)
	}{
		{
			name: "minimal spec copies metadata and connection ref",
			alpha: &Queue{
				ObjectMeta: testObjectMeta("q-min"),
				Spec: QueueSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					QueueName:     "APP.ORDERS",
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Queue) {
				t.Helper()
				if beta.Spec.QueueName != "APP.ORDERS" || beta.Spec.ConnectionRef.Name != "qm1" {
					t.Fatalf("spec = %+v", beta.Spec)
				}
			},
			assertBack: func(t *testing.T, orig, back *Queue) {
				t.Helper()
				if back.Spec.QueueName != orig.Spec.QueueName {
					t.Fatalf("queueName = %q", back.Spec.QueueName)
				}
			},
		},
		{
			name: "folds maxdepth attribute to typed field on hub",
			alpha: &Queue{
				ObjectMeta: testObjectMeta("q-fold"),
				Spec: QueueSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					QueueName:     "APP.ORDERS",
					Attributes: map[string]string{
						"maxdepth": "5000",
						"custom":   "keep-me",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Queue) {
				t.Helper()
				if beta.Spec.MaxDepth == nil || *beta.Spec.MaxDepth != 5000 {
					t.Fatalf("maxDepth = %v", beta.Spec.MaxDepth)
				}
				if mapHasKey(beta.Spec.Attributes, "maxdepth") {
					t.Fatalf("promoted key maxdepth should be removed from attributes: %v", beta.Spec.Attributes)
				}
				if beta.Spec.Attributes["custom"] != "keep-me" {
					t.Fatalf("attributes = %v", beta.Spec.Attributes)
				}
			},
			assertBack: func(t *testing.T, orig, back *Queue) {
				t.Helper()
				if back.Spec.MaxDepth == nil || *back.Spec.MaxDepth != 5000 {
					t.Fatalf("maxDepth = %v", back.Spec.MaxDepth)
				}
				if back.Spec.Attributes["custom"] != "keep-me" {
					t.Fatalf("attributes = %v", back.Spec.Attributes)
				}
				if mapHasKey(back.Spec.Attributes, "maxdepth") {
					t.Fatalf("folded key should not reappear in attributes after round-trip")
				}
			},
		},
		{
			name: "case-insensitive attribute key folding",
			alpha: &Queue{
				ObjectMeta: testObjectMeta("q-case"),
				Spec: QueueSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					QueueName:     "APP.ORDERS",
					Attributes:    map[string]string{"MAXDEPTH": "42"},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Queue) {
				t.Helper()
				if beta.Spec.MaxDepth == nil || *beta.Spec.MaxDepth != 42 {
					t.Fatalf("maxDepth = %v", beta.Spec.MaxDepth)
				}
			},
			assertBack: func(t *testing.T, orig, back *Queue) {
				t.Helper()
				if back.Spec.MaxDepth == nil || *back.Spec.MaxDepth != 42 {
					t.Fatalf("maxDepth = %v", back.Spec.MaxDepth)
				}
			},
		},
		{
			name: "folds alternate remote queue attribute keys",
			alpha: &Queue{
				ObjectMeta: testObjectMeta("q-remote"),
				Spec: QueueSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					QueueName:     "APP.REMOTE",
					Type:          QueueTypeRemote,
					Attributes: map[string]string{
						"transmissionqueue": "SYSTEM.XMIT",
						"remotemanager":     "QM2",
					},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Queue) {
				t.Helper()
				if beta.Spec.XmitQueue != "SYSTEM.XMIT" || beta.Spec.RemoteQueueManager != "QM2" {
					t.Fatalf("spec = %+v", beta.Spec)
				}
				if len(beta.Spec.Attributes) != 0 {
					t.Fatalf("attributes = %v", beta.Spec.Attributes)
				}
			},
			assertBack: func(t *testing.T, orig, back *Queue) {
				t.Helper()
				if back.Spec.XmitQueue != "SYSTEM.XMIT" || back.Spec.RemoteQueueManager != "QM2" {
					t.Fatalf("spec = %+v", back.Spec)
				}
			},
		},
		{
			name: "typed field wins over conflicting attribute on hub",
			alpha: &Queue{
				ObjectMeta: testObjectMeta("q-prefer-typed"),
				Spec: QueueSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					QueueName:     "APP.ORDERS",
					MaxDepth:      int32Ptr(100),
					Attributes:    map[string]string{"maxdepth": "9999"},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Queue) {
				t.Helper()
				if beta.Spec.MaxDepth == nil || *beta.Spec.MaxDepth != 100 {
					t.Fatalf("maxDepth = %v", beta.Spec.MaxDepth)
				}
				if mapHasKey(beta.Spec.Attributes, "maxdepth") {
					t.Fatalf("conflicting attribute should be dropped when typed field is set")
				}
			},
			assertBack: func(t *testing.T, orig, back *Queue) {
				t.Helper()
				if back.Spec.MaxDepth == nil || *back.Spec.MaxDepth != 100 {
					t.Fatalf("maxDepth = %v", back.Spec.MaxDepth)
				}
			},
		},
		{
			name: "typed-only spec round-trips through hub unchanged",
			alpha: &Queue{
				ObjectMeta: testObjectMeta("q-typed"),
				Spec: QueueSpec{
					ConnectionRef:             LocalObjectReference{Name: "qm1"},
					QueueName:                 "APP.ORDERS",
					Type:                      QueueTypeLocal,
					MaxDepth:                  int32Ptr(2000),
					Description:               "orders queue",
					DefPersistence:            QueueDefaultPersistenceYes,
					Get:                       QueueAccessEnabledEnabled,
					Put:                       QueueAccessEnabledDisabled,
					Suspend:                   true,
					WorkloadLifecyclePolicies: testWorkloadPolicies(),
				},
				Status: QueueStatus{
					Conditions:           testSyncedCondition(),
					ObservedGeneration:   7,
					DesiredMQSC:          "DEFINE QLOCAL(APP.ORDERS)",
					MQObjectStatusFields: testMQObjectStatus(),
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.Queue) {
				t.Helper()
				if beta.Spec.Description != "orders queue" || !beta.Spec.Suspend {
					t.Fatalf("spec = %+v", beta.Spec)
				}
				if beta.Status.ObservedGeneration != 7 || beta.Status.DesiredMQSC == "" {
					t.Fatalf("status = %+v", beta.Status)
				}
			},
			assertBack: func(t *testing.T, orig, back *Queue) {
				t.Helper()
				if back.Spec.Description != orig.Spec.Description || back.Spec.Suspend != orig.Spec.Suspend {
					t.Fatalf("spec = %+v", back.Spec)
				}
				if back.Status.ObservedGeneration != orig.Status.ObservedGeneration {
					t.Fatalf("observedGeneration = %d", back.Status.ObservedGeneration)
				}
				if back.Status.DesiredMQSC != orig.Status.DesiredMQSC {
					t.Fatalf("desiredMQSC = %q", back.Status.DesiredMQSC)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			beta, back := roundTripQueue(t, tt.alpha)
			if tt.assertBeta != nil {
				tt.assertBeta(t, beta)
			}
			if tt.assertBack != nil {
				tt.assertBack(t, tt.alpha, back)
			}
		})
	}
}

// TestQueueLosslessRoundTrip asserts that a fully-populated v1alpha1 Queue survives
// a hub round-trip byte-for-byte (reflect.DeepEqual). v1alpha1 is the storage
// version, so every field crosses the hub on each read/write; this guards against a
// converter silently dropping a field during storage migration.
func TestQueueLosslessRoundTrip(t *testing.T) {
	t.Parallel()

	orig := &Queue{
		ObjectMeta: testObjectMeta("q-lossless"),
		Spec: QueueSpec{
			ConnectionRef: LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
			Type:          QueueTypeRemote,
			// Attributes intentionally uses only a non-foldable ("custom") key:
			// foldable MQSC keys are by-design normalized into typed fields on the hub
			// (mutually exclusive with them per CRD CEL), so a foldable key would not
			// round-trip as an attribute. That fold is the one intentionally-lossy path.
			Attributes:                map[string]string{"custom": "keep-me"},
			MaxDepth:                  int32Ptr(5000),
			Description:               "orders queue",
			DefPersistence:            QueueDefaultPersistenceYes,
			Get:                       QueueAccessEnabledEnabled,
			Put:                       QueueAccessEnabledDisabled,
			TargetQueue:               "TARGET.Q",
			XmitQueue:                 "SYSTEM.XMIT",
			RemoteQueueManager:        "QM2",
			Suspend:                   true,
			WorkloadLifecyclePolicies: testWorkloadPolicies(),
		},
		Status: QueueStatus{
			Conditions:           testSyncedCondition(),
			ObservedGeneration:   7,
			DesiredMQSC:          "DEFINE QLOCAL(APP.ORDERS)",
			MQObjectStatusFields: testMQObjectStatus(),
		},
	}

	_, back := roundTripQueue(t, orig.DeepCopy())
	assertLossless(t, orig, back)
}

func TestQueueConvertFromHubRoundTrip(t *testing.T) {
	t.Parallel()

	beta := &messagingv1beta1.Queue{
		ObjectMeta: testObjectMeta("q-hub"),
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.HUB",
			MaxDepth:      int32Ptr(3000),
			Attributes:    map[string]string{"custom": "only"},
		},
		Status: messagingv1beta1.QueueStatus{
			ObservedGeneration: 5,
			DesiredMQSC:        "DEFINE QLOCAL(APP.HUB)",
		},
	}

	_, back := roundTripBetaQueue(t, beta)
	if back.Spec.MaxDepth == nil || *back.Spec.MaxDepth != 3000 {
		t.Fatalf("maxDepth = %v", back.Spec.MaxDepth)
	}
	if back.Spec.Attributes["custom"] != "only" {
		t.Fatalf("attributes = %v", back.Spec.Attributes)
	}
	if back.Status.ObservedGeneration != 5 {
		t.Fatalf("observedGeneration = %d", back.Status.ObservedGeneration)
	}
}
