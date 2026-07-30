package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// MQObject is implemented by workload CRs reconciled against IBM MQ. Method sets live on the
// api workload types; the interface lives here so controller-gen does not deepcopy it. The
// status-fields accessor returns the v1beta1 hub struct, satisfied by every v1beta1 workload.
type MQObject interface {
	GetMQConditions() *[]metav1.Condition
	GetMQStatusFields() *messagingv1beta1.MQObjectStatusFields
	GetStatusObservedGeneration() *int64
	SetStatusObservedGeneration(int64)
	ConnectionRefName() string
}
