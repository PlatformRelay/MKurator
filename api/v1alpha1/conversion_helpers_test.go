package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

func testObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:              name,
		Namespace:         "default",
		UID:               "test-uid",
		ResourceVersion:   "42",
		Generation:        3,
		CreationTimestamp: metav1.Time{Time: time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)},
		Labels:            map[string]string{"app": "mkurator"},
		Annotations:       map[string]string{"note": "fixture"},
		Finalizers:        []string{"messaging.mkurator.dev/finalizer"},
	}
}

func testSyncedCondition() []metav1.Condition {
	return []metav1.Condition{
		{
			Type:               ConditionSynced,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Time{Time: time.Date(2026, 6, 18, 13, 0, 0, 0, time.UTC)},
			Reason:             ReasonAvailable,
			Message:            "synced",
		},
	}
}

func testMQObjectStatus() MQObjectStatusFields {
	synced := metav1.Now()
	exists := true
	return MQObjectStatusFields{
		Message:        "ok",
		LastSyncTime:   &synced,
		MQObjectExists: &exists,
	}
}

func testWorkloadPolicies() WorkloadLifecyclePolicies {
	return WorkloadLifecyclePolicies{
		DeletionPolicy: DeletionPolicyDelete,
		AdoptionPolicy: AdoptionPolicyAdopt,
	}
}

func int32Ptr(v int32) *int32 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func mapHasKey(m map[string]string, key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func roundTripQueue(t *testing.T, alpha *Queue) (*messagingv1beta1.Queue, *Queue) {
	t.Helper()
	beta := &messagingv1beta1.Queue{}
	if err := alpha.ConvertTo(beta); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	back := &Queue{}
	if err := back.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return beta, back
}

func roundTripBetaQueue(t *testing.T, beta *messagingv1beta1.Queue) (*Queue, *messagingv1beta1.Queue) {
	t.Helper()
	alpha := &Queue{}
	if err := alpha.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	back := &messagingv1beta1.Queue{}
	if err := alpha.ConvertTo(back); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	return alpha, back
}

func roundTripTopic(t *testing.T, alpha *Topic) (*messagingv1beta1.Topic, *Topic) {
	t.Helper()
	beta := &messagingv1beta1.Topic{}
	if err := alpha.ConvertTo(beta); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	back := &Topic{}
	if err := back.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return beta, back
}

func roundTripBetaTopic(t *testing.T, beta *messagingv1beta1.Topic) (*Topic, *messagingv1beta1.Topic) {
	t.Helper()
	alpha := &Topic{}
	if err := alpha.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	back := &messagingv1beta1.Topic{}
	if err := alpha.ConvertTo(back); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	return alpha, back
}

func roundTripChannel(t *testing.T, alpha *Channel) (*messagingv1beta1.Channel, *Channel) {
	t.Helper()
	beta := &messagingv1beta1.Channel{}
	if err := alpha.ConvertTo(beta); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	back := &Channel{}
	if err := back.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return beta, back
}

func roundTripBetaChannel(t *testing.T, beta *messagingv1beta1.Channel) (*Channel, *messagingv1beta1.Channel) {
	t.Helper()
	alpha := &Channel{}
	if err := alpha.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	back := &messagingv1beta1.Channel{}
	if err := alpha.ConvertTo(back); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	return alpha, back
}

func roundTripChannelAuthRule(
	t *testing.T,
	alpha *ChannelAuthRule,
) (*messagingv1beta1.ChannelAuthRule, *ChannelAuthRule) {
	t.Helper()
	beta := &messagingv1beta1.ChannelAuthRule{}
	if err := alpha.ConvertTo(beta); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	back := &ChannelAuthRule{}
	if err := back.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return beta, back
}

func roundTripAuthorityRecord(
	t *testing.T,
	alpha *AuthorityRecord,
) (*messagingv1beta1.AuthorityRecord, *AuthorityRecord) {
	t.Helper()
	beta := &messagingv1beta1.AuthorityRecord{}
	if err := alpha.ConvertTo(beta); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	back := &AuthorityRecord{}
	if err := back.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return beta, back
}

func roundTripQueueManagerConnection(
	t *testing.T,
	alpha *QueueManagerConnection,
) (*messagingv1beta1.QueueManagerConnection, *QueueManagerConnection) {
	t.Helper()
	beta := &messagingv1beta1.QueueManagerConnection{}
	if err := alpha.ConvertTo(beta); err != nil {
		t.Fatalf("ConvertTo: %v", err)
	}
	back := &QueueManagerConnection{}
	if err := back.ConvertFrom(beta); err != nil {
		t.Fatalf("ConvertFrom: %v", err)
	}
	return beta, back
}
