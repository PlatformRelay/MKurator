package v1beta1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mqObject is the subset of the workload-CR method set exercised by the accessor tests.
// It mirrors the interface the controllers depend on (defined in internal/controller).
type mqObject interface {
	GetMQConditions() *[]metav1.Condition
	GetMQStatusFields() *MQObjectStatusFields
	GetStatusObservedGeneration() *int64
	SetStatusObservedGeneration(int64)
	ConnectionRefName() string
}

func TestMQObjectAccessors(t *testing.T) {
	t.Parallel()

	ref := LocalObjectReference{Name: "qm1"}
	cases := []struct {
		name string
		obj  mqObject
	}{
		{"Queue", &Queue{Spec: QueueSpec{ConnectionRef: ref}}},
		{"Topic", &Topic{Spec: TopicSpec{ConnectionRef: ref}}},
		{"Channel", &Channel{Spec: ChannelSpec{ConnectionRef: ref}}},
		{"ChannelAuthRule", &ChannelAuthRule{Spec: ChannelAuthRuleSpec{ConnectionRef: ref}}},
		{"AuthorityRecord", &AuthorityRecord{Spec: AuthorityRecordSpec{ConnectionRef: ref}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.obj.ConnectionRefName(); got != "qm1" {
				t.Fatalf("ConnectionRefName() = %q, want qm1", got)
			}

			// ObservedGeneration is a read/write accessor pair sharing the same field.
			tc.obj.SetStatusObservedGeneration(9)
			if got := tc.obj.GetStatusObservedGeneration(); got == nil || *got != 9 {
				t.Fatalf("GetStatusObservedGeneration() = %v, want 9", got)
			}

			// Conditions accessor returns a mutable pointer into the status.
			conds := tc.obj.GetMQConditions()
			if conds == nil {
				t.Fatalf("GetMQConditions() returned nil")
			}
			*conds = append(*conds, metav1.Condition{Type: ConditionSynced, Status: metav1.ConditionTrue})
			if got := tc.obj.GetMQConditions(); len(*got) != 1 {
				t.Fatalf("conditions not persisted through pointer: %v", *got)
			}

			// Status fields accessor returns a mutable pointer into the status.
			fields := tc.obj.GetMQStatusFields()
			if fields == nil {
				t.Fatalf("GetMQStatusFields() returned nil")
			}
			fields.Message = "reconciled"
			if got := tc.obj.GetMQStatusFields(); got.Message != "reconciled" {
				t.Fatalf("status fields not persisted through pointer: %q", got.Message)
			}
		})
	}
}

// hubMarker is satisfied by every v1beta1 hub type via its Hub() method.
type hubMarker interface{ Hub() }

func TestHubMarkers(t *testing.T) {
	t.Parallel()

	markers := []hubMarker{
		&Queue{},
		&Topic{},
		&Channel{},
		&ChannelAuthRule{},
		&AuthorityRecord{},
		&QueueManagerConnection{},
	}
	for _, m := range markers {
		// Hub() is a no-op marker; calling it proves the type satisfies conversion.Hub.
		m.Hub()
	}
}
