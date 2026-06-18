package v1beta1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MQObjectStatusFields are shared status fields for Queue, Topic, Channel, and auth CRs.
type MQObjectStatusFields struct {
	// Message is a short, user-facing summary of reconcile state (especially when Synced=False).
	// +optional
	Message string `json:"message,omitempty"`

	// LastSyncTime is set when the object last reconciled successfully.
	// +optional
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// MQObjectExists is true when the IBM MQ object was last observed on the queue manager.
	// +optional
	MQObjectExists *bool `json:"mqObjectExists,omitempty"`
}
