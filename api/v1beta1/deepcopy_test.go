package v1beta1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// deepCopyable is implemented by the generated zz_generated.deepcopy.go for every
// registered API object and object list.
type deepCopyable interface {
	runtime.Object
}

func testStatusFields() MQObjectStatusFields {
	now := metav1.Now()
	exists := true
	return MQObjectStatusFields{
		Message:        "ok",
		LastSyncTime:   &now,
		MQObjectExists: &exists,
	}
}

func testConditions() []metav1.Condition {
	return []metav1.Condition{{
		Type:               ConditionSynced,
		Status:             metav1.ConditionTrue,
		Reason:             ReasonAvailable,
		Message:            "synced",
		LastTransitionTime: metav1.Now(),
	}}
}

func testMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        "obj",
		Namespace:   "default",
		Labels:      map[string]string{"app": "mkurator"},
		Annotations: map[string]string{"note": "fixture"},
		Finalizers:  []string{"messaging.mkurator.dev/finalizer"},
	}
}

// TestDeepCopyObjectRoundTrip asserts every registered type produces an independent,
// deeply-equal copy — the generated deepcopy must not share pointers/maps/slices.
func TestDeepCopyObjectRoundTrip(t *testing.T) {
	t.Parallel()

	ref := LocalObjectReference{Name: "qm1"}
	objects := []deepCopyable{
		&Queue{
			ObjectMeta: testMeta(),
			Spec: QueueSpec{
				ConnectionRef: ref, QueueName: "APP.ORDERS", MaxDepth: int32Ptr(5000),
				Attributes: map[string]string{"custom": "keep"},
			},
			Status: QueueStatus{Conditions: testConditions(), MQObjectStatusFields: testStatusFields()},
		},
		&QueueList{Items: []Queue{{ObjectMeta: testMeta()}}},
		&Topic{
			ObjectMeta: testMeta(),
			Spec:       TopicSpec{ConnectionRef: ref, TopicName: "PRICES", Attributes: map[string]string{"a": "b"}},
			Status:     TopicStatus{Conditions: testConditions(), MQObjectStatusFields: testStatusFields()},
		},
		&TopicList{Items: []Topic{{ObjectMeta: testMeta()}}},
		&Channel{
			ObjectMeta: testMeta(),
			Spec: ChannelSpec{
				ConnectionRef: ref, ChannelName: "APP.SVRCONN",
				ShareConv: int32Ptr(10), Attributes: map[string]string{"a": "b"},
			},
			Status: ChannelStatus{Conditions: testConditions(), MQObjectStatusFields: testStatusFields()},
		},
		&ChannelList{Items: []Channel{{ObjectMeta: testMeta()}}},
		&ChannelAuthRule{
			ObjectMeta: testMeta(),
			Spec:       ChannelAuthRuleSpec{ConnectionRef: ref},
			Status:     ChannelAuthRuleStatus{Conditions: testConditions(), MQObjectStatusFields: testStatusFields()},
		},
		&ChannelAuthRuleList{Items: []ChannelAuthRule{{ObjectMeta: testMeta()}}},
		&AuthorityRecord{
			ObjectMeta: testMeta(),
			Spec:       AuthorityRecordSpec{ConnectionRef: ref},
			Status:     AuthorityRecordStatus{Conditions: testConditions(), MQObjectStatusFields: testStatusFields()},
		},
		&AuthorityRecordList{Items: []AuthorityRecord{{ObjectMeta: testMeta()}}},
		&QueueManagerConnection{ObjectMeta: testMeta()},
		&QueueManagerConnectionList{Items: []QueueManagerConnection{{ObjectMeta: testMeta()}}},
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
			// Mutating the original's labels must not leak into the clone (proves map copy).
			// List types embed ListMeta, not ObjectMeta, so they do not carry labels.
			orig, ok := obj.(metav1.Object)
			if !ok {
				return
			}
			orig.GetLabels()["mutated"] = "true"
			cloneLabels := clone.(metav1.Object).GetLabels()
			if _, leaked := cloneLabels["mutated"]; leaked {
				t.Fatalf("%s: mutation of original labels leaked into clone", name)
			}
		})
	}
}
