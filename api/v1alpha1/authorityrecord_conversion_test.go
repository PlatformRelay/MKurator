package v1alpha1

import (
	"testing"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestAuthorityRecordConvertToFromRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		alpha      *AuthorityRecord
		assertBeta func(t *testing.T, beta *messagingv1beta1.AuthorityRecord)
		assertBack func(t *testing.T, orig, back *AuthorityRecord)
	}{
		{
			name: "queue profile copies principal and authorities",
			alpha: &AuthorityRecord{
				ObjectMeta: testObjectMeta("auth-queue"),
				Spec: AuthorityRecordSpec{
					ConnectionRef:             LocalObjectReference{Name: "qm1"},
					Profile:                   "APP.ORDERS",
					ObjectType:                AuthorityObjectTypeQueue,
					Principal:                 "app",
					Authorities:               []string{"GET", "PUT"},
					Suspend:                   true,
					WorkloadLifecyclePolicies: testWorkloadPolicies(),
				},
				Status: AuthorityRecordStatus{
					Conditions:           testSyncedCondition(),
					ObservedGeneration:   6,
					DesiredMQSC:          "SET AUTHREC PROFILE(APP.ORDERS)",
					MQObjectStatusFields: testMQObjectStatus(),
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.AuthorityRecord) {
				t.Helper()
				if len(beta.Spec.Authorities) != 2 || beta.Spec.Principal != "app" {
					t.Fatalf("spec = %+v", beta.Spec)
				}
				if beta.Status.ObservedGeneration != 6 {
					t.Fatalf("observedGeneration = %d", beta.Status.ObservedGeneration)
				}
			},
			assertBack: func(t *testing.T, orig, back *AuthorityRecord) {
				t.Helper()
				if len(back.Spec.Authorities) != 2 || back.Spec.Principal != "app" {
					t.Fatalf("spec = %+v", back.Spec)
				}
				if back.Status.DesiredMQSC != orig.Status.DesiredMQSC {
					t.Fatalf("desiredMQSC = %q", back.Status.DesiredMQSC)
				}
			},
		},
		{
			name: "group principal copies object type",
			alpha: &AuthorityRecord{
				ObjectMeta: testObjectMeta("auth-group"),
				Spec: AuthorityRecordSpec{
					ConnectionRef: LocalObjectReference{Name: "qm1"},
					Profile:       "ORDERS.APP",
					ObjectType:    AuthorityObjectTypeChannel,
					Group:         "apps",
					Authorities:   []string{"CONNECT"},
				},
			},
			assertBeta: func(t *testing.T, beta *messagingv1beta1.AuthorityRecord) {
				t.Helper()
				if beta.Spec.Group != "apps" || beta.Spec.ObjectType != messagingv1beta1.AuthorityObjectTypeChannel {
					t.Fatalf("spec = %+v", beta.Spec)
				}
			},
			assertBack: func(t *testing.T, orig, back *AuthorityRecord) {
				t.Helper()
				if back.Spec.Group != "apps" {
					t.Fatalf("group = %q", back.Spec.Group)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			beta, back := roundTripAuthorityRecord(t, tt.alpha)
			if tt.assertBeta != nil {
				tt.assertBeta(t, beta)
			}
			if tt.assertBack != nil {
				tt.assertBack(t, tt.alpha, back)
			}
		})
	}
}

// TestAuthorityRecordLosslessRoundTrip asserts a fully-populated v1alpha1
// AuthorityRecord survives a hub round-trip byte-for-byte (reflect.DeepEqual).
// Storage-migration guardrail: see TestQueueLosslessRoundTrip. Principal and Group
// are mutually exclusive per CRD CEL, so this fixture populates only Principal; the
// Group copy is guarded independently by the sibling
// TestAuthorityRecordConvertToFromRoundTrip "group principal" case, so an empty Group
// here is a valid natural value rather than a masked drop.
func TestAuthorityRecordLosslessRoundTrip(t *testing.T) {
	t.Parallel()

	orig := &AuthorityRecord{
		ObjectMeta: testObjectMeta("auth-lossless"),
		Spec: AuthorityRecordSpec{
			ConnectionRef:             LocalObjectReference{Name: "qm1"},
			Profile:                   "APP.ORDERS",
			ObjectType:                AuthorityObjectTypeQueue,
			Principal:                 "app",
			Authorities:               []string{"GET", "PUT", "CONNECT"},
			Suspend:                   true,
			WorkloadLifecyclePolicies: testWorkloadPolicies(),
		},
		Status: AuthorityRecordStatus{
			Conditions:           testSyncedCondition(),
			ObservedGeneration:   6,
			DesiredMQSC:          "SET AUTHREC PROFILE(APP.ORDERS)",
			MQObjectStatusFields: testMQObjectStatus(),
		},
	}

	_, back := roundTripAuthorityRecord(t, orig.DeepCopy())
	assertLossless(t, orig, back)
}
