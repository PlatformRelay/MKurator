package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestDeepCopyObjectRoundTrip asserts every registered v1alpha1 type produces an
// independent, deeply-equal copy — the generated deepcopy must not share
// pointers/maps/slices between original and clone.
func TestDeepCopyObjectRoundTrip(t *testing.T) {
	t.Parallel()

	ref := LocalObjectReference{Name: "qm1"}
	status := func() (c []metav1.Condition, f MQObjectStatusFields) {
		return testSyncedCondition(), testMQObjectStatus()
	}
	qc, qf := status()
	tc, tf := status()
	cc, cf := status()
	rc, rf := status()
	ac, af := status()

	objects := []runtime.Object{
		&Queue{
			ObjectMeta: testObjectMeta("q"),
			Spec: QueueSpec{
				ConnectionRef: ref, QueueName: "APP.ORDERS", MaxDepth: int32Ptr(5000),
				Attributes: map[string]string{"custom": "keep"},
			},
			Status: QueueStatus{Conditions: qc, MQObjectStatusFields: qf},
		},
		&QueueList{Items: []Queue{{ObjectMeta: testObjectMeta("q")}}},
		&Topic{
			ObjectMeta: testObjectMeta("t"),
			Spec:       TopicSpec{ConnectionRef: ref, TopicName: "PRICES", Attributes: map[string]string{"a": "b"}},
			Status:     TopicStatus{Conditions: tc, MQObjectStatusFields: tf},
		},
		&TopicList{Items: []Topic{{ObjectMeta: testObjectMeta("t")}}},
		&Channel{
			ObjectMeta: testObjectMeta("c"),
			Spec: ChannelSpec{
				ConnectionRef: ref, ChannelName: "APP.SVRCONN",
				ShareConv: int32Ptr(10), Attributes: map[string]string{"a": "b"},
			},
			Status: ChannelStatus{Conditions: cc, MQObjectStatusFields: cf},
		},
		&ChannelList{Items: []Channel{{ObjectMeta: testObjectMeta("c")}}},
		&ChannelAuthRule{
			ObjectMeta: testObjectMeta("r"),
			Spec:       ChannelAuthRuleSpec{ConnectionRef: ref},
			Status:     ChannelAuthRuleStatus{Conditions: rc, MQObjectStatusFields: rf},
		},
		&ChannelAuthRuleList{Items: []ChannelAuthRule{{ObjectMeta: testObjectMeta("r")}}},
		&AuthorityRecord{
			ObjectMeta: testObjectMeta("a"),
			Spec:       AuthorityRecordSpec{ConnectionRef: ref},
			Status:     AuthorityRecordStatus{Conditions: ac, MQObjectStatusFields: af},
		},
		&AuthorityRecordList{Items: []AuthorityRecord{{ObjectMeta: testObjectMeta("a")}}},
		&QueueManagerConnection{ObjectMeta: testObjectMeta("qmc")},
		&QueueManagerConnectionList{Items: []QueueManagerConnection{{ObjectMeta: testObjectMeta("qmc")}}},
	}

	for _, obj := range objects {
		name := reflect.TypeOf(obj).Elem().Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			clone := obj.DeepCopyObject()
			if clone == obj {
				t.Fatalf("%s: DeepCopyObject returned the same pointer", name)
			}
			if !reflect.DeepEqual(obj, clone) {
				t.Fatalf("%s: clone not deeply equal to original", name)
			}
			// List types embed ListMeta, not ObjectMeta, so they do not carry labels.
			orig, ok := obj.(metav1.Object)
			if !ok {
				return
			}
			orig.GetLabels()["mutated"] = "true"
			if _, leaked := clone.(metav1.Object).GetLabels()["mutated"]; leaked {
				t.Fatalf("%s: mutation of original labels leaked into clone", name)
			}
		})
	}
}
