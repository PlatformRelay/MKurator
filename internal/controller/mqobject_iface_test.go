package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestMQObject_Interface(t *testing.T) {
	t.Parallel()
	ref := func(name string) messagingv1beta1.LocalObjectReference {
		return messagingv1beta1.LocalObjectReference{Name: name}
	}
	objects := []MQObject{
		&messagingv1beta1.Queue{Spec: messagingv1beta1.QueueSpec{ConnectionRef: ref("qm1")}},
		&messagingv1beta1.Topic{Spec: messagingv1beta1.TopicSpec{ConnectionRef: ref("qm2")}},
		&messagingv1beta1.Channel{Spec: messagingv1beta1.ChannelSpec{ConnectionRef: ref("qm3")}},
		&messagingv1beta1.ChannelAuthRule{Spec: messagingv1beta1.ChannelAuthRuleSpec{ConnectionRef: ref("qm4")}},
		&messagingv1beta1.AuthorityRecord{Spec: messagingv1beta1.AuthorityRecordSpec{ConnectionRef: ref("qm5")}},
	}
	wantRefs := []string{"qm1", "qm2", "qm3", "qm4", "qm5"}

	for i, obj := range objects {
		if obj.ConnectionRefName() != wantRefs[i] {
			t.Fatalf("ConnectionRefName() = %q, want %q", obj.ConnectionRefName(), wantRefs[i])
		}
		conds := obj.GetMQConditions()
		if conds == nil {
			t.Fatal("GetMQConditions() nil")
		}
		*conds = []metav1.Condition{{Type: messagingv1beta1.ConditionSynced, Status: metav1.ConditionTrue}}
		if len(*obj.GetMQConditions()) != 1 {
			t.Fatal("conditions not updated via pointer")
		}
		fields := obj.GetMQStatusFields()
		if fields == nil {
			t.Fatal("GetMQStatusFields() nil")
		}
		fields.Message = "ok"
		if obj.GetMQStatusFields().Message != "ok" {
			t.Fatal("status fields not updated via pointer")
		}
		obj.SetStatusObservedGeneration(7)
		if *obj.GetStatusObservedGeneration() != 7 {
			t.Fatalf("ObservedGeneration = %d, want 7", *obj.GetStatusObservedGeneration())
		}
	}
}
